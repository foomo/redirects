package redirection

type RedirectSource string
type RedirectID string
type RedirectTarget string
type RedirectRequest string
type RedirectResponse string
type RedirectCode int

const (
	RedirectCodePermanent RedirectCode = 301
	RedirectCodeTemporary RedirectCode = 307 // will this be needed?
)

type RedirectDefinition struct {
	ID             RedirectID     `json:"id" bson:"id"`
	Source         RedirectSource `json:"source" bson:"source"`
	Target         RedirectTarget `json:"target" bson:"target"`
	Code           RedirectCode   `json:"code" bson:"code"`
	RespectParams  bool           `json:"respectparams" bson:"respectparams"`
	TransferParams bool           `json:"transferparams" bson:"transferparams"`
}

type Redirect struct {
	Response RedirectResponse
	Code     RedirectCode
}

func (r RedirectCode) Valid() bool {
	switch r {
	case
		RedirectCodePermanent:
		return true
	case
		RedirectCodeTemporary: // will this be needed
		return true
	default:
		return false
	}
}
