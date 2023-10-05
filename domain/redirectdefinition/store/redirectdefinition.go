package redirectstore

type RedirectSource string
type RedirectID string
type RedirectTarget string
type RedirectRequest string

type RedirectDefinition struct {
	ID             RedirectID     `json:"id" bson:"id"`
	Source         RedirectSource `json:"source" bson:"source"`
	Target         RedirectTarget `json:"target" bson:"target"`
	Code           RedirectCode   `json:"code" bson:"code"`
	RespectParams  bool           `json:"respectparams" bson:"respectparams"`
	TransferParams bool           `json:"transferparams" bson:"transferparams"`
}

type RedirectDefinitions map[RedirectSource]*RedirectDefinition
