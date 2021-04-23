package appstore

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	hrepo "github.com/helm/helm/pkg/repo"
	"github.com/pkg/errors"
	"golang.org/x/sync/singleflight"
	"goodrain.com/cloud-adaptor/internal/domain"
	"sigs.k8s.io/yaml"
)

type AppTemplater interface {
	Fetch(ctx context.Context, appStore *domain.AppStore) ([]*domain.AppTemplate, error)
}

// NewAppTemplater creates a new
func NewAppTemplater() AppTemplater {
	return &helmAppTemplate{}
}

type helmAppTemplate struct {
	singleflight.Group
}

func (h *helmAppTemplate) Fetch(ctx context.Context, appStore *domain.AppStore) ([]*domain.AppTemplate, error) {
	// single flight to avoid cache breakdown
	v, err, _ := h.Do(appStore.Key(), func() (interface{}, error) {
		return h.fetch(ctx, appStore)
	})
	appTemplates := v.([]*domain.AppTemplate)
	return appTemplates, err
}
func (h *helmAppTemplate) fetch(ctx context.Context, appStore *domain.AppStore) ([]*domain.AppTemplate, error) {
	req, err := http.NewRequest("GET", appStore.URL+"/index.yaml", nil)
	if err != nil {
		return nil, errors.Wrap(err, "new http request")
	}

	req = req.WithContext(ctx)

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
