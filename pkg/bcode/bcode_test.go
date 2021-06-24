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

package bcode

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErr2Coder(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{
			name: "ok",
			err:  new(400, 10000001),
			want: 10000001,
		},
		{
			name: "200",
			err:  nil,
			want: 200,
		},
		{
			name: "not recognized",
			err:  fmt.Errorf("not recognized error"),
			want: 500,
		},
	}

	for idx := range tests {
		tc := tests[idx]
		t.Run(tc.name, func(t *testing.T) {
			coder := Err2Coder(tc.err)
			assert.Equal(t, tc.want, coder.Code())
		})
	}
}

func TestStr2Coder(t *testing.T) {
	tests := []struct {
		name string
		str  string
		want int
	}{
		{
			name: "ok",
			str:  "10001",
			want: 10001,
		},
		{
			name: "200",
			str:  "",
			want: 200,
		},
		{
			name: "not int",
			str:  "not int",
			want: 500,
		},
	}

	for idx := range tests {
		tc := tests[idx]
		t.Run(tc.name, func(t *testing.T) {
			coder := Str2Coder(tc.str)
			assert.Equal(t, tc.want, coder.Code())
		})
	}
}
