package yaml

import (
	"gopkg.in/yaml.v3"
)

type unmarshaller struct{}

func New() *unmarshaller {
	return &unmarshaller{}
}

func (*unmarshaller) Unmarshal(bytes []byte) (map[string]interface{}, error) {
	m := map[string]interface{}{}
	if err := yaml.Unmarshal(bytes, &m); err != nil {
		return nil, err
	}
	return m, nil
}
