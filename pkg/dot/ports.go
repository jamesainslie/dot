package dot

import (
	"context"
	"io/fs"
)

// Filesystem Port

// FS defines the interface for filesystem operations.
// Implementations must handle context cancellation appropriately.
type FS interface {
	// Read operations
	Stat(ctx context.Context, name string) (FileInfo, error)
	ReadDir(ctx context.Context, name string) ([]DirEntry, error)
	ReadLink(ctx context.Context, name string) (string, error)
	ReadFile(ctx context.Context, name string) ([]byte, error)

	// Write operations
	WriteFile(ctx context.Context, name string, data []byte, perm fs.FileMode) error
	Mkdir(ctx context.Context, name string, perm fs.FileMode) error
	MkdirAll(ctx context.Context, name string, perm fs.FileMode) error
	Remove(ctx context.Context, name string) error
	RemoveAll(ctx context.Context, name string) error
	Symlink(ctx context.Context, oldname, newname string) error
	Rename(ctx context.Context, oldname, newname string) error

	// Query operations
	Exists(ctx context.Context, name string) bool
	IsDir(ctx context.Context, name string) (bool, error)
	IsSymlink(ctx context.Context, name string) (bool, error)
}

// FileInfo provides information about a file.
// Matches fs.FileInfo from standard library for compatibility.
type FileInfo interface {
	Name() string
	Size() int64
	Mode() fs.FileMode
	ModTime() any
	IsDir() bool
	Sys() any
}

// DirEntry provides information about a directory entry.
// Matches fs.DirEntry from standard library for compatibility.
type DirEntry interface {
	Name() string
	IsDir() bool
	Type() fs.FileMode
	Info() (FileInfo, error)
}

// Logger Port

// Logger defines the interface for structured logging.
// All methods accept context for correlation and structured key-value pairs.
type Logger interface {
	// Debug logs a debug-level message.
	Debug(ctx context.Context, msg string, args ...any)

	// Info logs an info-level message.
	Info(ctx context.Context, msg string, args ...any)

	// Warn logs a warning-level message.
	Warn(ctx context.Context, msg string, args ...any)

	// Error logs an error-level message.
	Error(ctx context.Context, msg string, args ...any)

	// With returns a new logger with additional context fields.
	With(args ...any) Logger
}

// Tracer Port

// Tracer defines the interface for distributed tracing.
type Tracer interface {
	// Start begins a new span with the given name and options.
	// Returns a new context with the span and the span itself.
	Start(ctx context.Context, name string, opts ...SpanOption) (context.Context, Span)
}

// Span represents a single unit of work in a trace.
type Span interface {
	// End completes the span.
	End()

	// RecordError records an error on the span.
	RecordError(err error)

	// SetAttributes adds attributes to the span.
	SetAttributes(attrs ...Attribute)
}

// SpanOption configures span creation.
type SpanOption interface {
	applySpanOption()
}

// Attribute represents a key-value attribute for spans.
type Attribute struct {
	Key   string
	Value any
}

// Metrics Port

// Metrics defines the interface for application metrics.
type Metrics interface {
	// Counter returns a counter metric.
	Counter(name string, labels ...string) Counter

	// Histogram returns a histogram metric.
	Histogram(name string, labels ...string) Histogram

	// Gauge returns a gauge metric.
	Gauge(name string, labels ...string) Gauge
}

// Counter represents a monotonically increasing counter.
type Counter interface {
	// Inc increments the counter by 1.
	Inc(labels ...string)

	// Add increments the counter by delta.
	Add(delta float64, labels ...string)
}

// Histogram represents a distribution of values.
type Histogram interface {
	// Observe records a value in the histogram.
	Observe(value float64, labels ...string)
}

// Gauge represents a value that can go up or down.
type Gauge interface {
	// Set sets the gauge to a specific value.
	Set(value float64, labels ...string)

	// Inc increments the gauge by 1.
	Inc(labels ...string)

	// Dec decrements the gauge by 1.
	Dec(labels ...string)
}
