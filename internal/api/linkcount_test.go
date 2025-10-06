package api

import (
	"testing"

	"github.com/jamesainslie/dot/pkg/dot"
	"github.com/stretchr/testify/assert"
)

func TestCountLinksInPlan(t *testing.T) {
	source1 := dot.NewFilePath("/stow/vim/.vimrc").Unwrap()
	dest1 := dot.NewFilePath("/home/.vimrc").Unwrap()
	
	source2 := dot.NewFilePath("/stow/vim/.vim/vimrc").Unwrap()
	dest2 := dot.NewFilePath("/home/.vim/vimrc").Unwrap()
	
	source3 := dot.NewFilePath("/home/.bashrc").Unwrap()
	dest3 := dot.NewFilePath("/stow/bash/.bashrc").Unwrap()

	tests := []struct {
		name string
		plan dot.Plan
		want int
	}{
		{
			name: "empty plan",
			plan: dot.Plan{Operations: []dot.Operation{}},
			want: 0,
		},
		{
			name: "one link",
			plan: dot.Plan{
				Operations: []dot.Operation{
					dot.NewLinkCreate("l1", source1, dest1),
				},
			},
			want: 1,
		},
		{
			name: "multiple links",
			plan: dot.Plan{
				Operations: []dot.Operation{
					dot.NewLinkCreate("l1", source1, dest1),
					dot.NewLinkCreate("l2", source2, dest2),
				},
			},
			want: 2,
		},
		{
			name: "mixed operations",
			plan: dot.Plan{
				Operations: []dot.Operation{
					dot.NewLinkCreate("l1", source1, dest1),
					dot.NewDirCreate("d1", dest1),
					dot.NewLinkCreate("l2", source2, dest2),
					dot.NewFileMove("m1", source3, dest3),
				},
			},
			want: 2, // Only link creates
		},
		{
			name: "link deletes not counted",
			plan: dot.Plan{
				Operations: []dot.Operation{
					dot.NewLinkDelete("l1", dest1),
					dot.NewLinkDelete("l2", dest2),
				},
			},
			want: 0, // Link deletes don't count
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := countLinksInPlan(tt.plan)
			assert.Equal(t, tt.want, got)
		})
	}
}

