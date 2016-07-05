package opened

import (
	"flag"
	"os"
	"testing"

	"github.com/golang/glog"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func TestDumpResourceRatings(t *testing.T) {
	db := setup()
	grade := "K"
	numRatings, _ := DumpResourceRatings(db, grade)
	teardown(db)
	glog.V(2).Infof("Number of ratings: %d\n", numRatings)
}

func TestListAssessmentRuns(t *testing.T) {
	db := setup()
	grade := "K"
	glog.V(1).Infof("Looking for runs in grade: %s", grade)
	runs, _ := ListAssessmentRuns(*db, grade)
	glog.V(1).Infof("Got %d runs", len(runs))
	teardown(db)
}

func TestListStandardGroups(t *testing.T) {
	token, _ := setupWs()
	results, err := ListStandardGroups(token)
	if err != nil {
		t.Errorf("Error from ListStandardGroups: %+v", err)
	}
	glog.V(1).Infof("%d groups returned", len(results.StandardGroups))
}

func TestListGradeGroups(t *testing.T) {
	token, _ := setupWs()
	sgResults, err := ListStandardGroups(token)
	results, err := ListGradeGroups(sgResults.StandardGroups[0].ID, token)
	if err != nil {
		t.Errorf("Error from ListGradeGroups: %+v", err)
	}
	glog.V(1).Infof("%d grade groups returned", len(results.GradeGroups))
}

// TestSearchResources calls SearchResources with some query parameters and checks if it gets back results
func TestSearchResources(t *testing.T) {
	token, _ := setupWs()
	queryParams := make(map[string]string)
	queryParams["descriptive"] = "counting"
	queryParams["grades_range"] = "K-1"
	results, err := SearchResources(queryParams, token)
	if err != nil {
		t.Errorf("Error from SearchResources: %+v", err)
	}
	glog.V(1).Infof("%d results returned", len(results.Resources))
	glog.V(2).Infof("First result: %+v", results.Resources[0])
}

// setup_ws sets up test for OpenEd package calls which use web services (instead of database)
func setupWs() (string, error) {
	flag.Set("alsologtostderr", "true")
	flag.Set("v", "3")
	token, err := GetToken("", "", "", "")
	return token, err
}

func TestGetToken(t *testing.T) {
	flag.Set("alsologtostderr", "true")
	flag.Set("v", "3")
	token, err := GetToken("", "", "", "")
	if err != nil {
		t.Errorf("Failed to get token: %+v", err)
	} else {
		glog.V(1).Infof("Got token %s", token)
	}
}

func TestListUsers(t *testing.T) {
	db := setup()
	users, err := ListUsers(*db)
	if err != nil {
		t.Errorf("Failed to get users: %+v", err)
	}
	glog.V(1).Infof("Got %d users", len(users))
	teardown(db)
}

func TestResourcesShareStandard(t *testing.T) {
	db := setup()
	// these resources DONT share
	r1 := Resource{ID: 183189}
	r2 := Resource{ID: 2043501}
	if r1.ResourcesShareStandard(*db, r2) == true {
		// bad error!
		t.Errorf("Resources %d and %d share standard!", r1.ID, r2.ID)
	} else {
		glog.V(2).Infof("Resources %d and %d do not share standard!", r1.ID, r2.ID)
	}

	r1 = Resource{ID: 4123630}
	r2 = Resource{ID: 4123755}
	if r1.ResourcesShareStandard(*db, r2) == true {
		glog.V(2).Infof("Resources %d and %d DO share standard!", r1.ID, r2.ID)
	} else {
		t.Errorf("Resources %d and %d do NOT share standard!", r1.ID, r2.ID)
	}
	teardown(db)
}

func TestResourcesShareCategory(t *testing.T) {
	db := setup()
	r1 := Resource{ID: 4123630}
	r2 := Resource{ID: 178375}
	if r1.ResourcesShareCategory(*db, r2) == true {
		glog.V(2).Infof("Resources %d and %d share category!", r1.ID, r2.ID)
	} else {
		glog.V(2).Infof("Resources %d and %d do not share category!", r1.ID, r2.ID)
	}
	teardown(db)
}

func TestGetResource(t *testing.T) {
	db := setup()
	r := Resource{ID: 183189}
	rp := &r
	err := rp.GetResource(*db)

	if err != nil {
		t.Errorf("Failed to get resource: %+v", err)
	}
	glog.V(1).Infof("Got resource: %+v", r)
	teardown(db)
}

func setup() *sqlx.DB {
	flag.Set("alsologtostderr", "true")
	flag.Set("v", "3")

	// connect to Postgres to get assessment runs and resource usages
	dbConnect := os.Getenv("FOLLOWER_DATABASE_URL")
	db, err := sqlx.Connect("postgres", dbConnect)
	if err != nil {
		glog.Fatalln(err)
	}
	glog.V(2).Infof("Connected to database: %s", dbConnect)
	return db
}

func teardown(db *sqlx.DB) {
	db.Close()
}
