package envsubst

import (
	"strings"

	envSubst "github.com/a8m/envsubst"
	"github.com/kroonprins/vals/pkg/api"
	"github.com/kroonprins/vals/pkg/providers/util"
)

type provider struct {
	api.StaticConfig
}

func New(cfg api.StaticConfig) *provider {
	p := &provider{
		cfg,
	}
	return p
}

func (p *provider) GetString(key string) (string, error) {
	key = strings.TrimSuffix(key, "/")
	key = strings.TrimSpace(key)

	str, err := envSubst.String(key)
	if err != nil {
		return "", err
	}
	return str, nil
}

func (p *provider) GetStringMap(key string) (map[string]interface{}, error) {
	key = strings.TrimSuffix(key, "/")
	key = strings.TrimSpace(key)

	bs, err := envSubst.Bytes([]byte(key))
	if err != nil {
		return nil, err
	}

	return util.Unmarshal(p.StaticConfig, bs)
}
