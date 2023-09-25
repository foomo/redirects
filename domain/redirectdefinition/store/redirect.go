package redirectstore

const (
	RedirectCodePermanent RedirectCode = 301
	RedirectCodeTemporary RedirectCode = 307 // will this be needed?
)

type RedirectResponse string
type RedirectCode int
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
