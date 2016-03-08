// Package opened provides structures for OpenEd objects such as Resources and Standards
package opened

import (
  "strconv"
  "time"
  "github.com/golang/glog"
  "github.com/jmoiron/sqlx"
  _ "github.com/lib/pq"
)

// A Resource has information such as Publisher, Title, Description for video, game or assessment
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

// ResourcesShareStandard tests if a supplied resources shares a standard with the
// resource used.  Returns true if they share standards
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
      for _,i := range standards1 {
        for _,x := range standards2 {
          if i == x {
            return true
          }
        }
      }
    }
  }
  return false
}

// ResourcesShareCategory tests if a supplied resources shares a standard category with the
// resource used.  Returns true if they share category
func (resource1 Resource) ResourcesShareCategory(db sqlx.DB, resource2 Resource) bool {

  query_base := "SELECT DISTINCT(category_id) FROM alignments INNER JOIN standards ON standards.id=alignments.standard_id AND resource_id="
  query1 := query_base + strconv.Itoa(resource1.Id) 
  categories1 := []int{}
  glog.V(2).Infof("Querying categories for %d: %s",resource1.Id,query1)
  err := db.Select(&categories1, query1)
  if err != nil {
    glog.Errorf("Couldn't retrieve categories for %d:%+v ",resource1.Id,err)
    return false
  } else {
    glog.V(2).Infof("Retrieved categories: %+v",categories1)
    query2 := query_base + strconv.Itoa(resource2.Id)
    categories2 := []int{}
    glog.V(2).Infof("Querying categories for %d: %s",resource2.Id,query2)
    err = db.Select(&categories2, query2)
    if err != nil {
      glog.Errorf("Couldn't retrieve categories for %d ", resource2.Id)
      return false
    } else {
      for _,i := range categories1 {
        glog.V(3).Infof("First category: %d",i) 
        for _,x := range categories2 {
          glog.V(3).Infof("Second category: %d",x) 
          if i == x {
            glog.V(3).Infof("Resources share category: %d",i)
            return true
          }
        }
      }
    }
  }
  glog.V(3).Infof("Resources do not share category")
  return false
}


// GetResource fills a Resource structure with the values given the OpenEd resource_id
func (r Resource) GetResource(db sqlx.DB, resource_id int) Resource {
  query := "SELECT FROM resources WHERE id=" + strconv.Itoa(resource_id)
  resource := Resource{}
  err := db.Get(&resource, query)
  if err != nil {
    glog.Errorf("Error retrieving resource: %d", err)
  }
  return resource
}

// GetAlignments retrieves all standard alignments for a given resource
func (r Resource) GetAlignments(db sqlx.DB) []int {
  query := "SELECT standard_id FROM alignments WHERE resource_id=" + strconv.Itoa(r.Id) 
  standards := []int{}
  err := db.Select(&standards, query)
  if err != nil {
    glog.Errorf("Error retrieving standards: %d", err)
    return nil
  }
  return standards
}

// An AssessmentRun has selected important information stored in OpenEd AssessmentRuns table.
type AssessmentRun struct {
  Assessment_id int
  Id            int
  User_id       int
  Score         float32
  First_run     bool
  Finished_at   time.Time
}

// A UserEvent has information on the user and what action they performed.
type UserEvent struct {
  Id                 int
  User_id            int
  User_event_type_id int
  Ref_user_id        int
  Value              string
  Created_at         time.Time
  Url                string
}

// An Alignment has information on resource and what standard its aligned to
type Alignment struct {
  Id          int
  Resource_id int
  Standard_id int
  Status      int
}




