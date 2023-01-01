package awssecrets

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/kroonprins/vals/pkg/api"
	"github.com/kroonprins/vals/pkg/awsclicompat"
	"github.com/kroonprins/vals/pkg/providers/util"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

type provider struct {
	api.StaticConfig
	// Keeping track of secretsmanager services since we need a service per region
	client *secretsmanager.SecretsManager

	// AWS SecretsManager global configuration
	Region, VersionStage, VersionId, Profile string

	Format string
}

func New(cfg api.StaticConfig) *provider {
	p := &provider{}
	p.StaticConfig = cfg
	p.Region = cfg.String("region")
	p.VersionStage = cfg.String("version_stage")
	p.VersionId = cfg.String("version_id")
	p.Profile = cfg.String("profile")
	return p
}

// Get gets an AWS SSM Parameter Store value
func (p *provider) GetString(key string) (string, error) {
	cli := p.getClient()

	in := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(key),
	}

	if p.VersionStage != "" {
		in = in.SetVersionStage(p.VersionStage)
	}

	if p.VersionId != "" {
		in = in.SetVersionId(p.VersionId)
	}

	out, err := cli.GetSecretValue(in)
	if err != nil {
		return "", fmt.Errorf("get parameter: %v", err)
	}

	var v string
	if out.SecretString != nil {
		v = *out.SecretString
	} else if out.SecretBinary != nil {
		v = string(out.SecretBinary)
	} else {
		return "", errors.New("awssecrets: get secret value: no SecretString nor SecretBinary is set")
	}

	p.debugf("awssecrets: successfully retrieved key=%s", key)

	return v, nil
}

func (p *provider) GetStringMap(key string) (map[string]interface{}, error) {

	str, err := p.GetString(key)
	if err == nil {
		return util.Unmarshal(p.StaticConfig, []byte(str))
	}

	metaKey := strings.TrimRight(key, "/") + "/meta"

	str, err = p.GetString(metaKey)
	if err != nil {
		return nil, err
	}

	meta, err := util.Unmarshal(p.StaticConfig, []byte(str))
	if err != nil {
		return nil, err
	}

	metaKeysField := "github.com/variantdev/vals"
	f, ok := meta[metaKeysField]
	if !ok {
		return nil, fmt.Errorf("%q not found", metaKeysField)
	}

	var suffixes []string
	switch f := f.(type) {
	case []string:
		suffixes = append(suffixes, f...)
	case []interface{}:
		for _, v := range f {
			suffixes = append(suffixes, fmt.Sprintf("%v", v))
		}
	default:
		return nil, fmt.Errorf("%q was not a kind of array: value=%v, type=%T", suffixes, suffixes, suffixes)
	}
	if !ok {
		return nil, fmt.Errorf("%q was not a string array", metaKeysField)
	}

	res := map[string]interface{}{}
	for _, suf := range suffixes {
		sufKey := strings.TrimLeft(suf, "/")
		full := strings.TrimRight(key, "/") + "/" + sufKey
		str, err := p.GetString(full)
		if err != nil {
			return nil, err
		}
		res[sufKey] = str
	}

	p.debugf("SSM: successfully retrieved key=%s", key)

	return res, nil
}

func (p *provider) debugf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
}

func (p *provider) getClient() *secretsmanager.SecretsManager {
	if p.client != nil {
		return p.client
	}

	sess := awsclicompat.NewSession(p.Region, p.Profile)

	p.client = secretsmanager.New(sess)
	return p.client
}
