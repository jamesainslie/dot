package dot_test

import (
	"testing"

	"github.com/jamesainslie/dot/pkg/dot"
	"github.com/stretchr/testify/assert"
)

func TestLinkCreateOperation(t *testing.T) {
	source := dot.NewFilePath("/home/user/.dotfiles/vim/vimrc").Unwrap()
	target := dot.NewFilePath("/home/user/.vimrc").Unwrap()
	
	op := dot.NewLinkCreate(source, target)
	
	assert.Equal(t, dot.OpKindLinkCreate, op.Kind())
	assert.Contains(t, op.String(), "vimrc")
	
	// Validate should check paths exist (not implemented yet, so should pass)
	err := op.Validate()
	assert.NoError(t, err)
	
	// Dependencies should be empty for link creation
	deps := op.Dependencies()
	assert.Empty(t, deps)
}

func TestLinkDeleteOperation(t *testing.T) {
	target := dot.NewFilePath("/home/user/.vimrc").Unwrap()
	
	op := dot.NewLinkDelete(target)
	
	assert.Equal(t, dot.OpKindLinkDelete, op.Kind())
	assert.Contains(t, op.String(), "vimrc")
	
	err := op.Validate()
	assert.NoError(t, err)
	
	deps := op.Dependencies()
	assert.Empty(t, deps)
}

func TestDirCreateOperation(t *testing.T) {
	path := dot.NewFilePath("/home/user/.vim").Unwrap()
	
	op := dot.NewDirCreate(path)
	
	assert.Equal(t, dot.OpKindDirCreate, op.Kind())
	assert.Contains(t, op.String(), ".vim")
	
	err := op.Validate()
	assert.NoError(t, err)
	
	deps := op.Dependencies()
	assert.Empty(t, deps)
}

func TestDirDeleteOperation(t *testing.T) {
	path := dot.NewFilePath("/home/user/.vim").Unwrap()
	
	op := dot.NewDirDelete(path)
	
	assert.Equal(t, dot.OpKindDirDelete, op.Kind())
	assert.Contains(t, op.String(), ".vim")
	
	err := op.Validate()
	assert.NoError(t, err)
	
	deps := op.Dependencies()
	assert.Empty(t, deps)
}

func TestFileMoveOperation(t *testing.T) {
	source := dot.NewFilePath("/home/user/.vimrc").Unwrap()
	dest := dot.NewFilePath("/home/user/.dotfiles/vim/vimrc").Unwrap()
	
	op := dot.NewFileMove(source, dest)
	
	assert.Equal(t, dot.OpKindFileMove, op.Kind())
	assert.Contains(t, op.String(), "vimrc")
	
	err := op.Validate()
	assert.NoError(t, err)
	
	deps := op.Dependencies()
	assert.Empty(t, deps)
}

func TestFileBackupOperation(t *testing.T) {
	source := dot.NewFilePath("/home/user/.vimrc").Unwrap()
	backup := dot.NewFilePath("/home/user/.vimrc.backup").Unwrap()
	
	op := dot.NewFileBackup(source, backup)
	
	assert.Equal(t, dot.OpKindFileBackup, op.Kind())
	assert.Contains(t, op.String(), "vimrc")
	
	err := op.Validate()
	assert.NoError(t, err)
	
	deps := op.Dependencies()
	assert.Empty(t, deps)
}

func TestOperationEquality(t *testing.T) {
	source := dot.NewFilePath("/home/user/.dotfiles/vim/vimrc").Unwrap()
	target := dot.NewFilePath("/home/user/.vimrc").Unwrap()
	
	op1 := dot.NewLinkCreate(source, target)
	op2 := dot.NewLinkCreate(source, target)
	op3 := dot.NewLinkDelete(target)
	
	assert.True(t, op1.Equals(op2))
	assert.False(t, op1.Equals(op3))
}

func TestOperationKindString(t *testing.T) {
	tests := []struct {
		kind dot.OperationKind
		want string
	}{
		{dot.OpKindLinkCreate, "LinkCreate"},
		{dot.OpKindLinkDelete, "LinkDelete"},
		{dot.OpKindDirCreate, "DirCreate"},
		{dot.OpKindDirDelete, "DirDelete"},
		{dot.OpKindFileMove, "FileMove"},
		{dot.OpKindFileBackup, "FileBackup"},
	}
	
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.kind.String())
		})
	}
}

