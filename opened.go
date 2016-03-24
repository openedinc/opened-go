// Package opened provides structures for OpenEd objects such as Resources and Standards
package opened

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/golang/glog"
	"github.com/jmcvetta/napping"
	"github.com/jmoiron/sqlx"
)

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
}

// A Resource has information such as Publisher, Title, Description for video, game or assessment
type Resource struct {
	ID             int
	Title          sql.NullString
	URL            sql.NullString
	PublisherID    sql.NullInt64 `db:"publisher_id"`
	ContributionID sql.NullInt64 `db:"contribution_id"`
	Description    sql.NullString
	ResourceTypeID sql.NullInt64  `db:"resource_type_id"`
	YoutubeID      sql.NullString `db:"youtube_id"`
	UsageCount     sql.NullInt64  `db:"usage_count"`
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
	glog.V(3).Infof("Response %+v", resp)
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
	glog.V(3).Infof("Resource is: %+v", *resource)
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
	runs := []AssessmentRun{}
	err := db.Select(&runs, query)
	if err != nil {
		glog.Errorf("Error retrieving run: %v", err)
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
