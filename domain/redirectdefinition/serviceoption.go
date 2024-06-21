package redirectdefinition

type enabledFunc func() bool

func defaultEnabledFunc() bool {
	return true
}

type ServiceOption func(*Service)

func WithEnabledFunc(f enabledFunc) ServiceOption {
	return func(s *Service) {
		s.enableCreationOfAutomaticRedirects = f
	}
}
