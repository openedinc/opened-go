package opened

import (
  "strconv"
  "time"
  "github.com/golang/glog"
  "github.com/jmoiron/sqlx"
)

type Resource struct {
  Id               int
  Title            string
  Url              string
  Publisher_id     int
  Contribution_id  int
  Description      string
  Resource_type_id int
  Youtube_id       string
}

func (resource1 Resource) ResourcesShareStandard(db sqlx.DB, resource2 Resource) bool {
  query_base := "SELECT standard_id FROM alignments WHERE resource_id="
  query1 := query_base + strconv.Itoa(resource1.Id)
  standards1 := []int{}
  err := db.Select(&standards1, query1)
  if err != nil {
    glog.Errorf("Couldn't retrieve standards for %d ",resource1.Id)
    return false
  } else {
    query2 := query_base + strconv.Itoa(resource2.Id)
    standards2 := []int{}
    err = db.Select(&standards2, query2)
    if err != nil {
      glog.Errorf("Couldn't retrieve standards for %d ", resource2.Id)
      return false
    } else {
      for i := range standards1 {
        for x := range standards2 {
          if i == x {
            return true
          }
        }
      }
    }
  }
  return false
}

func (r Resource) GetResource(db sqlx.DB, resource_id int) Resource {
  query := "SELECT FROM resources WHERE id=" + strconv.Itoa(resource_id)
  resource := Resource{}
  err := db.Get(&resource, query)
  if err != nil {
    glog.Errorf("Error retrieving resource: %d", err)
  }
  return resource
}

func (r Resource) GetAlignments(db sqlx.DB, resource_id int) []int {
  query := "SELECT standard_id FROM alignments WHERE resource_id=" + strconv.Itoa(resource_id)
  standards := []int{}
  err := db.Select(&standards, query)
  if err != nil {
    glog.Errorf("Error retrieving standards: %d", err)
  }
  return standards
}

type AssessmentRun struct {
  Assessment_id int
  Id            int
  User_id       int
  Score         float32
  First_run     bool
  Finished_at   time.Time
}

type UserEvent struct {
  Id                 int
  User_id            int
  User_event_type_id int
  Ref_user_id        int
  Value              string
  Created_at         time.Time
  Url                string
}

type Alignment struct {
  Id          int
  Resource_id int
  Standard_id int
  Status      int
}



