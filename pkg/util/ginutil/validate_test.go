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

package ginutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppStoreName(t *testing.T) {
	tests := []struct {
		tname string
		name  string
		want  bool
	}{
		{
			tname: "33 characters",
			name:  "abcabcabcabcabcaabcabcabcabcabcax",
			want:  false,
		},
		{
			tname: "3 characters",
			name:  "abc",
			want:  false,
		},
		{
			tname: "4 characters",
			name:  "abcd",
			want:  true,
		},
		{
			tname: "32 characters",
			name:  "abcabcabcabcabcaabcabcabcabcabca",
			want:  true,
		},
		{
			tname: "start with number",
			name:  "1abc",
			want:  false,
		},
		{
			tname: "invalid character",
			name:  ".,';",
			want:  false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.tname, func(t *testing.T) {
			assert.Equal(t, tc.want, validateAppStoreName(tc.name))
		})
	}
}
