PACKAGE DOCUMENTATION

package opened
    import "github.com/openedinc/opened"

    Package opened provides structures for OpenEd objects such as Resources
    and Standards

TYPES

type Alignment struct {
    Id          int
    Resource_id int
    Standard_id int
    Status      int
}
    An Alignment has information on resource and what standard its aligned
    to

type AssessmentRun struct {
    Assessment_id int
    Id            int
    User_id       int
    Score         float32
    First_run     bool
    Finished_at   time.Time
}
    An AssessmentRun has selected important information stored in OpenEd
    AssessmentRuns table.

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
    A Resource has information such as Publisher, Title, Description for
    video, game or assessment

func (r Resource) GetAlignments(db sqlx.DB, resource_id int) []int
    GetAlignments retrieves all standard alignments for a given resource

func (r Resource) GetResource(db sqlx.DB, resource_id int) Resource
    GetResource fills a Resource structure with the values given the OpenEd
    resource_id

func (resource1 Resource) ResourcesShareStandard(db sqlx.DB, resource2 Resource) bool
    ResourcesShareStandard tests if a supplied resources shares a standard
    with the resource that is the resource. Returns true if they share
    standards

type UserEvent struct {
    Id                 int
    User_id            int
    User_event_type_id int
    Ref_user_id        int
    Value              string
    Created_at         time.Time
    Url                string
}
    A UserEvent has information on the user and what action they performed.


