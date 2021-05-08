package repo

import (
	"github.com/pkg/errors"
	"goodrain.com/cloud-adaptor/pkg/bcode"
	"testing"

	"github.com/stretchr/testify/assert"

	"goodrain.com/cloud-adaptor/cmd/cloud-adaptor/config"
	"goodrain.com/cloud-adaptor/internal/repo/appstore"
)

func TestGetTemplateVersion(t *testing.T) {
	cfg := &config.Config{
		Helm: &config.Helm{
			RepoFile:  "/tmp/helm/repositories.yaml",
			RepoCache: "/tmp/helm/cache",
		},
	}
	templateVersioner := appstore.NewTemplateVersioner(cfg)

	templateVersionRepo := NewTemplateVersionRepo(templateVersioner)

	tests := []struct {
		name    string
		version string
		err     error
	}{
		{
			name:    "template version not found",
			version: "foobar",
			err:     bcode.ErrTemplateVersionNotFound,
		},
		{
			name:    "template version not found-2",
			version: "9.3.70",
			err:     bcode.ErrTemplateVersionNotFound,
		},
		{
			name:    "ok",
			version: "9.3.7",
			err:     nil,
		},
	}

	for _, tc := range tests {
		tc := tc
		version, err := templateVersionRepo.GetTemplateVersion("rainbond", "https://openchart.goodrain.com/goodrain/rainbond", "mariadb", tc.version)
		if !assert.Equal(t, tc.err, errors.Cause(err)) {
			t.FailNow()
		}
		if tc.err != nil {
			continue
		}
		assert.NotEmpty(t, version.Readme)
		assert.NotEmpty(t, version.Questions)
		assert.NotEmpty(t, version.Values)
	}
}
