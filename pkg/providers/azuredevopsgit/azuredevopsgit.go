package azuredevopsgit

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/microsoft/azure-devops-go-api/azuredevops"
	"github.com/microsoft/azure-devops-go-api/azuredevops/git"

	"github.com/kroonprins/vals/pkg/api"
	"gopkg.in/yaml.v3"
)

type provider struct {
	client git.Client

	organization string
	project      string
	repository   string
	version      string
	versionType  string
}

func New(cfg api.StaticConfig) *provider {
	return &provider{
		organization: cfg.String("organization"),
		project:      cfg.String("project"),
		repository:   cfg.String("repository"),
		version:      cfg.String("version"),
		versionType:  cfg.String("versionType"),
	}
}

func (p *provider) GetString(key string) (string, error) {
	key = strings.TrimSuffix(key, "/")

	client, err := p.ensureClient()
	if err != nil {
		return "", err
	}

	itemContentArgs := git.GetItemContentArgs{
		RepositoryId: &p.repository,
		Path:         &key,
		Project:      &p.project,
	}
	if p.version != "" {
		versionDescriptor := git.GitVersionDescriptor{
			Version: &p.version,
		}
		if p.versionType != "" {
			switch strings.ToLower(p.versionType) {
			case "branch":
				versionDescriptor.VersionType = &git.GitVersionTypeValues.Branch
			case "commit":
				versionDescriptor.VersionType = &git.GitVersionTypeValues.Commit
			case "tag":
				versionDescriptor.VersionType = &git.GitVersionTypeValues.Tag
			default:
				return "", fmt.Errorf("unknown versionType %s (must be one of branch, commit or tag)", p.versionType)
			}
		}
		itemContentArgs.VersionDescriptor = &versionDescriptor

	}

	reader, err := client.GetItemContent(context.Background(), itemContentArgs)
	if err != nil {
		return "", err
	}
	defer reader.Close()

	res, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return string(res), err
}

func (p *provider) GetStringMap(key string) (map[string]interface{}, error) {
	m := map[string]interface{}{}
	yamlStr, err := p.GetString(key)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal([]byte(yamlStr), &m)
	if err != nil {
		return nil, fmt.Errorf("error while parsing file for key %q as yaml: %v", key, err)
	}
	return m, nil
}

func (p *provider) ensureClient() (git.Client, error) {
	if p.client == nil {
		var organizationUrl string
		if strings.HasPrefix(p.organization, "https://dev.azure.com") {
			organizationUrl = p.organization
		} else {
			organizationUrl = fmt.Sprintf("https://dev.azure.com/%s", p.organization)
		}

		personalAccessToken := os.Getenv("AZURE_DEVOPS_PAT")
		if personalAccessToken == "" {
			return nil, fmt.Errorf("missing AZURE_DEVOPS_PAT environment variable")
		}

		connection := azuredevops.NewPatConnection(organizationUrl, personalAccessToken)

		ctx := context.Background()

		gitClient, err := git.NewClient(ctx, connection)
		if err != nil {
			return nil, fmt.Errorf("failed creating azure devops git client: %v", err)
		}
		p.client = gitClient
	}
	return p.client, nil
}
