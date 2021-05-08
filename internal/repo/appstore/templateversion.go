package appstore

import (
	"github.com/goodrain/rainbond/pkg/helm"
	"github.com/pkg/errors"
	"goodrain.com/cloud-adaptor/cmd/cloud-adaptor/config"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
)

// TemplateVersioner -
type TemplateVersioner struct {
	repoFile  string
	repoCache string
}

func NewTemplateVersioner(cfg *config.Config) *TemplateVersioner {
	return &TemplateVersioner{
		repoFile:  cfg.Helm.RepoFile,
		repoCache: cfg.Helm.RepoCache,
	}
}

func (t *TemplateVersioner) LoadChart(repo, repoURL, templateName, version string) (*chart.Chart, error) {
	helmCmd, err := helm.NewHelm("nonamespace", t.repoFile, t.repoCache)
	if err != nil {
		return nil, errors.WithMessage(err, "create a new helm command")
	}

	repoCmd := helm.NewRepo(t.repoFile, t.repoCache)
	if err := repoCmd.Add(repo, repoURL, "", ""); err != nil {
		return nil, errors.WithMessage(err, "add helm repo")
	}

	cp, err := helmCmd.Load(repo+"/"+templateName, version)
	if err != nil {
		return nil, err
	}

	return loader.Load(cp)
}
