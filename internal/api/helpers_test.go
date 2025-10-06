package api

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsManifestNotFoundError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "os.ErrNotExist",
			err:  os.ErrNotExist,
			want: true,
		},
		{
			name: "wrapped os.ErrNotExist",
			err:  errors.Join(errors.New("prefix"), os.ErrNotExist),
			want: true,
		},
		{
			name: "other error",
			err:  errors.New("some other error"),
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isManifestNotFoundError(tt.err)
			assert.Equal(t, tt.want, got)
		})
	}
}
