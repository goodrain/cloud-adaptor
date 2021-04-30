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
