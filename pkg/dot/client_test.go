package dot_test

import (
	"context"
	"testing"

	"github.com/jamesainslie/dot/pkg/dot"
	"github.com/stretchr/testify/require"
)

// mockClient implements the Client interface for testing
type mockClient struct {
	config dot.Config
}

func (m *mockClient) Manage(ctx context.Context, packages ...string) error {
	return nil
}

func (m *mockClient) PlanManage(ctx context.Context, packages ...string) (dot.Plan, error) {
	return dot.Plan{}, nil
}

func (m *mockClient) Unmanage(ctx context.Context, packages ...string) error {
	return nil
}

func (m *mockClient) PlanUnmanage(ctx context.Context, packages ...string) (dot.Plan, error) {
	return dot.Plan{}, nil
}

func (m *mockClient) Remanage(ctx context.Context, packages ...string) error {
	return nil
}

func (m *mockClient) PlanRemanage(ctx context.Context, packages ...string) (dot.Plan, error) {
	return dot.Plan{}, nil
}

func (m *mockClient) Adopt(ctx context.Context, files []string, pkg string) error {
	return nil
}

func (m *mockClient) PlanAdopt(ctx context.Context, files []string, pkg string) (dot.Plan, error) {
	return dot.Plan{}, nil
}

func (m *mockClient) Status(ctx context.Context, packages ...string) (dot.Status, error) {
	return dot.Status{}, nil
}

func (m *mockClient) List(ctx context.Context) ([]dot.PackageInfo, error) {
	return nil, nil
}

func (m *mockClient) Config() dot.Config {
	return m.config
}

// TestClientInterface verifies that mockClient implements Client
func TestClientInterface(t *testing.T) {
	var _ dot.Client = &mockClient{}
}

// Note: TestRegisterClientImpl will be added once internal/api is implemented
// and registered via init(). The registration mechanism is tested through
// the integration tests in internal/api/client_test.go.

