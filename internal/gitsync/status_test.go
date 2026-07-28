package gitsync

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePorcelain(t *testing.T) {
	tests := []struct {
		name  string
		out   string
		want  []StatusEntry
		count int
	}{
		{
			name:  "empty output",
			out:   "",
			want:  []StatusEntry{},
			count: 0,
		},
		{
			name:  "whitespace only",
			out:   "\n\n",
			want:  []StatusEntry{},
			count: 0,
		},
		{
			name: "modified tracked file",
			out:  " M dot-vim/dot-vimrc\n",
			want: []StatusEntry{
				{Index: ' ', Worktree: 'M', Path: "dot-vim/dot-vimrc"},
			},
			count: 1,
		},
		{
			name: "untracked file",
			out:  "?? dot-tmux/dot-tmux.conf\n",
			want: []StatusEntry{
				{Index: '?', Worktree: '?', Path: "dot-tmux/dot-tmux.conf"},
			},
			count: 1,
		},
		{
			name: "staged addition and deletion",
			out:  "A  dot-zsh/dot-zshrc\n D dot-vim/colors/solarized.vim\n",
			want: []StatusEntry{
				{Index: 'A', Worktree: ' ', Path: "dot-zsh/dot-zshrc"},
				{Index: ' ', Worktree: 'D', Path: "dot-vim/colors/solarized.vim"},
			},
			count: 2,
		},
		{
			name: "rename records both paths",
			out:  "R  dot-vim/old.vim -> dot-vim/new.vim\n",
			want: []StatusEntry{
				{Index: 'R', Worktree: ' ', Path: "dot-vim/new.vim", OrigPath: "dot-vim/old.vim"},
			},
			count: 1,
		},
		{
			name: "quoted path is unquoted",
			out:  "?? \"dot-vim/spaced name.vim\"\n",
			want: []StatusEntry{
				{Index: '?', Worktree: '?', Path: "dot-vim/spaced name.vim"},
			},
			count: 1,
		},
		{
			name: "unmerged entry",
			out:  "UU dot-vim/dot-vimrc\n",
			want: []StatusEntry{
				{Index: 'U', Worktree: 'U', Path: "dot-vim/dot-vimrc"},
			},
			count: 1,
		},
		{
			name: "repository root file",
			out:  " M README.md\n",
			want: []StatusEntry{
				{Index: ' ', Worktree: 'M', Path: "README.md"},
			},
			count: 1,
		},
		{
			name:  "malformed short line is skipped",
			out:   "M\n",
			want:  []StatusEntry{},
			count: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParsePorcelain(tt.out)
			assert.Len(t, got, tt.count)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestStatusEntry_Code(t *testing.T) {
	entry := StatusEntry{Index: '?', Worktree: '?', Path: "a"}
	assert.Equal(t, "??", entry.Code())
}

func TestGroupByPackage(t *testing.T) {
	tests := []struct {
		name    string
		entries []StatusEntry
		want    []PackageGroup
	}{
		{
			name:    "no entries",
			entries: nil,
			want:    []PackageGroup{},
		},
		{
			name: "single package",
			entries: []StatusEntry{
				{Index: ' ', Worktree: 'M', Path: "dot-vim/dot-vimrc"},
			},
			want: []PackageGroup{
				{
					Package: "dot-vim",
					Entries: []StatusEntry{{Index: ' ', Worktree: 'M', Path: "dot-vim/dot-vimrc"}},
				},
			},
		},
		{
			name: "packages sorted, files sorted within package",
			entries: []StatusEntry{
				{Index: ' ', Worktree: 'M', Path: "dot-zsh/dot-zshrc"},
				{Index: ' ', Worktree: 'M', Path: "dot-vim/z.vim"},
				{Index: '?', Worktree: '?', Path: "dot-vim/a.vim"},
			},
			want: []PackageGroup{
				{
					Package: "dot-vim",
					Entries: []StatusEntry{
						{Index: '?', Worktree: '?', Path: "dot-vim/a.vim"},
						{Index: ' ', Worktree: 'M', Path: "dot-vim/z.vim"},
					},
				},
				{
					Package: "dot-zsh",
					Entries: []StatusEntry{{Index: ' ', Worktree: 'M', Path: "dot-zsh/dot-zshrc"}},
				},
			},
		},
		{
			name: "root files grouped separately and listed first",
			entries: []StatusEntry{
				{Index: ' ', Worktree: 'M', Path: "dot-vim/dot-vimrc"},
				{Index: ' ', Worktree: 'M', Path: "README.md"},
			},
			want: []PackageGroup{
				{
					Package: "",
					Entries: []StatusEntry{{Index: ' ', Worktree: 'M', Path: "README.md"}},
				},
				{
					Package: "dot-vim",
					Entries: []StatusEntry{{Index: ' ', Worktree: 'M', Path: "dot-vim/dot-vimrc"}},
				},
			},
		},
		{
			name: "untracked directory entry",
			entries: []StatusEntry{
				{Index: '?', Worktree: '?', Path: "dot-tmux/"},
			},
			want: []PackageGroup{
				{
					Package: "dot-tmux",
					Entries: []StatusEntry{{Index: '?', Worktree: '?', Path: "dot-tmux/"}},
				},
			},
		},
		{
			name: "rename groups by new path",
			entries: []StatusEntry{
				{Index: 'R', Worktree: ' ', Path: "dot-zsh/new", OrigPath: "dot-vim/old"},
			},
			want: []PackageGroup{
				{
					Package: "dot-zsh",
					Entries: []StatusEntry{{Index: 'R', Worktree: ' ', Path: "dot-zsh/new", OrigPath: "dot-vim/old"}},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GroupByPackage(tt.entries)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPackageGroup_Label(t *testing.T) {
	assert.Equal(t, "dot-vim", PackageGroup{Package: "dot-vim"}.Label())
	assert.Equal(t, RootLabel, PackageGroup{Package: ""}.Label())
}

func TestPackageLabels(t *testing.T) {
	groups := []PackageGroup{{Package: ""}, {Package: "dot-vim"}}
	assert.Equal(t, []string{RootLabel, "dot-vim"}, PackageLabels(groups))
}

func TestShortHostname(t *testing.T) {
	tests := []struct {
		name string
		host string
		want string
	}{
		{"fqdn", "zeus.int.example.com", "zeus"},
		{"short name", "zeus", "zeus"},
		{"empty", "", "unknown-host"},
		{"leading dot", ".local", "unknown-host"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ShortHostname(tt.host))
		})
	}
}

func TestCommitMessage(t *testing.T) {
	tests := []struct {
		name   string
		host   string
		groups []PackageGroup
		want   string
	}{
		{
			name:   "single package",
			host:   "zeus",
			groups: []PackageGroup{{Package: "dot-vim"}},
			want:   "sync: zeus local edits (dot-vim)",
		},
		{
			name:   "multiple packages comma separated",
			host:   "zeus.local",
			groups: []PackageGroup{{Package: "dot-vim"}, {Package: "dot-zsh"}},
			want:   "sync: zeus local edits (dot-vim, dot-zsh)",
		},
		{
			name:   "root files use root label",
			host:   "luggage",
			groups: []PackageGroup{{Package: ""}},
			want:   "sync: luggage local edits (" + RootLabel + ")",
		},
		{
			name:   "no groups omits package list",
			host:   "luggage",
			groups: nil,
			want:   "sync: luggage local edits",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, CommitMessage(tt.host, tt.groups))
		})
	}
}

func TestParsePorcelainThenGroup(t *testing.T) {
	out := " M dot-vim/dot-vimrc\n?? dot-zsh/dot-zshrc\n M README.md\n"

	groups := GroupByPackage(ParsePorcelain(out))
	require.Len(t, groups, 3)

	assert.Equal(t, RootLabel, groups[0].Label())
	assert.Equal(t, "dot-vim", groups[1].Package)
	assert.Equal(t, "dot-zsh", groups[2].Package)
}
