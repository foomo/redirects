package redirectstore

type RedirectSource string
type RedirectTarget string
type RedirectRequest string
type Dimension string
type Site string

type RedirectDefinition struct {
	ID              EntityID        `json:"id" bson:"id"`
	ContentID       string          `json:"contentId" bson:"contentId"`
	Source          RedirectSource  `json:"source" bson:"source"`
	Target          RedirectTarget  `json:"target" bson:"target"`
	Code            RedirectCode    `json:"code" bson:"code"`
	RespectParams   bool            `json:"respectparams" bson:"respectparams"`
	TransferParams  bool            `json:"transferparams" bson:"transferparams"`
	RedirectionType RedirectionType `json:"redirectType" bson:"redirectType"`
	Dimension       Dimension       `json:"dimension" bson:"dimension"`
	Stale           bool            `json:"stale" bson:"stale"`
	Updated         DateTime        `json:"updated,omitempty" bson:"updated"`             // Timestamp of the last update
	LastUpdatedBy   string          `json:"lastUpdatedBy,omitempty" bson:"lastUpdatedBy"` // User who made the last update
}

type RedirectDefinitions map[RedirectSource]*RedirectDefinition
