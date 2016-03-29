# opened
--
    import "github.com/openedinc/opened-go"

Package opened provides structures for OpenEd objects such as resources and
standards.

## Usage

#### func  GetToken

```go
func GetToken(clientID string, secret string, username string, uri string) (string, error)
```
GetToken given a clientID and secret and username returns a token

#### func  ListAssessmentRuns

```go
func ListAssessmentRuns(db sqlx.DB, grade string) []AssessmentRun
```
ListAssessmentRuns shows all assessment runs in database for a given grade

#### func  ListUsers

```go
func ListUsers(db sqlx.DB) ([]User, error)
```
ListUsers retrieves all users with assessments

#### type Alignment

```go
type Alignment struct {
	ID         int
	ResourceID int `db:"resource_id"`
	StandardID int `db:"standard_id"`
	Status     int
}
```

An Alignment has information on resource and what standard its aligned to

#### type AssessmentRun

```go
type AssessmentRun struct {
	ID           int
	UserID       int       `db:"user_id"`
	FinishedAt   time.Time `db:"finished_at"`
	AssessmentID int       `db:"assessment_id"`
	Score        float32   `db:"score"`
	FirstRun     bool      `db:"first_run"`
}
```

An AssessmentRun has selected important information stored in OpenEd
AssessmentRuns table.

#### type GradeGroup

```go
type GradeGroup struct {
	ID          int
	Title       string
	GradesRange string `json:"grades_range"`
}
```

GradeGroup has info on a grade group such as Elementary

#### type GradeGroupList

```go
type GradeGroupList struct {
	GradeGroups []GradeGroup `json:"grade_groups"`
}
```

GradeGroupList is structure get back list of standard groups

#### func  ListGradeGroups

```go
func ListGradeGroups(ID int, token string) (GradeGroupList, error)
```
ListGradeGroups lists all of the standard groups

#### type Resource

```go
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
```

A Resource has information such as Publisher, Title, Description for video, game
or assessment

#### func (Resource) GetAlignments

```go
func (resource Resource) GetAlignments(db sqlx.DB) []int
```
GetAlignments retrieves all standard alignments for a given resource

#### func (*Resource) GetResource

```go
func (resource *Resource) GetResource(db sqlx.DB) error
```
GetResource fills a Resource structure with the values given the OpenEd
resource_id

#### func (Resource) ResourcesShareCategory

```go
func (resource Resource) ResourcesShareCategory(db sqlx.DB, resource2 Resource) bool
```
ResourcesShareCategory tests if a supplied resources shares a standard category
with the resource used. Returns true if they share category

#### func (*Resource) ResourcesShareStandard

```go
func (resource *Resource) ResourcesShareStandard(db sqlx.DB, resource2 Resource) bool
```
ResourcesShareStandard tests if a supplied resources shares a standard with the
resource used. Returns true if they share standards

#### func (Resource) ResourcesShareSubject

```go
func (resource Resource) ResourcesShareSubject(db sqlx.DB, resource2 Resource) bool
```
ResourcesShareSubject checks if resource that is receiver and second resource
share a subject

#### type ResourceList

```go
type ResourceList struct {
	Resources []WsResource
}
```

ResourceList is a list of WSResources.

#### func  SearchResources

```go
func SearchResources(queryParams map[string]string, token string) (ResourceList, error)
```
SearchResources searches OpenEd for resources given set of queryParams.

#### type StandardGroup

```go
type StandardGroup struct {
	ID          int
	Title       string
	GradesRange string `json:"grades_range"`
	AreaID      int    `json:"area_id"`
}
```

StandardGroup has the name and count of top level standard groups

#### type StandardGroupList

```go
type StandardGroupList struct {
	StandardGroups []StandardGroup `json:"standard_groups"`
}
```

StandardGroupList is structure get back list of standard groups

#### func  ListStandardGroups

```go
func ListStandardGroups(token string) (StandardGroupList, error)
```
ListStandardGroups lists all of the standard groups

#### type User

```go
type User struct {
	ID            sql.NullInt64
	Email         sql.NullString
	Username      sql.NullString
	Role          sql.NullString
	DistrictState sql.NullString `db:"district_state"`
	Provider      sql.NullString
	GradesRange   sql.NullString `db:"grades_range"`
}
```

User is type for OpenEd db user table

#### type UserEvent

```go
type UserEvent struct {
	ID              int
	UserID          int `db:"user_id"`
	UserEventTypeID int `db:"user_event_type_id"`
	RefUserID       int `db:"ref_user_id"`
	Value           string
	CreatedAt       time.Time `db:"created_at"`
	URL             string
}
```

A UserEvent has information on the user and what action they performed.

#### type WsResource

```go
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
```

WsResource is web service queryParams for OpenEd resources (not all attributes
in OpenEd).

### License

This is Free Software, released under the terms of the [GPL v3](http://www.gnu.org/copyleft/gpl.html).
