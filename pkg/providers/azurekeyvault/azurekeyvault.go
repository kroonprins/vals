package azurekeyvault

import (
	"context"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/keyvault/azsecrets"

	"github.com/kroonprins/vals/pkg/api"
	"github.com/kroonprins/vals/pkg/azureclicompat"
	"github.com/kroonprins/vals/pkg/providers/util"
)

type provider struct {
	api.StaticConfig
	// azure key vault client
	clients map[string]*azsecrets.Client
}

func New(cfg api.StaticConfig) *provider {
	return &provider{
		StaticConfig: cfg,
		clients:      make(map[string]*azsecrets.Client),
	}
}

func (p *provider) GetString(key string) (string, error) {
	spec, err := parseKey(key)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(spec.secretName) == "" {
		return "", fmt.Errorf("missing secret name: %q", key)
	}

	client, err := p.getClientForKeyVault(spec.vaultBaseURL)
	if err != nil {
		return "", err
	}

	secretBundle, err := client.GetSecret(context.Background(), spec.secretName, spec.secretVersion, nil)
	if err != nil {
		return "", err
	}
	return *secretBundle.Value, err
}

func (p *provider) GetStringMap(key string) (map[string]interface{}, error) {
	spec, err := parseKey(key)
	if err != nil {
		return nil, err
	}
	if spec.secretName != "" {
		str, err := p.GetString(key)
		if err != nil {
			return nil, err
		}

		return util.Unmarshal(p.StaticConfig, []byte(str))
	} else {
		client, err := p.getClientForKeyVault(spec.vaultBaseURL)
		if err != nil {
			return nil, err
		}

		mp := make(map[string]interface{})
		pager := client.NewListSecretsPager(&azsecrets.ListSecretsOptions{})

		for pager.More() {
			page, err := pager.NextPage(context.Background())
			if err != nil {
				return nil, fmt.Errorf("failed to retrieve secrets from vault '%s': %v", spec.vaultBaseURL, err)
			}
			for _, secret := range page.Value {
				if secret.Managed != nil && *secret.Managed {
					continue
				}
				secretVal, err := p.GetString(fmt.Sprintf("%s/%s", key, secret.ID.Name()))
				if err != nil {
					return nil, err
				}
				mp[secret.ID.Name()] = secretVal
			}
		}

		return mp, nil
	}
}

func (p *provider) getClientForKeyVault(vaultBaseURL string) (*azsecrets.Client, error) {
	if val, ok := p.clients[vaultBaseURL]; val != nil || ok {
		return p.clients[vaultBaseURL], nil
	}

	cred, err := getTokenCredential()
	if err != nil {
		return nil, err
	}

	client, err := azsecrets.NewClient(vaultBaseURL, cred, nil)
	if err != nil {
		return nil, err
	}
	p.clients[vaultBaseURL] = client
	return client, nil
}

func getTokenCredential() (azcore.TokenCredential, error) {
	cred, err := azureclicompat.ResolveIdentity()
	if err != nil {
		return nil, err
	}

	return cred, nil
}

type secretSpec struct {
	vaultBaseURL  string
	secretName    string
	secretVersion string
}

func parseKey(key string) (spec secretSpec, err error) {
	components := strings.Split(strings.TrimSuffix(key, "/"), "/")
	if len(components) < 1 || len(components) > 3 {
		err = fmt.Errorf("invalid secret specifier: %q", key)
		return
	}

	if strings.TrimSpace(components[0]) == "" {
		err = fmt.Errorf("missing key vault name: %q", key)
		return
	}

	spec.vaultBaseURL = makeEndpoint(components[0])
	if len(components) > 1 {
		spec.secretName = components[1]
	}
	if len(components) > 2 {
		spec.secretVersion = components[2]
	}
	return
}

func makeEndpoint(endpoint string) string {
	endpoint = "https://" + endpoint
	if !strings.Contains(endpoint, ".") {
		endpoint += ".vault.azure.net"
	}
	return endpoint
}
