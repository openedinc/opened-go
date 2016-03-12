package opened

import (
  "testing"
  "flag"
  "os"
  "github.com/golang/glog"
  "github.com/jmoiron/sqlx"
)

func TestSearchResources(t *testing.T) {
  token,_:=setup_ws()
  query_params:=make(map[string]string)
  query_params["descriptive"]="counting"
  query_params["grades_range"]="K-1"
  results,err:=SearchResources(query_params,token)
  if err!=nil {
    t.Errorf("Error from SearchResources",err)    
  }
  if len(results)==0 {
    t.Errorf("No resources returned!")
  }
}

// setup_ws sets up test for OpenEd package calls which use web services (instead of database)
func setup_ws() (string,error) { 
  flag.Set("alsologtostderr", "true")
  flag.Set("v","3")
  token,err:=GetToken("","","","")
  return token,err
}

func TestGetToken(t *testing.T) {
  flag.Set("alsologtostderr", "true")
  flag.Set("v","3")
  token,err:=GetToken("","","","")
  if err!=nil {
    t.Errorf("Failed to get token: %+v",err)
  } else {
    glog.V(1).Infof("Got token %s",token) 
  }
}

func TestListAssessmentRuns(t *testing.T) {
  db:=setup()
  runs,err:=ListAssessmentRuns(*db)
  if err!=nil {
    t.Errorf("Failed to get runs: %+v",err)
  }
  glog.V(1).Infof("Got %d runs",len(runs))
  teardown(db)  
}

func TestListUsers(t *testing.T) {
  db:=setup()
  users,err:=ListUsers(*db)
  if err!=nil {
    t.Errorf("Failed to get users: %+v",err)
  }
  glog.V(1).Infof("Got %d users",len(users))
  teardown(db)  
}

func TestGetResource(t *testing.T) {
  db := setup()
  r:=Resource{Id:183189}
  rp:=&r
  err:=rp.GetResource(*db)

  if err!=nil {
    t.Errorf("Failed to get resource: %+v",err)
  }
  glog.V(1).Infof("Got resource: %+v",r)
  teardown(db)
}

func TestResourcesShareStandard(t *testing.T) {
  db := setup()
  // these resources DONT share
  r1:=Resource{Id:183189}
  r2:=Resource{Id:2043501}
  if r1.ResourcesShareStandard(*db,r2)==true {
    // bad error!
    t.Errorf("Resources %d and %d share standard!",r1.Id,r2.Id)
  } else {
    glog.V(2).Infof("Resources %d and %d do not share standard!",r1.Id,r2.Id)
  }

  r1=Resource{Id:4123630}
  r2=Resource{Id:4123755}
  if r1.ResourcesShareStandard(*db,r2)==true {
    glog.V(2).Infof("Resources %d and %d DO share standard!",r1.Id,r2.Id)
  } else {
    t.Errorf("Resources %d and %d do NOT share standard!",r1.Id,r2.Id)
  }
  teardown(db)
}

func TestResourcesShareCategory(t *testing.T) {
  db := setup()
  r1:=Resource{Id:4123630}
  r2:=Resource{Id:178375}
  if r1.ResourcesShareCategory(*db,r2)==true {
    glog.V(2).Infof("Resources %d and %d share category!",r1.Id,r2.Id)
  } else {
    glog.V(2).Infof("Resources %d and %d do not share category!",r1.Id,r2.Id)
  }
  teardown(db)
}

func setup() (*sqlx.DB) {
  flag.Set("alsologtostderr", "true")
  flag.Set("v","3")

  // connect to Postgres to get assessment runs and resource usages
  db_connect := os.Getenv("DATABASE_URL")
  db, err := sqlx.Connect("postgres", db_connect)
  if err != nil {
    glog.Fatalln(err)
  }
  glog.V(2).Infof("Connected to database: %s",db_connect)

  return db
}

func teardown(db *sqlx.DB) {
  db.Close()
}