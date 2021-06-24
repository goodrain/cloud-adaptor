// RAINBOND, Application Management Platform
// Copyright (C) 2020-2021 Goodrain Co., Ltd.

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. For any non-GPL usage of Rainbond,
// one or multiple Commercial Licenses authorized by Goodrain Co., Ltd.
// must be obtained first.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

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
