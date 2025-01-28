package redirectstore

type RedirectionType string
type ActiveStateType string

const (
	ActiveStateTypeAll      ActiveStateType = "all"
	ActiveStateTypeEnabled  ActiveStateType = "enabled"
	ActiveStateTypeDisabled ActiveStateType = "disabled"
)

func (a ActiveStateType) IsValid() bool {
	return a == ActiveStateTypeEnabled || a == ActiveStateTypeDisabled || a == ActiveStateTypeAll
}

func (a ActiveStateType) ToFilter() (interface{}, bool) {
	switch a {
	case ActiveStateTypeEnabled:
		return false, true
	case ActiveStateTypeDisabled:
		return true, true
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
