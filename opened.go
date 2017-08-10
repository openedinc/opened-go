// Package opened provides structures for OpenEd objects
//such as resources and standards.
package opened

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/golang/glog"
	"github.com/jmcvetta/napping"
	"github.com/jmoiron/sqlx"
	goredis "gopkg.in/redis.v3"
)

// A Resource has information such as Publisher, Title, Description for video, game or assessment
type Resource struct {
	ID             int
	Title          sql.NullString
	URL            sql.NullString `db:"share_url"`
	PublisherID    sql.NullInt64 `db:"publisher_id"`
	ContributionID sql.NullInt64 `db:"contribution_id"`
	Description    sql.NullString
	ResourceTypeID sql.NullInt64  `db:"resource_type_id"`
	YoutubeID      sql.NullString `db:"youtube_id"`
	UsageCount     sql.NullInt64  `db:"usage_count"`
	Effectiveness  string         `json:"effectiveness"`
	Subject        string         `json:"subject"`
}

// WsResource is web service queryParams for OpenEd resources (not all attributes in OpenEd).
type WsResource struct {
	ID             int
	Title          string
	URL            string
	PublisherID    int `json:"publisher_id"`
	ContributionID int `json:"contribution_id"`
	Description    string
	ResourceTypeID int    `json:"resource_type_id"`
	YoutubeID      string `json:"youtube_id"`
	UseRightsURL   string `json:"use_rights_url"`
}

// ResourceList is a list of WSResources.
type ResourceList struct {
	Resources []WsResource
}

// SearchResources searches OpenEd for resources given set of queryParams.
func SearchResources(queryParams map[string]string, token string) (ResourceList, error) {
	var err error
	uri := os.Getenv("PARTNER_BASE_URI") + "/1/resources.json"
	s := napping.Session{}
	h := &http.Header{}
	h.Set("Content-Type", "application/json")
	h.Set("Authorization", "Bearer "+token)
	s.Header = h
	glog.V(2).Infof("Headers %+v", h)
	var params napping.Params
	params = napping.Params(queryParams)
	p := params.AsUrlValues()
	glog.V(2).Infof("Query parameters %+v", p)
	resources := ResourceList{}
	resp, err := s.Get(uri, &p, &resources, nil)
	if err != nil {
		glog.Fatal(err)
	}
	glog.V(2).Infof("Response: %s", resp.RawText())
	return resources, err
}

// GetToken given a clientID and secret and username returns a token
func GetToken(clientID string, secret string, username string, uri string) (string, error) {
	v := url.Values{}
	if clientID == "" {
		clientID = os.Getenv("CLIENT_ID")
	}
	v.Set("client_id", clientID)
	if secret == "" {
		secret = os.Getenv("CLIENT_SECRET")
	}
	v.Set("secret", secret)
	if username == "" {
		username = os.Getenv("USERNAME")
	}
	v.Set("username", username)
	if uri == "" {
		uri = os.Getenv("PARTNER_BASE_URI") + "/1/oauth/get_token"
	}
	glog.V(1).Infof("Getting token for %s", clientID)
	glog.V(1).Infof("To URL %s", uri)
	resp, err := http.PostForm(uri, v)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	var data map[string]string
	json.Unmarshal(body, &data)
	token := data["access_token"]
	return token, err
}

// GetResource fills a Resource structure with the values given the OpenEd resource_id
func (resource *Resource) GetResource(db sqlx.DB) error {
	query := "SELECT ID,Title,Publisher_id,Contribution_id,Description,Resource_type_id,Youtube_id FROM resources WHERE ID=" + strconv.Itoa(resource.ID)
	glog.V(3).Infof("Querying with: %s", query)
	err := db.Get(resource, query)

	if err != nil {
		glog.Errorf("Error retrieving resource %d: %+v", resource.ID, err)
		return err
	}
	glog.V(1).Infof("Resource is: %+v", *resource)
	return nil
}

// A Standard has fields that describe an educational standard
/* Standard has CREATE TABLE standards (
    id integer NOT NULL,
    identifier character varying(255),
    "group" character varying(255),
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    category character varying(255),
    description text,
    subcategory text,
    grade character varying(255),
    subject character varying(255),
    number character varying(255),
    category_id integer,
    fullname character varying(255),
    grade_group integer,
    grade_group_id integer,
    playlist character varying(255),
    curated boolean,
    source character varying(255),
    title text,
    modified_at timestamp without time zone,
    sort_key integer,
    substandard_num integer,
    identifier_code character varying(255),
    key_words text,
    more_information text,
    min_grade integer,
    max_grade integer,
    parent_id integer,
    guid character varying(255),
    confirmed_resources_count integer DEFAULT 0 NOT NULL,
    prerequisites integer[] DEFAULT '{}'::integer[]
);*/
type Standard struct {
	ID          int
	Grade       string
	Title       string
	Description string
}

// GetStandard fills in fields in standards structure
func (standard *Standard) GetStandard(db sqlx.DB) error {
	var query string
	query = "SELECT ID,Grade,Title,Description FROM Standards WHERE ID=" + strconv.Itoa(standard.ID)
	err := db.Get(standard, query)

	if err != nil {
		glog.Errorf("Error retrieving standards %d: %+v", standard.ID, err)
		return err
	}
	glog.V(3).Infof("Standard is: %+v", *standard)
	return nil
}

// ResourcesShareStandard tests if a supplied resources shares a standard with the
// resource used.  Returns true if they share standards
func (resource *Resource) ResourcesShareStandard(db sqlx.DB, resource2 Resource) bool {
	queryBase := "SELECT standard_id FROM alignments WHERE resource_id="
	query1 := queryBase + strconv.Itoa(resource.ID)
	standards1 := []int{}
	err := db.Select(&standards1, query1)
	if err != nil {
		glog.Errorf("Couldn't retrieve standards for %d ", resource.ID)
		return false
	}
	query2 := queryBase + strconv.Itoa(resource2.ID)
	standards2 := []int{}
	err = db.Select(&standards2, query2)
	if err != nil {
		glog.Errorf("Couldn't retrieve standards for %d ", resource2.ID)
		return false
	}
	for _, i := range standards1 {
		for _, x := range standards2 {
			if i == x {
				glog.V(2).Infof("Resources %d,%d do share standard %d", resource.ID, resource2.ID, i)
				return true
			}
		}
	}
	glog.V(3).Infof("Resources do not share standard")
	return false
}

// ResourcesShareCategory tests if a supplied resources shares a standard category with the
// resource used.  Returns true if they share category
func (resource Resource) ResourcesShareCategory(db sqlx.DB, resource2 Resource) bool {
	queryBase := "SELECT DISTINCT(category_id) FROM alignments INNER JOIN standards ON standards.ID=alignments.standard_id AND resource_id="
	query1 := queryBase + strconv.Itoa(resource.ID)
	categories1 := []int{}
	glog.V(3).Infof("Querying categories for %d: %s", resource.ID, query1)
	err := db.Select(&categories1, query1)
	if err != nil {
		glog.Errorf("Couldn't retrieve categories for %d:%+v ", resource.ID, err)
		return false
	}
	glog.V(3).Infof("Retrieved categories: %+v", categories1)
	query2 := queryBase + strconv.Itoa(resource2.ID)
	categories2 := []int{}
	glog.V(3).Infof("Querying categories for %d: %s", resource2.ID, query2)
	err = db.Select(&categories2, query2)
	if err != nil {
		glog.Errorf("Couldn't retrieve categories for %d ", resource2.ID)
		return false
	}
	for _, i := range categories1 {
		glog.V(3).Infof("First category: %d", i)
		for _, x := range categories2 {
			glog.V(3).Infof("Second category: %d", x)
			if i == x {
				glog.V(2).Infof("Resources %d,%d share category: %d", resource.ID, resource2.ID, i)
				return true
			}
		}
	}

	glog.V(3).Infof("Resources do not share category")
	return false
}

// ResourcesShareSubject checks if resource that is receiver and second resource share a subject
func (resource Resource) ResourcesShareSubject(db sqlx.DB, resource2 Resource) bool {
	queryBase := "SELECT subject_id FROM resources_subjects WHERE resources_subjects.resource_id="
	query1 := queryBase + strconv.Itoa(resource.ID)
	subjects1 := []int{}
	glog.V(3).Infof("Querying subjects for %d: %s", resource.ID, query1)
	err := db.Select(&subjects1, query1)
	if err != nil {
		glog.Errorf("Couldn't retrieve subjects for %d:%+v ", resource.ID, err)
		return false
	}
	glog.V(3).Infof("Retrieved subjects: %+v", subjects1)
	query2 := queryBase + strconv.Itoa(resource2.ID)
	subjects2 := []int{}
	glog.V(3).Infof("Querying subjects for %d: %s", resource2.ID, query2)
	err = db.Select(&subjects2, query2)
	if err != nil {
		glog.Errorf("Couldn't retrieve categories for %d ", resource2.ID)
		return false
	}
	for _, i := range subjects1 {
		glog.V(3).Infof("First resource subjects: %d", i)
		for _, x := range subjects2 {
			glog.V(3).Infof("Second resource subjects: %d", x)
			if i == x {
				glog.V(2).Infof("Resources %d,%d share category: %d", resource.ID, resource2.ID, i)
				return true
			}
		}
	}
	glog.V(3).Infof("Resources do not share category")
	return false
}

// User is type for OpenEd db user table
type User struct {
	ID            sql.NullInt64
	Email         sql.NullString
	Username      sql.NullString
	Role          sql.NullString
	DistrictState sql.NullString `db:"district_state"`
	Provider      sql.NullString
	GradesRange   sql.NullString `db:"grades_range"`
}

// ListUsers retrieves all users with assessments
func ListUsers(db sqlx.DB) ([]User, error) {
	// retrieve only users with assessment runs
	query := "SELECT distinct(users.ID),email,username,role,district_state,provider,grades_range FROM users INNER JOIN assessment_runs ON (users.ID=assessment_runs.user_id)"
	users := []User{}
	err := db.Select(&users, query)
	if err != nil {
		glog.Errorf("Error retrieving users: %v", err)
		return nil, err
	}
	glog.Infof("Retrieved %d users", len(users))
	return users, err
}

// An AssessmentRun has selected important information stored in OpenEd AssessmentRuns table.
type AssessmentRun struct {
	ID           int
	UserID       int       `db:"user_id"`
	FinishedAt   time.Time `db:"finished_at"`
	AssessmentID int       `db:"assessment_id"`
	Score        float32   `db:"score"`
	FirstRun     bool      `db:"first_run"`
}

// ListAssessmentRuns shows all assessment runs in database for a given grade
func ListAssessmentRuns(db sqlx.DB, grade string) ([]AssessmentRun, error) {
	// retrieve only users with assessment runs
	query := `SELECT distinct(a.id),a.user_id,a.finished_at,a.assessment_id,a.score,a.first_run
		FROM assessment_runs a INNER JOIN resources ON resources.ID=a.ID
		WHERE finished_at is not null and score>0 `
	if grade != "" {
		if grade == "K" {
			grade = "0"
		}
		query = fmt.Sprintf("%s AND min_grade<=%s and max_grade>=%s", query, grade, grade)
	}
	glog.V(2).Infof("Query for assessment runs: %s", query)
	runs := []AssessmentRun{}
	err := db.Select(&runs, query)
	if err != nil {
		glog.Errorf("Error retrieving run: %+v", err)
		return nil, err
	}
	glog.Infof("Retrieved %d runs", len(runs))
	return runs, err
}

// An Alignment has information on resource and what standard its aligned to
type Alignment struct {
	ID         int
	ResourceID int `db:"resource_id"`
	StandardID int `db:"standard_id"`
	Status     int
}

// GetAlignments retrieves all standard alignments for a given resource
func (resource Resource) GetAlignments(db sqlx.DB) []int {
	query := "SELECT standard_id FROM alignments WHERE resource_id=" + strconv.Itoa(resource.ID)
	standards := []int{}
	err := db.Select(&standards, query)
	if err != nil {
		glog.Errorf("Error retrieving standards: %+v", err)
		return nil
	}
	return standards
}

// A UserEvent has information on the user and what action they performed.
type UserEvent struct {
	ID              int
	UserID          int `db:"user_id"`
	UserEventTypeID int `db:"user_event_type_id"`
	RefUserID       int `db:"ref_user_id"`
	Value           string
	CreatedAt       time.Time `db:"created_at"`
	URL             string
}

// StandardGroup has the name and count of top level standard groups
type StandardGroup struct {
	ID          int
	Title       string
	GradesRange string `json:"grades_range"`
	AreaID      int    `json:"area_id"`
}

// StandardGroupList is structure get back list of standard groups
type StandardGroupList struct {
	StandardGroups []StandardGroup `json:"standard_groups"`
}

// ListStandardGroups lists all of the standard groups
func ListStandardGroups(token string) (StandardGroupList, error) {
	// SearchResources searches OpenEd for resources given set of queryParams.
	var err error
	uri := os.Getenv("PARTNER_BASE_URI") + "/1/standard_groups.json"
	s := napping.Session{}
	h := &http.Header{}
	h.Set("Content-Type", "application/json")
	h.Set("Authorization", "Bearer "+token)
	s.Header = h
	glog.V(2).Infof("Headers %+v", h)
	groups := StandardGroupList{}
	glog.V(2).Infof("Hitting URI %s", uri)

	resp, err := s.Get(uri, nil, &groups, nil)
	if err != nil {
		glog.V(2).Infof("Error: %+v", err)
		glog.Fatal(err)
	}
	glog.V(2).Infof("Response: %s", resp.RawText())
	glog.V(2).Infof("Groups: %+v", groups)
	return groups, err
}

// GradeGroup has info on a grade group such as Elementary
type GradeGroup struct {
	ID          int
	Title       string
	GradesRange string `json:"grades_range"`
}

// GradeGroupList is structure get back list of standard groups
type GradeGroupList struct {
	GradeGroups []GradeGroup `json:"grade_groups"`
}

// ListGradeGroups lists all of the standard groups
func ListGradeGroups(ID int, token string) (GradeGroupList, error) {
	// SearchResources searches OpenEd for resources given set of queryParams.
	var err error
	uri := os.Getenv("PARTNER_BASE_URI") + "/1/grade_groups.json"
	s := napping.Session{}
	h := &http.Header{}
	h.Set("Content-Type", "application/json")
	h.Set("Authorization", "Bearer "+token)
	s.Header = h
	glog.V(2).Infof("Headers %+v", h)
	groups := GradeGroupList{}
	glog.V(2).Infof("Hitting URI %s", uri)

	var params napping.Params
	params = napping.Params{}
	params["standard_group"] = strconv.Itoa(ID)
	p := params.AsUrlValues()
	glog.V(2).Infof("Query parameters %+v", p)

	resp, err := s.Get(uri, &p, &groups, nil)
	if err != nil {
		glog.V(2).Infof("Error: %+v", err)
		glog.Fatal(err)
	}
	glog.V(2).Infof("Response: %s", resp.RawText())
	glog.V(2).Infof("Groups: %+v", groups)
	return groups, err
}

// DumpResourceRatings writes a file with each resource and its rating for each standard
func DumpResourceRatings(db *sqlx.DB, grade string) (numRatings int, err error) {
	redisConnect := os.Getenv("REDIS_URL")
	redisURL, _ := url.Parse(redisConnect)
	redisPassword := ""
	if redisURL.User != nil {
		redisPassword, _ = redisURL.User.Password()
	}
	redisOptions := goredis.Options{
		Addr:     redisURL.Host,
		Password: redisPassword,
	}
	c := goredis.NewClient(&redisOptions)
	var cursor int64
	var n int
	var keys []string
	var content string
	numRatings = 0
	content = "Resource,Rating\n"

	for {
		cursor, keys, err = c.Scan(cursor, "resource:*", 10).Result()
		if err != nil {
			glog.Fatalf("Scan error: %s", err)
		}
		n += len(keys)
		if cursor == 0 {
			break
		}
		for _, k := range keys {
			ratings := c.HGetAllMap(k)
			fmt.Printf("Resource %s ratings: %+v\n", k, ratings.Val())
			glog.V(1).Infof("Resource %s ratings: %+v\n", k, ratings.Val())
			re, _ := regexp.Compile("[0-9]+")
			resbytes := []byte(k)
			resNum := re.Find(resbytes)
			glog.V(1).Infof("Resource #: %d\n", resNum)
			id, _ := strconv.Atoi(string(resNum))
			r := Resource{ID: id}
			rp := &r
			err = rp.GetResource(*db)

			content = content + fmt.Sprintf("%s,", r.URL.String)
			for stdID, rating := range ratings.Val() {
				ID, _ := strconv.Atoi(stdID)
				s := Standard{ID: ID}
				sp := &s
				err = sp.GetStandard(*db)
				content = content + fmt.Sprintf("%s,%s,", s.Title, rating)
			}
			content = content + "\n"
			numRatings = numRatings + 1
		}
	}
	glog.V(1).Infof("Found %d keys\n", n)
	filename := fmt.Sprintf("%s-%s", grade, "ratings.csv")
	glog.V(1).Infof("Writing result to %s\n", filename)
	S3WriteFile(filename, content)
	return numRatings, err
}

// S3WriteFile write specified content to file with filename
func S3WriteFile(filename string, content string) error {
	glog.V(2).Infof("Write to file %s: %s", filename, content)
	svc := s3.New(session.New(), &aws.Config{Region: aws.String("us-east-1")}) // implicit works with AWS_ACCESS__KEY_ID and AWS_SECRET_ACCESS_KEY
	bucket := os.Getenv("AWS_S3_BUCKET")

	putParams := s3.PutObjectInput{
		Bucket: &bucket,
		Key:    &filename,
		Body:   bytes.NewReader([]byte(content)),
	}
	_, putErr := svc.PutObject(&putParams)
	if putErr == nil {
		glog.V(2).Infof("Wrote content: %+v", content)
	} else {
		glog.V(2).Infof("Error writing content: %+v", putErr)
	}
	return putErr
}
