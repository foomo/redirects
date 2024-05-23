package redirectstore

type RedirectDefinitionError string

func NewRedirectDefinitionError(err string) *RedirectDefinitionError {
	e := RedirectDefinitionError(err)
	return &e
}

func (r *RedirectDefinitionError) Error() string {
	return string(*r)
}
