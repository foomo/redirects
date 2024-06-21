package redirectstore

type RedirectSource string
type RedirectTarget string
type RedirectRequest string
type RedirectionType string
type Dimension string
type Site string

const (
	Manual    RedirectionType = "manual"
	Automatic RedirectionType = "automatic"
)

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
}

type RedirectDefinitions map[RedirectSource]*RedirectDefinition
