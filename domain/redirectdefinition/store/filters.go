package redirectstore

type RedirectionType string
type ActiveStateType string

const (
	ActiveStateAll      ActiveStateType = "all"
	ActiveStateEnabled  ActiveStateType = "enabled"
	ActiveStateDisabled ActiveStateType = "disabled"
)

func (a ActiveStateType) IsValid() bool {
	return a == ActiveStateEnabled || a == ActiveStateDisabled || a == ActiveStateAll
}

func (a ActiveStateType) ToFilter() (interface{}, bool) {
	switch a {
	case ActiveStateEnabled:
		return true, true
	case ActiveStateDisabled:
		return false, true
	default: // ActiveStateAll
		return nil, false
	}
}

const (
	RedirectionTypeAll       RedirectionType = "all"
	RedirectionTypeManual    RedirectionType = "manual"
	RedirectionTypeAutomatic RedirectionType = "automatic"
)

func (r RedirectionType) IsValid() bool {
	return r == RedirectionTypeAutomatic || r == RedirectionTypeManual || r == RedirectionTypeAll
}

func (r RedirectionType) ToFilter() (interface{}, bool) {
	switch r {
	case RedirectionTypeManual:
		return RedirectionTypeManual, true
	case RedirectionTypeAutomatic:
		return RedirectionTypeAutomatic, true
	default: // RedirectionTypeAll
		return nil, false
	}
}
