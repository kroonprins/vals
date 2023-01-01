package util

import (
	"fmt"

	"github.com/kroonprins/vals/pkg/api"
	"github.com/kroonprins/vals/pkg/providers/util/unmarshallers/dotenv"
	"github.com/kroonprins/vals/pkg/providers/util/unmarshallers/yaml"
)

type Unmarshaller interface {
	Unmarshal([]byte) (map[string]interface{}, error)
}

var (
	yamlOrJsonUnmarshaller = yaml.New()
	unmarshallers          = map[string]Unmarshaller{
		"yaml":   yamlOrJsonUnmarshaller,
		"json":   yamlOrJsonUnmarshaller,
		"dotenv": dotenv.New(),
	}
)

func Unmarshal(cfg api.StaticConfig, bytes []byte) (map[string]interface{}, error) {
	format := cfg.String("format")
	if format == "" {
		return yamlOrJsonUnmarshaller.Unmarshal(bytes)
	}
	unmarshaller, exists := unmarshallers[format]
	if !exists {
		return nil, fmt.Errorf("no unmarshaller exists for: %s", format)
	}
	return unmarshaller.Unmarshal(bytes)
}
