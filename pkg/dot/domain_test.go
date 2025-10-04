package dot_test

import (
	"testing"

	"github.com/jamesainslie/dot/pkg/dot"
	"github.com/stretchr/testify/assert"
)

func TestPackage(t *testing.T) {
	path := dot.NewPackagePath("/home/user/.dotfiles/vim").Unwrap()
	
	pkg := dot.Package{
		Name: "vim",
		Path: path,
	}
	
	assert.Equal(t, "vim", pkg.Name)
	assert.Equal(t, path, pkg.Path)
}

func TestNodeType(t *testing.T) {
	assert.Equal(t, "File", dot.NodeFile.String())
	assert.Equal(t, "Dir", dot.NodeDir.String())
	assert.Equal(t, "Symlink", dot.NodeSymlink.String())
}

func TestNode(t *testing.T) {
	path := dot.NewFilePath("/home/user/.dotfiles/vim/vimrc").Unwrap()
	
	node := dot.Node{
		Path:     path,
		Type:     dot.NodeFile,
		Children: nil,
	}
	
	assert.Equal(t, path, node.Path)
	assert.Equal(t, dot.NodeFile, node.Type)
	assert.Nil(t, node.Children)
	assert.True(t, node.IsFile())
	assert.False(t, node.IsDir())
	assert.False(t, node.IsSymlink())
}

func TestNodeDirectory(t *testing.T) {
	dirPath := dot.NewFilePath("/home/user/.dotfiles/vim").Unwrap()
	filePath := dot.NewFilePath("/home/user/.dotfiles/vim/vimrc").Unwrap()
	
	fileNode := dot.Node{
		Path: filePath,
		Type: dot.NodeFile,
	}
	
	dirNode := dot.Node{
		Path:     dirPath,
		Type:     dot.NodeDir,
		Children: []dot.Node{fileNode},
	}
	
	assert.True(t, dirNode.IsDir())
	assert.Len(t, dirNode.Children, 1)
	assert.Equal(t, fileNode, dirNode.Children[0])
}

func TestPlan(t *testing.T) {
	source := dot.NewFilePath("/home/user/.dotfiles/vim/vimrc").Unwrap()
	target := dot.NewFilePath("/home/user/.vimrc").Unwrap()
	
	op := dot.NewLinkCreate(source, target)
	
	plan := dot.Plan{
		Operations: []dot.Operation{op},
		Metadata: dot.PlanMetadata{
			PackageCount: 1,
		},
	}
	
	assert.Len(t, plan.Operations, 1)
	assert.Equal(t, 1, plan.Metadata.PackageCount)
}

func TestPlanValidation(t *testing.T) {
	t.Run("empty plan is valid", func(t *testing.T) {
		plan := dot.Plan{
			Operations: []dot.Operation{},
		}
		
		err := plan.Validate()
		assert.NoError(t, err)
	})
	
	t.Run("plan with operations is valid", func(t *testing.T) {
		source := dot.NewFilePath("/home/user/.dotfiles/vim/vimrc").Unwrap()
		target := dot.NewFilePath("/home/user/.vimrc").Unwrap()
		
		plan := dot.Plan{
			Operations: []dot.Operation{
				dot.NewLinkCreate(source, target),
			},
		}
		
		err := plan.Validate()
		assert.NoError(t, err)
	})
}

func TestPlanMetadata(t *testing.T) {
	metadata := dot.PlanMetadata{
		PackageCount:   3,
		OperationCount: 10,
		LinkCount:      7,
		DirCount:       3,
	}
	
	assert.Equal(t, 3, metadata.PackageCount)
	assert.Equal(t, 10, metadata.OperationCount)
	assert.Equal(t, 7, metadata.LinkCount)
	assert.Equal(t, 3, metadata.DirCount)
}

