# Phase 18: Performance Optimization - Implementation Plan

## Overview

Phase 18 focuses on profiling and optimizing critical performance paths in the dot CLI. This phase applies measurement-driven optimization to ensure the tool scales efficiently for large package sets, complex directory trees, and concurrent operations.

## Design Principles

- **Measure First**: Profile before optimizing to identify actual bottlenecks
- **Data-Driven**: Use benchmarks and profiling data to guide optimization decisions
- **Regression Detection**: Establish performance baselines to prevent regressions
- **Streaming by Default**: Optimize for bounded memory usage with large datasets
- **Parallel Execution**: Leverage concurrency for independent operations
- **Cache Strategically**: Add caching only where profiling shows benefit

## Prerequisites

- Phase 17 complete: Integration testing provides realistic workloads for profiling
- Phase 16 complete: Property-based tests verify correctness during optimization
- Phase 13-14 complete: CLI commands available for end-to-end performance testing
- Phase 10 complete: Executor with parallel execution capabilities
- Phase 9 complete: Pipeline infrastructure for optimization

## Objectives

1. **Profile Critical Paths**: Identify performance bottlenecks through systematic profiling
2. **Optimize Hot Paths**: Improve performance of frequently-executed code
3. **Reduce Allocations**: Minimize memory allocations in hot paths
4. **Enhance Parallelism**: Tune concurrency parameters for optimal throughput
5. **Implement Caching**: Add strategic caching where profiling shows benefit
6. **Streaming Optimization**: Ensure bounded memory usage for large operations
7. **Benchmark Suite**: Create comprehensive benchmarks for regression detection

## Success Criteria

- CPU profiling identifies all hot paths consuming >5% of execution time
- Memory profiling identifies allocation hot spots
- Benchmark suite covers all critical operations
- Performance improvements documented with before/after metrics
- No functional regressions (all tests pass)
- Memory usage bounded for arbitrarily large package sets
- Parallel operations show linear speedup up to 8 cores
- 80% test coverage maintained throughout optimization

## Architecture Context

### Performance-Critical Components

From Architecture.md, the following components are performance-critical:

1. **Scanner**: Recursive directory traversal, tree building
2. **Planner**: Desired state computation, state diffing
3. **Resolver**: Conflict detection, resolution policy application
4. **Topological Sorter**: Graph construction, dependency analysis
5. **Executor**: Operation execution, parallel batch processing
6. **Ignore System**: Pattern matching, compiled pattern caching
7. **Manifest**: Content hashing, state serialization

### Streaming Architecture

The architecture emphasizes streaming APIs for memory efficiency:

```go
type OperationStream <-chan Result[Operation]

func PlanStowStream(ctx context.Context, input ScanInput, opts PipelineOpts) OperationStream
func CollectStream[T any](ctx context.Context, stream <-chan Result[T]) Result[[]T]
func StreamMap[A, B any](ctx context.Context, stream <-chan Result[A], f func(A) B) <-chan Result[B]
func StreamFilter[T any](ctx context.Context, stream <-chan Result[T], pred func(T) bool) <-chan Result[T]
```

### Parallelization Strategy

The dependency graph enables safe parallel execution:

```go
func (g *DependencyGraph) ParallelizationPlan() [][]Operation
func (e *Executor) executeParallel(ctx context.Context, plan Plan, checkpoint Checkpoint) ExecutionResult
func (e *Executor) executeBatch(ctx context.Context, batch []Operation, checkpoint Checkpoint) ExecutionResult
```

## Phase Breakdown

### Phase 18.1: Profiling Infrastructure

#### Objectives

- Set up CPU, memory, and goroutine profiling
- Create realistic benchmark scenarios
- Establish performance baselines
- Identify optimization targets

#### Tasks

##### 18.1.1: Profiling Support in CLI

**Test**: `cmd/dot/profile_test.go`

```go
func TestProfileFlagsCPU(t *testing.T) {
    // Test that --cpuprofile flag enables CPU profiling
    // Verify profile file is created with valid data
}

func TestProfileFlagsMemory(t *testing.T) {
    // Test that --memprofile flag enables memory profiling
    // Verify profile file is created with valid data
}

func TestProfileFlagsGoroutine(t *testing.T) {
    // Test that --goroutineprofile flag enables goroutine profiling
    // Verify profile file is created with valid data
}
```

**Implementation**: Add profiling flags to root command

```go
// cmd/dot/root.go
var (
    cpuProfile       string
    memProfile       string
    goroutineProfile string
)

func init() {
    rootCmd.PersistentFlags().StringVar(&cpuProfile, "cpuprofile", "", "Write CPU profile to file")
    rootCmd.PersistentFlags().StringVar(&memProfile, "memprofile", "", "Write memory profile to file")
    rootCmd.PersistentFlags().StringVar(&goroutineProfile, "goroutineprofile", "", "Write goroutine profile to file")
}

func Execute() error {
    if cpuProfile != "" {
        f, err := os.Create(cpuProfile)
        if err != nil {
            return fmt.Errorf("create CPU profile: %w", err)
        }
        defer f.Close()
        
        if err := pprof.StartCPUProfile(f); err != nil {
            return fmt.Errorf("start CPU profile: %w", err)
        }
        defer pprof.StopCPUProfile()
    }
    
    if err := rootCmd.Execute(); err != nil {
        return err
    }
    
    if memProfile != "" {
        f, err := os.Create(memProfile)
        if err != nil {
            return fmt.Errorf("create memory profile: %w", err)
        }
        defer f.Close()
        
        runtime.GC()
        if err := pprof.WriteHeapProfile(f); err != nil {
            return fmt.Errorf("write memory profile: %w", err)
        }
    }
    
    if goroutineProfile != "" {
        f, err := os.Create(goroutineProfile)
        if err != nil {
            return fmt.Errorf("create goroutine profile: %w", err)
        }
        defer f.Close()
        
        if err := pprof.Lookup("goroutine").WriteTo(f, 0); err != nil {
            return fmt.Errorf("write goroutine profile: %w", err)
        }
    }
    
    return nil
}
```

**Commit**: `feat(cli): add profiling flags for performance analysis`

##### 18.1.2: Benchmark Infrastructure

**Test**: `internal/benchmark/benchmark_test.go`

```go
func BenchmarkScanPackage(b *testing.B) {
    // Benchmark scanning single package with various sizes
    sizes := []int{10, 100, 1000, 10000}
    for _, size := range sizes {
        b.Run(fmt.Sprintf("files_%d", size), func(b *testing.B) {
            fs := setupFixturePackage(size)
            b.ResetTimer()
            for i := 0; i < b.N; i++ {
                scanPackage(context.Background(), fs, stowPath, "pkg", ignoreSet)
            }
        })
    }
}

func BenchmarkScanParallel(b *testing.B) {
    // Benchmark parallel package scanning
    // Test with 1, 2, 4, 8 packages
}

func BenchmarkPlanDesiredState(b *testing.B) {
    // Benchmark desired state computation
    // Test with various package sizes
}

func BenchmarkDiffStates(b *testing.B) {
    // Benchmark state diffing
    // Test with various numbers of operations
}

func BenchmarkTopologicalSort(b *testing.B) {
    // Benchmark graph sorting
    // Test with various dependency graph sizes
}

func BenchmarkExecuteSequential(b *testing.B) {
    // Benchmark sequential operation execution
    // Test with various operation counts
}

func BenchmarkExecuteParallel(b *testing.B) {
    // Benchmark parallel operation execution
    // Test with various batch sizes
}

func BenchmarkIgnorePatternMatch(b *testing.B) {
    // Benchmark pattern matching
    // Test with various pattern counts and path lengths
}

func BenchmarkContentHash(b *testing.B) {
    // Benchmark content hashing
    // Test with various file sizes
}
```

**Implementation**: Create benchmark package with test data generators

```go
// internal/benchmark/fixtures.go
func setupFixturePackage(fileCount int) FS {
    fs := memfs.New()
    // Create package with fileCount files
    // Use realistic directory structure
    // Include nested directories
    return fs
}

func setupFixtureMultiPackage(packageCount, filesPerPackage int) FS {
    // Create multiple packages for parallel benchmarks
}

func setupFixtureDependencyGraph(nodeCount, avgDegree int) []Operation {
    // Create operation graph with specified topology
}
```

**Commit**: `test(benchmark): add benchmark infrastructure and fixtures`

##### 18.1.3: Baseline Profiling

**Task**: Run benchmarks and collect baseline profiles

**Script**: `scripts/profile.sh`

```bash
#!/bin/bash
# Profile baseline performance

set -e

PROFILE_DIR="profiles/baseline"
mkdir -p "$PROFILE_DIR"

# CPU profile: scan large package set
go run ./cmd/dot manage \
  --cpuprofile="$PROFILE_DIR/cpu_scan.prof" \
  --dir=./test/fixtures/large-packages \
  --target=/tmp/profile-target \
  --dry-run \
  pkg1 pkg2 pkg3 pkg4 pkg5

# Memory profile: scan large package set
go run ./cmd/dot manage \
  --memprofile="$PROFILE_DIR/mem_scan.prof" \
  --dir=./test/fixtures/large-packages \
  --target=/tmp/profile-target \
  --dry-run \
  pkg1 pkg2 pkg3 pkg4 pkg5

# Goroutine profile: parallel execution
go run ./cmd/dot manage \
  --goroutineprofile="$PROFILE_DIR/goroutine_exec.prof" \
  --dir=./test/fixtures/large-packages \
  --target=/tmp/profile-target \
  pkg1 pkg2 pkg3 pkg4 pkg5

# Run benchmarks
go test -bench=. -benchmem -cpuprofile="$PROFILE_DIR/bench_cpu.prof" \
  -memprofile="$PROFILE_DIR/bench_mem.prof" \
  ./internal/benchmark/... | tee "$PROFILE_DIR/baseline_bench.txt"

echo "Baseline profiles collected in $PROFILE_DIR"
```

**Analysis**: `scripts/analyze_profiles.sh`

```bash
#!/bin/bash
# Analyze collected profiles

PROFILE_DIR="$1"

echo "=== Top 20 CPU Consumers ==="
go tool pprof -top20 "$PROFILE_DIR/cpu_scan.prof"

echo ""
echo "=== Top 20 Memory Allocators ==="
go tool pprof -top20 "$PROFILE_DIR/mem_scan.prof"

echo ""
echo "=== Allocation Sites ==="
go tool pprof -alloc_space -top20 "$PROFILE_DIR/mem_scan.prof"

echo ""
echo "=== Goroutine Counts ==="
go tool pprof -top10 "$PROFILE_DIR/goroutine_exec.prof"
```

**Commit**: `chore(perf): add profiling and analysis scripts`

##### 18.1.4: Performance Test Fixtures

**Test**: `test/fixtures/performance/README.md`

Document fixture structure:

```markdown
# Performance Test Fixtures

## small-packages/
- 5 packages with 10-20 files each
- Use for fast iteration during development

## medium-packages/
- 10 packages with 100-200 files each
- Use for realistic performance testing

## large-packages/
- 20 packages with 500-1000 files each
- Use for stress testing and scalability verification

## deep-nesting/
- Packages with deeply nested directory structures (20+ levels)
- Use for testing tree traversal performance

## wide-directories/
- Packages with directories containing many files (1000+ per directory)
- Use for testing readdir performance

## many-symlinks/
- Large number of existing symlinks in target
- Use for testing conflict detection performance
```

**Task**: Generate performance fixtures

```go
// test/fixtures/performance/generate.go
package main

import (
    "flag"
    "fmt"
    "os"
    "path/filepath"
)

func main() {
    fixtureType := flag.String("type", "medium", "Fixture type: small, medium, large, deep, wide, symlinks")
    flag.Parse()
    
    switch *fixtureType {
    case "small":
        generateSmallPackages()
    case "medium":
        generateMediumPackages()
    case "large":
        generateLargePackages()
    case "deep":
        generateDeepNesting()
    case "wide":
        generateWideDirectories()
    case "symlinks":
        generateManySymlinks()
    default:
        fmt.Fprintf(os.Stderr, "Unknown fixture type: %s\n", *fixtureType)
        os.Exit(1)
    }
}

func generateMediumPackages() {
    // Create 10 packages with 100-200 files each
    for i := 1; i <= 10; i++ {
        pkgDir := filepath.Join("medium-packages", fmt.Sprintf("pkg%d", i))
        fileCount := 100 + i*10
        generatePackage(pkgDir, fileCount, 5)
    }
}

func generatePackage(dir string, fileCount, maxDepth int) {
    // Create package with specified file count and nesting
    // Distribute files across nested directories
    // Include mix of dotfiles and regular files
}
```

**Commit**: `test(fixtures): add performance test fixtures and generator`

**Deliverable 18.1**: Profiling infrastructure with baseline measurements

---

### Phase 18.2: Hot Path Optimization

#### Objectives

- Optimize functions identified in profiling as hot paths
- Reduce memory allocations in critical sections
- Improve algorithmic complexity where possible
- Maintain functional correctness throughout

#### Tasks

##### 18.2.1: Scanner Optimization

**Analysis**: Profile scanner to identify hot paths

Common optimization opportunities:
- Path allocation reduction through pooling
- Minimize string concatenation in path operations
- Batch filesystem operations (readdir)
- Parallelize independent tree walks

**Test**: `internal/scanner/scanner_bench_test.go`

```go
func BenchmarkScanTreeSequential(b *testing.B) {
    fs := setupLargeTree(1000)
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        scanTree(context.Background(), fs, rootPath, ignoreSet)
    }
}

func BenchmarkScanTreeParallel(b *testing.B) {
    fs := setupLargeTree(1000)
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        scanTreeParallel(context.Background(), fs, rootPath, ignoreSet)
    }
}

func BenchmarkBuildNode(b *testing.B) {
    // Benchmark node construction
    // Test allocation reduction strategies
}
```

**Optimization**: Reduce allocations in tree building

```go
// internal/scanner/tree.go

// Before: Creates new nodes with individual allocations
func buildNode(ctx context.Context, fs FS, path Path, ignore IgnoreSet) (*Node, error) {
    info, err := fs.Stat(ctx, path.String())
    if err != nil {
        return nil, err
    }
    
    node := &Node{
        Name: filepath.Base(path.String()),
        Path: path,
        Type: nodeTypeFromInfo(info),
        Info: info,
    }
    
    if info.IsDir() {
        entries, err := fs.ReadDir(ctx, path.String())
        if err != nil {
            return nil, err
        }
        
        node.Children = make([]*Node, 0, len(entries))
        for _, entry := range entries {
            childPath := path.Join(entry.Name())
            if ignore.Match(childPath.String()) {
                continue
            }
            
            child, err := buildNode(ctx, fs, childPath, ignore)
            if err != nil {
                return nil, err
            }
            node.Children = append(node.Children, child)
        }
    }
    
    return node, nil
}

// After: Preallocate and reuse buffers
type nodeBuilder struct {
    pathBuf   []byte
    entryBuf  []DirEntry
}

func newNodeBuilder() *nodeBuilder {
    return &nodeBuilder{
        pathBuf:  make([]byte, 0, 4096),
        entryBuf: make([]DirEntry, 0, 256),
    }
}

func (b *nodeBuilder) buildNode(ctx context.Context, fs FS, path Path, ignore IgnoreSet) (*Node, error) {
    info, err := fs.Stat(ctx, path.String())
    if err != nil {
        return nil, err
    }
    
    node := &Node{
        Name: filepath.Base(path.String()),
        Path: path,
        Type: nodeTypeFromInfo(info),
        Info: info,
    }
    
    if info.IsDir() {
        // Reuse entry buffer
        b.entryBuf = b.entryBuf[:0]
        entries, err := fs.ReadDir(ctx, path.String())
        if err != nil {
            return nil, err
        }
        
        // Preallocate children slice
        node.Children = make([]*Node, 0, len(entries))
        
        for _, entry := range entries {
            // Reuse path buffer for string building
            childPathStr := b.buildChildPath(path.String(), entry.Name())
            
            if ignore.Match(childPathStr) {
                continue
            }
            
            childPath, _ := NewFilePath(childPathStr)
            child, err := b.buildNode(ctx, fs, childPath, ignore)
            if err != nil {
                return nil, err
            }
            node.Children = append(node.Children, child)
        }
    }
    
    return node, nil
}

func (b *nodeBuilder) buildChildPath(parent, name string) string {
    b.pathBuf = b.pathBuf[:0]
    b.pathBuf = append(b.pathBuf, parent...)
    b.pathBuf = append(b.pathBuf, filepath.Separator)
    b.pathBuf = append(b.pathBuf, name...)
    return string(b.pathBuf)
}
```

**Commit**: `perf(scanner): reduce allocations in tree building`

##### 18.2.2: Planner Optimization

**Analysis**: Profile planner to identify hot paths

Common optimization opportunities:
- Map preallocation with size hints
- Avoid redundant path operations
- Optimize state diffing algorithm
- Cache computed values

**Test**: `internal/planner/planner_bench_test.go`

```go
func BenchmarkComputeDesiredState(b *testing.B) {
    packages := setupPackages(10, 100)
    opts := PlanOpts{Folding: true}
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        computeDesiredState(packages, opts)
    }
}

func BenchmarkDiffStates(b *testing.B) {
    desired := setupDesiredState(1000)
    current := setupCurrentState(500)
    opts := PlanOpts{}
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        diffStates(desired, current, opts)
    }
}
```

**Optimization**: Preallocate maps with size hints

```go
// internal/planner/desired.go

// Before: Maps grow dynamically
func computeDesiredState(packages []Package, opts PlanOpts) DesiredState {
    desired := DesiredState{
        Links: make(map[TargetPath]LinkSpec),
        Dirs:  make(map[TargetPath]DirSpec),
    }
    
    // Build state...
    
    return desired
}

// After: Preallocate based on estimated size
func computeDesiredState(packages []Package, opts PlanOpts) DesiredState {
    // Estimate total file count across packages
    estimatedFiles := 0
    estimatedDirs := 0
    for _, pkg := range packages {
        stats := pkg.Files.Stats()
        estimatedFiles += stats.FileCount
        estimatedDirs += stats.DirCount
    }
    
    desired := DesiredState{
        Links: make(map[TargetPath]LinkSpec, estimatedFiles),
        Dirs:  make(map[TargetPath]DirSpec, estimatedDirs),
    }
    
    // Build state...
    
    return desired
}

// Add stats method to FileTree
func (t FileTree) Stats() TreeStats {
    return t.Root.Stats()
}

func (n *Node) Stats() TreeStats {
    stats := TreeStats{}
    n.Walk(func(node *Node) {
        switch node.Type {
        case NodeFile:
            stats.FileCount++
        case NodeDir:
            stats.DirCount++
        case NodeSymlink:
            stats.SymlinkCount++
        }
    })
    return stats
}

type TreeStats struct {
    FileCount    int
    DirCount     int
    SymlinkCount int
}
```

**Commit**: `perf(planner): preallocate maps with size hints`

##### 18.2.3: Ignore Pattern Optimization

**Analysis**: Profile pattern matching

Common optimization opportunities:
- Compile patterns once and cache
- Use sync.Map for concurrent access
- Short-circuit evaluation for common patterns
- Optimize regex for common cases

**Test**: `internal/ignore/pattern_bench_test.go`

```go
func BenchmarkPatternMatch(b *testing.B) {
    patterns := []string{"*.log", "*.tmp", ".git/**", "node_modules/**"}
    set := compilePatterns(patterns)
    path := "/home/user/.config/nvim/plugin/settings.vim"
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        set.Match(path)
    }
}

func BenchmarkPatternMatchCached(b *testing.B) {
    // Benchmark with compiled pattern cache
}

func BenchmarkPatternMatchParallel(b *testing.B) {
    // Benchmark concurrent pattern matching
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            set.Match(path)
        }
    })
}
```

**Optimization**: Implement pattern cache with sync.Map

```go
// internal/ignore/cache.go

type PatternCache struct {
    compiled sync.Map // map[string]*regexp.Regexp
    hits     atomic.Int64
    misses   atomic.Int64
}

func NewPatternCache() *PatternCache {
    return &PatternCache{}
}

func (c *PatternCache) GetOrCompile(pattern string) (*regexp.Regexp, error) {
    // Try cache first
    if cached, ok := c.compiled.Load(pattern); ok {
        c.hits.Add(1)
        return cached.(*regexp.Regexp), nil
    }
    
    // Compile pattern
    c.misses.Add(1)
    re, err := compilePattern(pattern)
    if err != nil {
        return nil, err
    }
    
    // Store in cache
    c.compiled.Store(pattern, re)
    return re, nil
}

func (c *PatternCache) Stats() CacheStats {
    return CacheStats{
        Hits:   c.hits.Load(),
        Misses: c.misses.Load(),
    }
}

type CacheStats struct {
    Hits   int64
    Misses int64
}

// Update IgnoreSet to use cache
type IgnoreSet struct {
    patterns []string
    cache    *PatternCache
}

func NewIgnoreSet(patterns []string) *IgnoreSet {
    return &IgnoreSet{
        patterns: patterns,
        cache:    NewPatternCache(),
    }
}

func (s *IgnoreSet) Match(path string) bool {
    for _, pattern := range s.patterns {
        re, err := s.cache.GetOrCompile(pattern)
        if err != nil {
            continue
        }
        if re.MatchString(path) {
            return true
        }
    }
    return false
}
```

**Commit**: `perf(ignore): implement pattern cache with sync.Map`

##### 18.2.4: Executor Optimization

**Analysis**: Profile executor

Common optimization opportunities:
- Tune worker pool size
- Optimize channel buffer sizes
- Reduce context switching overhead
- Batch filesystem operations

**Test**: `internal/executor/executor_bench_test.go`

```go
func BenchmarkExecuteSequential(b *testing.B) {
    plan := setupPlan(100)
    fs := memfs.New()
    exec := NewExecutor(ExecutorOpts{FS: fs})
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        exec.executeSequential(context.Background(), plan, Checkpoint{})
    }
}

func BenchmarkExecuteParallel(b *testing.B) {
    plan := setupPlan(100)
    fs := memfs.New()
    exec := NewExecutor(ExecutorOpts{FS: fs})
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        exec.executeParallel(context.Background(), plan, Checkpoint{})
    }
}

func BenchmarkExecuteBatch(b *testing.B) {
    // Test different batch sizes
    sizes := []int{10, 50, 100, 500}
    for _, size := range sizes {
        b.Run(fmt.Sprintf("batch_%d", size), func(b *testing.B) {
            batch := setupBatch(size)
            fs := memfs.New()
            exec := NewExecutor(ExecutorOpts{FS: fs})
            checkpoint := Checkpoint{}
            
            b.ResetTimer()
            for i := 0; i < b.N; i++ {
                exec.executeBatch(context.Background(), batch, checkpoint)
            }
        })
    }
}
```

**Optimization**: Tune concurrency parameters

```go
// internal/executor/parallel.go

// Add concurrency configuration
type ParallelConfig struct {
    MaxWorkers     int
    ChannelBuffer  int
    BatchSizeHint  int
}

func DefaultParallelConfig() ParallelConfig {
    return ParallelConfig{
        MaxWorkers:    runtime.NumCPU(),
        ChannelBuffer: 100,
        BatchSizeHint: 50,
    }
}

// Optimize batch execution with semaphore
func (e *Executor) executeBatch(ctx context.Context, batch []Operation, checkpoint Checkpoint) ExecutionResult {
    result := ExecutionResult{}
    var mu sync.Mutex
    
    // Use semaphore to limit concurrency
    sem := make(chan struct{}, e.config.MaxWorkers)
    var wg sync.WaitGroup
    
    for _, op := range batch {
        wg.Add(1)
        
        // Acquire semaphore
        sem <- struct{}{}
        
        go func(operation Operation) {
            defer wg.Done()
            defer func() { <-sem }() // Release semaphore
            
            opID := operation.(interface{ ID() OperationID }).ID()
            
            if err := operation.Execute(ctx, e.fs); err != nil {
                mu.Lock()
                result.Failed = append(result.Failed, opID)
                result.Errors = append(result.Errors, err)
                mu.Unlock()
                return
            }
            
            mu.Lock()
            result.Executed = append(result.Executed, opID)
            checkpoint.Record(opID, operation)
            mu.Unlock()
        }(op)
    }
    
    wg.Wait()
    return result
}
```

**Commit**: `perf(executor): tune concurrency parameters and add semaphore`

**Deliverable 18.2**: Optimized hot paths with measurable improvements

---

### Phase 18.3: Streaming Optimization

#### Objectives

- Ensure bounded memory usage for large operations
- Optimize channel buffer sizes
- Implement backpressure handling
- Profile streaming performance

#### Tasks

##### 18.3.1: Stream Buffer Optimization

**Test**: `internal/pipeline/streaming_bench_test.go`

```go
func BenchmarkOperationStream(b *testing.B) {
    // Test different buffer sizes
    bufferSizes := []int{0, 10, 50, 100, 500, 1000}
    for _, size := range bufferSizes {
        b.Run(fmt.Sprintf("buffer_%d", size), func(b *testing.B) {
            input := setupScanInput(1000)
            opts := PipelineOpts{StreamBuffer: size}
            
            b.ResetTimer()
            for i := 0; i < b.N; i++ {
                stream := PlanStowStream(context.Background(), input, opts)
                for range stream {
                    // Consume stream
                }
            }
        })
    }
}

func BenchmarkStreamCollect(b *testing.B) {
    // Benchmark stream collection
    stream := generateOperationStream(1000)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        CollectStream(context.Background(), stream)
    }
}

func BenchmarkStreamMap(b *testing.B) {
    // Benchmark stream transformation
}

func BenchmarkStreamFilter(b *testing.B) {
    // Benchmark stream filtering
}
```

**Optimization**: Implement adaptive buffer sizing

```go
// internal/pipeline/streaming.go

type StreamConfig struct {
    InitialBuffer int
    MaxBuffer     int
    MinBuffer     int
}

func DefaultStreamConfig() StreamConfig {
    return StreamConfig{
        InitialBuffer: 100,
        MaxBuffer:     1000,
        MinBuffer:     10,
    }
}

func PlanStowStream(ctx context.Context, input ScanInput, opts PipelineOpts) OperationStream {
    ch := make(chan Result[Operation], opts.StreamConfig.InitialBuffer)
    
    go func() {
        defer close(ch)
        
        // Stream package scanning
        for pkg := range scanPackagesStream(ctx, input) {
            if !pkg.IsOk() {
                select {
                case ch <- Err[Operation](pkg.err):
                case <-ctx.Done():
                    return
                }
                continue
            }
            
            // Stream operations for this package
            for op := range planPackageStream(ctx, pkg.value, opts) {
                select {
                case ch <- op:
                case <-ctx.Done():
                    return
                }
            }
        }
    }()
    
    return ch
}
```

**Commit**: `perf(pipeline): optimize stream buffer sizing`

##### 18.3.2: Backpressure Implementation

**Test**: `internal/pipeline/backpressure_test.go`

```go
func TestBackpressureSlowConsumer(t *testing.T) {
    // Test that fast producer doesn't overwhelm slow consumer
    // Verify bounded memory usage
}

func TestBackpressureCancellation(t *testing.T) {
    // Test that cancellation propagates through stream
    // Verify no goroutine leaks
}

func TestBackpressureMultiStage(t *testing.T) {
    // Test backpressure through multi-stage pipeline
}
```

**Implementation**: Add backpressure-aware stream operators

```go
// internal/pipeline/streaming.go

// StreamMapWithBackpressure applies function with backpressure control
func StreamMapWithBackpressure[A, B any](
    ctx context.Context,
    stream <-chan Result[A],
    f func(A) B,
    bufferSize int,
) <-chan Result[B] {
    out := make(chan Result[B], bufferSize)
    
    go func() {
        defer close(out)
        
        for result := range stream {
            if !result.IsOk() {
                select {
                case out <- Err[B](result.err):
                case <-ctx.Done():
                    return
                }
                continue
            }
            
            mapped := f(result.value)
            
            select {
            case out <- Ok(mapped):
            case <-ctx.Done():
                return
            }
        }
    }()
    
    return out
}

// StreamBatch groups stream items into batches
func StreamBatch[T any](
    ctx context.Context,
    stream <-chan Result[T],
    batchSize int,
    timeout time.Duration,
) <-chan Result[[]T] {
    out := make(chan Result[[]T])
    
    go func() {
        defer close(out)
        
        batch := make([]T, 0, batchSize)
        timer := time.NewTimer(timeout)
        defer timer.Stop()
        
        for {
            select {
            case result, ok := <-stream:
                if !ok {
                    // Stream closed, emit final batch
                    if len(batch) > 0 {
                        out <- Ok(batch)
                    }
                    return
                }
                
                if !result.IsOk() {
                    out <- Err[[]T](result.err)
                    continue
                }
                
                batch = append(batch, result.value)
                
                if len(batch) >= batchSize {
                    out <- Ok(batch)
                    batch = make([]T, 0, batchSize)
                    timer.Reset(timeout)
                }
                
            case <-timer.C:
                // Timeout, emit partial batch
                if len(batch) > 0 {
                    out <- Ok(batch)
                    batch = make([]T, 0, batchSize)
                }
                timer.Reset(timeout)
                
            case <-ctx.Done():
                return
            }
        }
    }()
    
    return out
}
```

**Commit**: `feat(pipeline): add backpressure-aware stream operators`

##### 18.3.3: Memory Profiling for Streams

**Test**: Verify bounded memory usage

```go
func TestStreamMemoryBounded(t *testing.T) {
    // Create large stream (10000 operations)
    input := setupLargeScanInput(100, 100) // 100 packages, 100 files each
    
    // Capture initial memory
    var m1 runtime.MemStats
    runtime.GC()
    runtime.ReadMemStats(&m1)
    
    // Process stream
    stream := PlanStowStream(context.Background(), input, PipelineOpts{})
    count := 0
    for range stream {
        count++
        
        // Capture memory periodically
        if count%1000 == 0 {
            var m2 runtime.MemStats
            runtime.ReadMemStats(&m2)
            
            // Memory growth should be bounded
            growth := m2.Alloc - m1.Alloc
            maxGrowth := uint64(100 * 1024 * 1024) // 100MB
            
            if growth > maxGrowth {
                t.Errorf("Memory growth too large: %d bytes (max %d)", growth, maxGrowth)
            }
        }
    }
}
```

**Benchmark**: Profile memory usage

```bash
# Run with memory profiling
go test -bench=BenchmarkOperationStream -memprofile=mem.prof ./internal/pipeline/

# Analyze allocations
go tool pprof -alloc_space mem.prof
```

**Commit**: `test(pipeline): verify bounded memory usage for streams`

##### 18.3.4: Streaming Scanner Implementation

**Test**: `internal/scanner/streaming_test.go`

```go
func TestScanPackagesStream(t *testing.T) {
    // Test streaming package scan
    // Verify packages emitted as discovered
}

func TestScanPackagesStreamError(t *testing.T) {
    // Test error handling in stream
    // Verify error propagation
}

func TestScanPackagesStreamCancellation(t *testing.T) {
    // Test cancellation during scan
    // Verify clean shutdown
}
```

**Implementation**: Streaming package scanner

```go
// internal/scanner/streaming.go

func scanPackagesStream(ctx context.Context, input ScanInput) <-chan Result[Package] {
    ch := make(chan Result[Package], input.StreamConfig.InitialBuffer)
    
    go func() {
        defer close(ch)
        
        // Limit concurrent scans
        sem := make(chan struct{}, runtime.NumCPU())
        var wg sync.WaitGroup
        
        for _, pkgName := range input.Packages {
            wg.Add(1)
            
            // Acquire semaphore
            select {
            case sem <- struct{}{}:
            case <-ctx.Done():
                wg.Done()
                return
            }
            
            go func(name string) {
                defer wg.Done()
                defer func() { <-sem }()
                
                pkg := scanPackage(ctx, input.FS, input.PackageDir, name, input.Ignore)
                
                select {
                case ch <- pkg:
                case <-ctx.Done():
                    return
                }
            }(pkgName)
        }
        
        wg.Wait()
    }()
    
    return ch
}

func planPackageStream(ctx context.Context, pkg Package, opts PipelineOpts) <-chan Result[Operation] {
    ch := make(chan Result[Operation], opts.StreamConfig.InitialBuffer)
    
    go func() {
        defer close(ch)
        
        // Compute desired state for this package
        desired := computePackageDesiredState(pkg, opts)
        
        // Stream operations
        for target, spec := range desired.Links {
            op := LinkCreate{
                ID:     genID(),
                Source: spec.Source,
                Target: spec.Target,
                Mode:   spec.Mode,
            }
            
            select {
            case ch <- Ok[Operation](op):
            case <-ctx.Done():
                return
            }
        }
    }()
    
    return ch
}
```

**Commit**: `feat(scanner): implement streaming package scanner`

**Deliverable 18.3**: Streaming operations with bounded memory usage

---

### Phase 18.4: Caching Strategy

#### Objectives

- Implement strategic caching based on profiling results
- Add LRU eviction for bounded memory usage
- Profile cache effectiveness
- Measure hit rates

#### Tasks

##### 18.4.1: Path Resolution Cache

**Test**: `internal/domain/path_cache_test.go`

```go
func TestPathCache(t *testing.T) {
    cache := NewPathCache(100)
    
    path1, _ := cache.GetOrCreate("/home/user/.bashrc")
    path2, _ := cache.GetOrCreate("/home/user/.bashrc")
    
    // Should return same instance
    if path1 != path2 {
        t.Error("Path cache not returning same instance")
    }
}

func TestPathCacheEviction(t *testing.T) {
    cache := NewPathCache(10)
    
    // Add 20 paths (exceeds capacity)
    for i := 0; i < 20; i++ {
        cache.GetOrCreate(fmt.Sprintf("/path/%d", i))
    }
    
    // Cache should have evicted oldest entries
    stats := cache.Stats()
    if stats.Size > 10 {
        t.Errorf("Cache size %d exceeds capacity 10", stats.Size)
    }
}

func TestPathCacheConcurrent(t *testing.T) {
    cache := NewPathCache(100)
    
    // Concurrent access
    var wg sync.WaitGroup
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for j := 0; j < 100; j++ {
                cache.GetOrCreate(fmt.Sprintf("/path/%d", j%50))
            }
        }()
    }
    wg.Wait()
    
    // Verify stats
    stats := cache.Stats()
    if stats.Hits == 0 {
        t.Error("No cache hits in concurrent test")
    }
}
```

**Implementation**: LRU path cache

```go
// internal/domain/path_cache.go

import "github.com/hashicorp/golang-lru/v2/expirable"

type PathCache struct {
    cache *expirable.LRU[string, interface{}]
    hits  atomic.Int64
    misses atomic.Int64
}

func NewPathCache(size int) *PathCache {
    cache := expirable.NewLRU[string, interface{}](size, nil, 0)
    return &PathCache{
        cache: cache,
    }
}

func (c *PathCache) GetOrCreate(pathStr string) (interface{}, error) {
    // Try cache
    if cached, ok := c.cache.Get(pathStr); ok {
        c.hits.Add(1)
        return cached, nil
    }
    
    // Create and validate
    c.misses.Add(1)
    
    // Determine path type and create
    var path interface{}
    var err error
    
    // Simple heuristic: check if path looks absolute
    if filepath.IsAbs(pathStr) {
        path, err = NewFilePath(pathStr)
    } else {
        return nil, fmt.Errorf("relative path not supported in cache")
    }
    
    if err != nil {
        return nil, err
    }
    
    // Store in cache
    c.cache.Add(pathStr, path)
    return path, nil
}

func (c *PathCache) Stats() CacheStats {
    return CacheStats{
        Size:   c.cache.Len(),
        Hits:   c.hits.Load(),
        Misses: c.misses.Load(),
    }
}

type CacheStats struct {
    Size   int
    Hits   int64
    Misses int64
}

func (s CacheStats) HitRate() float64 {
    total := s.Hits + s.Misses
    if total == 0 {
        return 0
    }
    return float64(s.Hits) / float64(total)
}
```

**Commit**: `feat(domain): add LRU path resolution cache`

##### 18.4.2: Filesystem Metadata Cache

**Test**: `internal/adapters/osfs_cache_test.go`

```go
func TestStatCache(t *testing.T) {
    fs := NewCachedFS(osfs.New(), 100)
    
    // First stat
    info1, err := fs.Stat(context.Background(), "/tmp/test")
    if err != nil {
        t.Fatal(err)
    }
    
    // Second stat (should hit cache)
    info2, err := fs.Stat(context.Background(), "/tmp/test")
    if err != nil {
        t.Fatal(err)
    }
    
    // Verify cache hit
    stats := fs.CacheStats()
    if stats.Hits == 0 {
        t.Error("Expected cache hit")
    }
}

func TestStatCacheInvalidation(t *testing.T) {
    fs := NewCachedFS(osfs.New(), 100)
    
    // Stat file
    fs.Stat(context.Background(), "/tmp/test")
    
    // Modify file (should invalidate cache)
    fs.WriteFile(context.Background(), "/tmp/test", []byte("new content"), 0644)
    
    // Stat again (should miss cache)
    // Verify cache was invalidated
}
```

**Implementation**: Cached filesystem adapter

```go
// internal/adapters/cached_fs.go

type CachedFS struct {
    inner     FS
    statCache *expirable.LRU[string, FileInfo]
    hits      atomic.Int64
    misses    atomic.Int64
}

func NewCachedFS(inner FS, cacheSize int) *CachedFS {
    return &CachedFS{
        inner:     inner,
        statCache: expirable.NewLRU[string, FileInfo](cacheSize, nil, 5*time.Second),
    }
}

func (fs *CachedFS) Stat(ctx context.Context, path string) (FileInfo, error) {
    // Try cache
    if cached, ok := fs.statCache.Get(path); ok {
        fs.hits.Add(1)
        return cached, nil
    }
    
    // Stat filesystem
    fs.misses.Add(1)
    info, err := fs.inner.Stat(ctx, path)
    if err != nil {
        return nil, err
    }
    
    // Cache result
    fs.statCache.Add(path, info)
    return info, nil
}

func (fs *CachedFS) WriteFile(ctx context.Context, path string, data []byte, perm os.FileMode) error {
    // Invalidate cache entry
    fs.statCache.Remove(path)
    
    return fs.inner.WriteFile(ctx, path, data, perm)
}

// Implement other FS methods with cache invalidation...

func (fs *CachedFS) CacheStats() CacheStats {
    return CacheStats{
        Size:   fs.statCache.Len(),
        Hits:   fs.hits.Load(),
        Misses: fs.misses.Load(),
    }
}
```

**Commit**: `feat(adapters): add filesystem metadata cache`

##### 18.4.3: Content Hash Cache

**Test**: `internal/manifest/hash_cache_test.go`

```go
func TestHashCache(t *testing.T) {
    hasher := NewCachedHasher(100)
    
    content := []byte("test content")
    
    // First hash
    hash1 := hasher.Hash(content)
    
    // Second hash (should hit cache)
    hash2 := hasher.Hash(content)
    
    if hash1 != hash2 {
        t.Error("Hash mismatch")
    }
    
    stats := hasher.Stats()
    if stats.Hits == 0 {
        t.Error("Expected cache hit")
    }
}

func TestHashCacheFileContent(t *testing.T) {
    fs := memfs.New()
    fs.WriteFile(context.Background(), "/test.txt", []byte("content"), 0644)
    
    hasher := NewCachedHasher(100)
    
    // Hash file
    hash1, _ := hasher.HashFile(context.Background(), fs, "/test.txt")
    hash2, _ := hasher.HashFile(context.Background(), fs, "/test.txt")
    
    if hash1 != hash2 {
        t.Error("Hash mismatch")
    }
}
```

**Implementation**: Content hash cache

```go
// internal/manifest/hash_cache.go

type CachedHasher struct {
    cache  *expirable.LRU[string, string]
    hits   atomic.Int64
    misses atomic.Int64
}

func NewCachedHasher(size int) *CachedHasher {
    return &CachedHasher{
        cache: expirable.NewLRU[string, string](size, nil, 10*time.Minute),
    }
}

func (h *CachedHasher) HashFile(ctx context.Context, fs FS, path string) (string, error) {
    // Create cache key from path + mtime
    info, err := fs.Stat(ctx, path)
    if err != nil {
        return "", err
    }
    
    cacheKey := fmt.Sprintf("%s:%d", path, info.ModTime().Unix())
    
    // Try cache
    if cached, ok := h.cache.Get(cacheKey); ok {
        h.hits.Add(1)
        return cached, nil
    }
    
    // Compute hash
    h.misses.Add(1)
    content, err := fs.ReadFile(ctx, path)
    if err != nil {
        return "", err
    }
    
    hash := computeHash(content)
    
    // Cache result
    h.cache.Add(cacheKey, hash)
    return hash, nil
}

func (h *CachedHasher) Stats() CacheStats {
    return CacheStats{
        Size:   h.cache.Len(),
        Hits:   h.hits.Load(),
        Misses: h.misses.Load(),
    }
}
```

**Commit**: `feat(manifest): add content hash cache`

**Deliverable 18.4**: Strategic caching with measurable hit rates

---

### Phase 18.5: Performance Regression Suite

#### Objectives

- Create comprehensive benchmark suite
- Integrate benchmarks into CI
- Establish performance baselines
- Detect regressions automatically

#### Tasks

##### 18.5.1: Benchmark Suite

**Test**: `internal/benchmark/suite_test.go`

```go
// Comprehensive benchmark suite covering all critical paths
func BenchmarkSuite(b *testing.B) {
    b.Run("Scanner", func(b *testing.B) {
        b.Run("Small", benchmarkScanSmall)
        b.Run("Medium", benchmarkScanMedium)
        b.Run("Large", benchmarkScanLarge)
        b.Run("Deep", benchmarkScanDeep)
        b.Run("Wide", benchmarkScanWide)
    })
    
    b.Run("Planner", func(b *testing.B) {
        b.Run("DesiredState", benchmarkPlanDesiredState)
        b.Run("DiffStates", benchmarkPlanDiff)
        b.Run("Incremental", benchmarkPlanIncremental)
    })
    
    b.Run("Resolver", func(b *testing.B) {
        b.Run("NoConflicts", benchmarkResolveClean)
        b.Run("WithConflicts", benchmarkResolveConflicts)
    })
    
    b.Run("Sorter", func(b *testing.B) {
        b.Run("TopologicalSort", benchmarkTopoSort)
        b.Run("ParallelPlan", benchmarkParallelPlan)
    })
    
    b.Run("Executor", func(b *testing.B) {
        b.Run("Sequential", benchmarkExecuteSequential)
        b.Run("Parallel", benchmarkExecuteParallel)
    })
    
    b.Run("Ignore", func(b *testing.B) {
        b.Run("PatternMatch", benchmarkIgnoreMatch)
        b.Run("PatternCache", benchmarkIgnoreCache)
    })
    
    b.Run("Manifest", func(b *testing.B) {
        b.Run("Hash", benchmarkManifestHash)
        b.Run("Serialize", benchmarkManifestSerialize)
    })
    
    b.Run("EndToEnd", func(b *testing.B) {
        b.Run("StowSmall", benchmarkE2EStowSmall)
        b.Run("StowLarge", benchmarkE2EStowLarge)
        b.Run("Restow", benchmarkE2ERestow)
    })
}
```

**Commit**: `test(benchmark): add comprehensive benchmark suite`

##### 18.5.2: CI Integration

**Task**: Add benchmark CI job

`.github/workflows/benchmark.yml`:

```yaml
name: Benchmark

on:
  pull_request:
    branches: [main]
  push:
    branches: [main]

jobs:
  benchmark:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0  # Need full history for comparison
      
      - uses: actions/setup-go@v5
        with:
          go-version: '1.25.1'
      
      - name: Run benchmarks
        run: |
          go test -bench=. -benchmem -benchtime=5s \
            ./internal/benchmark/... | tee benchmark.txt
      
      - name: Compare with baseline
        if: github.event_name == 'pull_request'
        run: |
          # Checkout main branch
          git checkout main
          go test -bench=. -benchmem -benchtime=5s \
            ./internal/benchmark/... | tee baseline.txt
          
          # Compare
          go install golang.org/x/perf/cmd/benchstat@latest
          benchstat baseline.txt benchmark.txt
      
      - name: Upload results
        uses: actions/upload-artifact@v4
        with:
          name: benchmark-results
          path: |
            benchmark.txt
            baseline.txt
```

**Commit**: `ci(benchmark): add benchmark CI workflow`

##### 18.5.3: Performance Baseline Documentation

**Documentation**: `docs/Performance-Baseline.md`

```markdown
# Performance Baseline

Baseline performance measurements for dot CLI operations.

## Environment

- CPU: Intel Core i7-9750H (6 cores, 12 threads)
- RAM: 16GB DDR4
- OS: Ubuntu 22.04
- Go: 1.25.1

## Benchmarks

### Scanner

| Operation | Package Size | Duration | Allocs | Memory |
|-----------|--------------|----------|--------|--------|
| ScanPackage | 10 files | 125 μs | 45 | 8.2 KB |
| ScanPackage | 100 files | 980 μs | 412 | 76 KB |
| ScanPackage | 1000 files | 9.2 ms | 4021 | 742 KB |

### Planner

| Operation | Input Size | Duration | Allocs | Memory |
|-----------|------------|----------|--------|--------|
| DesiredState | 100 files | 215 μs | 89 | 18 KB |
| DiffStates | 100 ops | 342 μs | 124 | 32 KB |

### Executor

| Operation | Ops Count | Duration | Parallelism |
|-----------|-----------|----------|-------------|
| Sequential | 100 | 12 ms | 1x |
| Parallel | 100 | 3.2 ms | 4x speedup |

### End-to-End

| Operation | Packages | Files | Duration |
|-----------|----------|-------|----------|
| Stow | 5 | 50 | 18 ms |
| Stow | 10 | 500 | 125 ms |
| Stow | 20 | 2000 | 480 ms |
| Restow (no changes) | 10 | 500 | 8 ms |

## Acceptance Criteria

Performance regressions exceeding these thresholds require justification:

- **Scanner**: < 10 μs per file
- **Planner**: < 5 μs per operation
- **Executor**: > 3x parallelism speedup for 100+ ops
- **End-to-end**: < 100ms for typical usage (10 packages, 100 files)
- **Memory**: < 100MB for large operations (100 packages, 10000 files)

## Updating Baselines

Baselines should be updated when:
- Performance improvements are intentionally made
- Benchmark methodology changes
- Test environment changes significantly

Update process:
1. Run full benchmark suite: `make benchmark`
2. Update this document with new measurements
3. Commit with justification for changes
```

**Commit**: `docs(perf): add performance baseline documentation`

##### 18.5.4: Makefile Integration

**Task**: Add benchmark targets to Makefile

```makefile
# Performance targets
.PHONY: benchmark benchmark-compare benchmark-cpu benchmark-mem profile-analyze

benchmark: ## Run benchmark suite
    @echo "Running benchmarks..."
    go test -bench=. -benchmem -benchtime=5s ./internal/benchmark/... | tee benchmark.txt

benchmark-compare: ## Compare benchmarks with main branch
    @echo "Running baseline benchmarks on main..."
    git stash
    git checkout main
    go test -bench=. -benchmem -benchtime=5s ./internal/benchmark/... | tee baseline.txt
    git checkout -
    git stash pop || true
    @echo "Running current benchmarks..."
    go test -bench=. -benchmem -benchtime=5s ./internal/benchmark/... | tee current.txt
    @echo "Comparing results..."
    benchstat baseline.txt current.txt

benchmark-cpu: ## Run benchmarks with CPU profiling
	go test -bench=. -benchmem -cpuprofile=cpu.prof ./internal/benchmark/...
	go tool pprof -http=:8080 cpu.prof

benchmark-mem: ## Run benchmarks with memory profiling
	go test -bench=. -benchmem -memprofile=mem.prof ./internal/benchmark/...
	go tool pprof -http=:8080 mem.prof

profile-analyze: ## Analyze collected profiles
	@./scripts/analyze_profiles.sh profiles/baseline
```

**Commit**: `build(make): add performance benchmark targets`

**Deliverable 18.5**: Performance regression detection in CI

---

## Testing Strategy

### Unit Testing

Each optimization must include:
- Benchmark showing improvement
- Unit tests verifying correctness
- Property tests (where applicable) ensuring invariants maintained

### Integration Testing

- End-to-end benchmarks for realistic workflows
- Memory profiling for large operations
- Concurrency testing for parallel execution

### Regression Testing

- CI runs benchmark suite on every PR
- Significant regressions block merge
- Performance baselines updated with justification

## Commit Strategy

Follow atomic commit principle with conventional commits:

```text
perf(scanner): reduce allocations in tree building

Optimize node construction to reuse buffers and preallocate slices.

Benchmark results:
  Before: 9.2ms, 4021 allocs, 742 KB
  After:  6.8ms, 2156 allocs, 428 KB
  Improvement: 26% faster, 46% fewer allocs

All tests pass, no functional changes.
```

## Documentation

### Performance Guide

Create `docs/Performance-Guide.md`:

- Performance characteristics
- Tuning parameters
- Profiling instructions
- Common optimization patterns
- Known bottlenecks and workarounds

### API Documentation

Update godoc comments:
- Document performance characteristics of public APIs
- Note caching behavior where relevant
- Document concurrency safety guarantees

## Effort Estimation

| Phase | Tasks | Estimated Hours |
|-------|-------|-----------------|
| 18.1 | Profiling Infrastructure | 8-12 |
| 18.2 | Hot Path Optimization | 16-24 |
| 18.3 | Streaming Optimization | 12-16 |
| 18.4 | Caching Strategy | 12-16 |
| 18.5 | Regression Suite | 8-12 |
| **Total** | | **56-80 hours** |

## Success Metrics

Phase 18 is complete when:

- [ ] CPU profiling shows no function consuming >10% of execution time
- [ ] Memory profiling shows reasonable allocation patterns
- [ ] Benchmark suite covers all critical operations
- [ ] Performance improvements documented with measurements
- [ ] CI detects performance regressions
- [ ] Memory usage bounded for large operations (verified with tests)
- [ ] Parallel execution shows 3x+ speedup for 100+ operations
- [ ] Cache hit rates >80% for realistic workloads
- [ ] All tests pass (80%+ coverage maintained)
- [ ] Documentation updated with performance characteristics

## Dependencies

**Requires**:
- Phase 17 (Integration Testing) - provides realistic test scenarios
- Phase 16 (Property Testing) - verifies correctness during optimization
- Phase 13-14 (CLI) - provides end-to-end test surface

**Blocks**:
- None (optimization is final polish phase)

## Risks and Mitigations

| Risk | Mitigation |
|------|------------|
| Premature optimization | Profile first, optimize only proven bottlenecks |
| Breaking functional correctness | Maintain comprehensive test suite, run after each change |
| Performance regressions | CI benchmark suite catches regressions |
| Over-optimization complexity | Keep optimizations simple, document trade-offs |
| Platform-specific behaviors | Test on multiple platforms (Linux, macOS, Windows) |

## Related Documentation

- [Architecture.md](./Architecture.md) - System architecture
- [Implementation-Plan.md](./Implementation-Plan.md) - Overall implementation plan
- [Features.md](./Features.md) - Performance feature requirements
- [Phase-17-Plan.md](./Phase-17-Plan.md) - Integration testing (prerequisite)

## Acceptance Criteria

Phase 18 is accepted when:

1. **Profiling Complete**: CPU, memory, and goroutine profiles collected and analyzed
2. **Hot Paths Optimized**: All functions consuming >5% CPU optimized with measurable improvements
3. **Streaming Efficient**: Memory usage bounded for arbitrarily large operations (verified with tests)
4. **Caching Effective**: Strategic caches implemented with >80% hit rates
5. **Regression Detection**: CI pipeline catches performance regressions
6. **Documentation Updated**: Performance guide and baselines documented
7. **Tests Pass**: All existing tests pass, 80% coverage maintained
8. **Benchmarks Comprehensive**: Benchmark suite covers all critical paths
9. **Measurements Documented**: Before/after metrics for all optimizations

---

**Version**: 1.0  
**Status**: Draft  
**Last Updated**: 2025-10-05

