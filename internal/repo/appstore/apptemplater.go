package appstore

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	hrepo "github.com/helm/helm/pkg/repo"
	"github.com/pkg/errors"
	"goodrain.com/cloud-adaptor/internal/domain"
	"sigs.k8s.io/yaml"
)

type AppTemplater interface {
	Fetch() ([]*domain.AppTemplate, error)
}

// NewAppTemplater creates a new
func NewAppTemplater(ctx context.Context, appStore *domain.AppStore) AppTemplater {
	return &helmAppTemplate{
		ctx:      ctx,
		appStore: appStore,
	}
}

type helmAppTemplate struct {
	ctx      context.Context
	appStore *domain.AppStore
}

func (h *helmAppTemplate) Fetch() ([]*domain.AppTemplate, error) {
	req, err := http.NewRequest("GET", h.appStore.URL+"/index.yaml", nil)
	if err != nil {
		return nil, errors.Wrap(err, "new http request")
	}

	req = req.WithContext(h.ctx)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "do request")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "read response body")
	}

	jbody, err := yaml.YAMLToJSON(body)
	if err != nil {
		return nil, errors.Wrap(err, "converts YAML to JSON")
	}

	var indexFile hrepo.IndexFile
	if err := json.Unmarshal(jbody, &indexFile); err != nil {
		return nil, errors.Wrap(err, "read index file")
	}
	if len(indexFile.Entries) == 0 {
		return nil, errors.New("entries not found")
	}

	var appTemplates []*domain.AppTemplate
	for name, versions := range indexFile.Entries {
		appTemplate := &domain.AppTemplate{
			Name:     name,
			Versions: versions,
		}
		appTemplates = append(appTemplates, appTemplate)
	}

	return appTemplates, nil
}
