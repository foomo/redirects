package redirectstore

type RedirectSource string
type RedirectID string
type RedirectTarget string
type RedirectRequest string

type RedirectDefinition struct {
	ID             RedirectID   `json:"id" bson:"id"`
	Source         string       `json:"source" bson:"source"`
	Target         string       `json:"target" bson:"target"`
	Code           RedirectCode `json:"code" bson:"code"`
	RespectParams  bool         `json:"respectparams" bson:"respectparams"`
	TransferParams bool         `json:"transferparams" bson:"transferparams"`
}

type RedirectDefinitions map[RedirectID]*RedirectDefinition
