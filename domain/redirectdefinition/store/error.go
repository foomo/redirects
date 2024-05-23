package redirectstore

type RedirectDefinitionError string

func (r *RedirectDefinitionError) Error() string {
	return string(*r)
}
