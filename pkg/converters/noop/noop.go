package noop

type provider struct {
}

func New() *provider {
	p := &provider{}
	return p
}

func (p *provider) Convert(value string) (string, error) {
	return value, nil
}
