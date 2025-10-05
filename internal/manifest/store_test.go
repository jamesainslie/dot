package manifest

import (
	"context"
	"testing"

	"github.com/jamesainslie/dot/pkg/dot"
)

func TestManifestStore_Interface(t *testing.T) {
	// Verify interface is implemented by mock
	var _ ManifestStore = (*mockManifestStore)(nil)
}

type mockManifestStore struct {
	loadFn func(context.Context, dot.TargetPath) dot.Result[Manifest]
	saveFn func(context.Context, dot.TargetPath, Manifest) error
}

func (m *mockManifestStore) Load(ctx context.Context, target dot.TargetPath) dot.Result[Manifest] {
	return m.loadFn(ctx, target)
}

func (m *mockManifestStore) Save(ctx context.Context, target dot.TargetPath, manifest Manifest) error {
	return m.saveFn(ctx, target, manifest)
}
