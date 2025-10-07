# Phase 16: Property-Based Testing - Detailed Implementation Plan

## Overview

Phase 16 implements a comprehensive property-based testing suite to verify mathematical correctness, algebraic laws, and domain invariants of the dot symlink manager. Using the gopter framework, this phase establishes rigorous verification of system behavior through generative testing.

**Scope**: Verification of algebraic properties, domain invariants, performance characteristics, and error handling correctness through property-based testing.

**Prerequisites**: Phases 1-15 complete with functional core, execution engine, and CLI operational.

**Estimated Effort**: 16-20 hours

## Objectives

1. **Mathematical Correctness**: Verify algebraic laws (idempotence, commutativity, reversibility, associativity)
2. **Invariant Preservation**: Ensure domain invariants hold across all operations
3. **Comprehensive Coverage**: Generate test cases exploring edge cases beyond manual testing
4. **Regression Prevention**: Catch subtle bugs through high-volume random testing
5. **Documentation**: Codify system properties as executable specifications
6. **CI Integration**: Automate property verification in continuous integration

## Architecture Context

### Property-Based Testing Philosophy

Property-based testing verifies universal truths about the system by:
- **Generating** random inputs that satisfy constraints
- **Executing** operations with generated data
- **Asserting** that properties hold for all inputs
- **Shrinking** failing cases to minimal reproducible examples

### Algebraic Laws as Properties

The dot system exhibits several algebraic properties:

```
Idempotence:     manage(manage(P)) = manage(P)
Reversibility:   unmanage(manage(P)) = identity
Commutativity:   manage([A,B]) = manage([B,A])
Associativity:   manage(manage(A) + manage(B)) = manage(A+B)
Conservation:    content(adopt(F)) = content(F)
```

### Integration with Existing Tests

Property tests complement unit and integration tests:
- **Unit Tests**: Verify specific known cases
- **Integration Tests**: Verify end-to-end workflows
- **Property Tests**: Verify universal mathematical properties

## Detailed Implementation Plan

### 16.1: Test Infrastructure Setup

**Objective**: Establish gopter framework integration and test organization.

#### 16.1.1: Gopter Integration

**File**: `tests/properties/framework_test.go`

```go
package properties_test

import (
  "testing"
  "github.com/leanovate/gopter"
  "github.com/leanovate/gopter/gen"
  "github.com/leanovate/gopter/prop"
)

// Standard test parameters across all property tests
func testParameters() *gopter.TestParameters {
  params := gopter.DefaultTestParameters()
  params.MinSuccessfulTests = 100  // Default iteration count
  params.MaxDiscardRatio = 5       // Discard/success ratio
  params.Workers = 4               // Parallel test workers
  params.Rng.Seed(1234)           // Deterministic in CI
  return params
}

// Create properties suite with standard configuration
func newPropertiesSuite() *gopter.Properties {
  return gopter.NewProperties(testParameters())
}

// Helper to run properties with consistent reporting
func runProperties(t *testing.T, props *gopter.Properties, name string) {
  t.Helper()
  result := props.Run(gopter.ConsoleReporter(t))
  if !result {
    t.Errorf("Property suite '%s' failed", name)
  }
}
```

**Tasks**:
- [ ] Add gopter dependency to go.mod
- [ ] Create test/properties/ directory structure
- [ ] Implement standard test parameter configuration
- [ ] Create test harness with consistent reporting
- [ ] Add helper functions for test setup
- [ ] Write infrastructure verification test

**Testing**: Verify gopter framework loads and executes basic properties.

#### 16.1.2: Test Organization

**Directory Structure**:
```
tests/properties/
├── framework_test.go       # Test infrastructure
├── generators_test.go      # Data generators
├── laws_test.go           # Algebraic law tests
├── invariants_test.go     # Domain invariant tests
├── performance_test.go    # Performance property tests
├── errors_test.go         # Error handling tests
├── helpers_test.go        # Test utilities
└── fixtures.go            # Test fixtures
```

**Tasks**:
- [ ] Create package structure with clear separation
- [ ] Define consistent naming conventions
- [ ] Implement test discovery and filtering
- [ ] Add build tags for property test control
- [ ] Create test documentation templates

**Testing**: Verify test organization allows selective execution.

#### 16.1.3: CI Integration

**File**: `.github/workflows/properties.yml`

```yaml
name: Property Tests

on:
  pull_request:
    branches: [ main ]
  schedule:
    - cron: '0 2 * * *'  # Nightly extended runs

jobs:
  property-tests:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        test-config:
          - name: standard
            iterations: 100
          - name: extended
            iterations: 1000
          - name: stress
            iterations: 10000
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.25.1'
      - name: Run Property Tests
        run: |
          go test -v -tags=properties \
            -propertyIterations=${{ matrix.test-config.iterations }} \
            ./tests/properties/...
```

**Tasks**:
- [ ] Create CI workflow for property tests
- [ ] Configure standard (100), extended (1000), stress (10000) runs
- [ ] Add nightly extended test execution
- [ ] Implement property test reporting
- [ ] Add failure artifact collection
- [ ] Document CI configuration

**Testing**: Verify CI correctly executes property tests.

---

### 16.2: Data Generators

**Objective**: Implement random data generators for all domain types.

#### 16.2.1: Path Generators

**File**: `tests/properties/generators_test.go`

```go
package properties_test

import (
  "path/filepath"
  "github.com/leanovate/gopter"
  "github.com/leanovate/gopter/gen"
)

// Generate valid absolute paths
func genAbsolutePath() gopter.Gen {
  return gen.SliceOfN(3, genPathSegment()).
    Map(func(segments []string) string {
      return filepath.Join(append([]string{"/"}, segments...)...)
    })
}

// Generate valid path segments (no special characters)
func genPathSegment() gopter.Gen {
  return gen.Identifier().
    SuchThat(func(s string) bool {
      return s != "." && s != ".." && len(s) > 0 && len(s) < 256
    })
}

// Generate typed StowPath
func genStowPath() gopter.Gen {
  return genAbsolutePath().
    Map(func(p string) StowPath {
      path, _ := NewStowPath(p)
      return path.MustUnwrap()
    })
}

// Generate typed TargetPath
func genTargetPath() gopter.Gen {
  return genAbsolutePath().
    Map(func(p string) TargetPath {
      path, _ := NewTargetPath(p)
      return path.MustUnwrap()
    })
}

// Generate PackagePath relative to stow directory
func genPackagePath(stowDir StowPath) gopter.Gen {
  return genPathSegment().
    Map(func(pkg string) PackagePath {
      return stowDir.Join(pkg)
    })
}

// Generate FilePath with depth control
func genFilePath(base Path, maxDepth int) gopter.Gen {
  return gen.SliceOfN(gen.IntRange(1, maxDepth), genPathSegment()).
    Map(func(segments []string) FilePath {
      p := base.String()
      for _, seg := range segments {
        p = filepath.Join(p, seg)
      }
      return FilePath{p}
    })
}
```

**Tasks**:
- [ ] Implement genAbsolutePath() for valid absolute paths
- [ ] Implement genPathSegment() for valid filename components
- [ ] Implement genStowPath() for typed stow directory paths
- [ ] Implement genTargetPath() for typed target directory paths
- [ ] Implement genPackagePath() for package directory paths
- [ ] Implement genFilePath() with depth control
- [ ] Add path constraint validation (length, characters)
- [ ] Write generator verification tests

**Testing**: Verify all generated paths are valid and type-safe.

#### 16.2.2: Package Structure Generators

**File**: `tests/properties/generators_test.go`

```go
// Generate file tree structure with controlled depth
func genFileTree(maxDepth int, maxBreadth int) gopter.Gen {
  return genFileTreeNode(maxDepth, maxBreadth, 0)
}

func genFileTreeNode(maxDepth, maxBreadth, currentDepth int) gopter.Gen {
  if currentDepth >= maxDepth {
    // Leaf node (file)
    return genPathSegment().Map(func(name string) *Node {
      return &Node{
        Name: name,
        Type: NodeFile,
      }
    })
  }
  
  // Generate directory with children
  return gopter.CombineGens(
    genPathSegment(),
    gen.IntRange(0, maxBreadth),
  ).FlatMap(func(vals []interface{}) gopter.Gen {
    name := vals[0].(string)
    childCount := vals[1].(int)
    
    // Generate children
    childGens := make([]gopter.Gen, childCount)
    for i := 0; i < childCount; i++ {
      childGens[i] = genFileTreeNode(maxDepth, maxBreadth, currentDepth+1)
    }
    
    return gopter.CombineGens(childGens...).Map(func(children []interface{}) *Node {
      node := &Node{
        Name:     name,
        Type:     NodeDir,
        Children: make([]*Node, len(children)),
      }
      for i, child := range children {
        node.Children[i] = child.(*Node)
      }
      return node
    })
  }, reflect.TypeOf(&Node{}))
}

// Generate complete package with metadata
func genPackage(stowDir StowPath) gopter.Gen {
  return gopter.CombineGens(
    genPathSegment(),           // Package name
    genFileTree(4, 5),          // File tree (depth 4, breadth 5)
    genIgnorePatterns(),        // Ignore patterns
    gen.Bool(),                 // Folding enabled
  ).Map(func(vals []interface{}) Package {
    name := vals[0].(string)
    tree := vals[1].(*Node)
    ignore := vals[2].([]Pattern)
    folding := vals[3].(bool)
    
    return Package{
      Name: name,
      Path: stowDir.Join(name),
      Files: FileTree{Root: tree},
      Metadata: PackageMetadata{
        IgnorePatterns: ignore,
        Folding:        folding,
      },
    }
  })
}

// Generate list of packages with controlled count
func genPackageList(stowDir StowPath, minCount, maxCount int) gopter.Gen {
  return gen.IntRange(minCount, maxCount).
    FlatMap(func(count int) gopter.Gen {
      gens := make([]gopter.Gen, count)
      for i := 0; i < count; i++ {
        gens[i] = genPackage(stowDir)
      }
      return gopter.CombineGens(gens...).Map(func(pkgs []interface{}) []Package {
        result := make([]Package, len(pkgs))
        for i, pkg := range pkgs {
          result[i] = pkg.(Package)
        }
        return result
      })
    }, reflect.TypeOf([]Package{}))
}
```

**Tasks**:
- [ ] Implement genFileTree() with depth and breadth control
- [ ] Implement genFileTreeNode() for recursive tree generation
- [ ] Implement genPackage() with complete metadata
- [ ] Implement genPackageList() for multiple packages
- [ ] Add constraints to prevent degenerate cases
- [ ] Add package name uniqueness guarantee
- [ ] Write generator verification tests

**Testing**: Verify generated package structures are valid and realistic.

#### 16.2.3: Operation Generators

**File**: `tests/properties/generators_test.go`

```go
// Generate LinkCreate operation
func genLinkCreate(targetDir TargetPath, stowDir StowPath) gopter.Gen {
  return gopter.CombineGens(
    genFilePath(stowDir, 3),
    genFilePath(targetDir, 3),
    gen.OneConstOf(LinkRelative, LinkAbsolute),
  ).Map(func(vals []interface{}) Operation {
    return LinkCreate{
      ID:     OperationID(uuid.New().String()),
      Source: vals[0].(FilePath),
      Target: vals[1].(TargetPath),
      Mode:   vals[2].(LinkMode),
    }
  })
}

// Generate operation with valid dependencies
func genOperationWithDeps(existing []Operation) gopter.Gen {
  return gen.OneGenOf(
    genLinkCreate(defaultTarget, defaultStow),
    genLinkDelete(defaultTarget),
    genDirCreate(defaultTarget),
    genDirDelete(defaultTarget),
  ).Map(func(op Operation) Operation {
    // Assign dependencies from existing operations
    if len(existing) > 0 && rand.Float32() < 0.3 {
      depCount := rand.Intn(min(3, len(existing))) + 1
      deps := make([]OperationID, depCount)
      for i := 0; i < depCount; i++ {
        deps[i] = existing[rand.Intn(len(existing))].ID()
      }
      op.SetDependencies(deps)
    }
    return op
  })
}

// Generate valid operation sequence (topologically sorted)
func genOperationSequence(minOps, maxOps int) gopter.Gen {
  return gen.IntRange(minOps, maxOps).
    FlatMap(func(count int) gopter.Gen {
      ops := make([]Operation, 0, count)
      
      // Build operations incrementally with valid dependencies
      for i := 0; i < count; i++ {
        gen := genOperationWithDeps(ops)
        // Generate and append
        result := gen.Sample()
        ops = append(ops, result.(Operation))
      }
      
      return gopter.CombineGens(gen.Const(ops))
    }, reflect.TypeOf([]Operation{}))
}
```

**Tasks**:
- [ ] Implement genLinkCreate() for link creation operations
- [ ] Implement genLinkDelete() for link deletion operations
- [ ] Implement genDirCreate() for directory creation operations
- [ ] Implement genDirDelete() for directory deletion operations
- [ ] Implement genFileMove() for file adoption operations
- [ ] Implement genOperationWithDeps() ensuring valid dependency references
- [ ] Implement genOperationSequence() with topological validity
- [ ] Write generator verification tests

**Testing**: Verify generated operations have valid dependencies and form DAGs.

#### 16.2.4: Filesystem State Generators

**File**: `tests/properties/generators_test.go`

```go
// Generate filesystem state with files and directories
func genFilesystemState(targetDir TargetPath, stowDir StowPath) gopter.Gen {
  return gopter.CombineGens(
    genExistingFiles(targetDir, 5, 15),      // 5-15 existing files
    genExistingDirs(targetDir, 3, 10),       // 3-10 existing dirs
    genExistingLinks(targetDir, stowDir, 2, 8), // 2-8 existing links
  ).Map(func(vals []interface{}) FilesystemState {
    return FilesystemState{
      Files:   vals[0].(map[TargetPath]FileInfo),
      Dirs:    vals[1].(map[TargetPath]DirInfo),
      Links:   vals[2].(map[TargetPath]LinkInfo),
    }
  })
}

// Generate existing files with random content
func genExistingFiles(base TargetPath, minCount, maxCount int) gopter.Gen {
  return gen.IntRange(minCount, maxCount).
    FlatMap(func(count int) gopter.Gen {
      fileGens := make([]gopter.Gen, count)
      for i := 0; i < count; i++ {
        fileGens[i] = gopter.CombineGens(
          genFilePath(base, 3),
          gen.SliceOfN(gen.IntRange(0, 1024), gen.Byte()), // Random content
          gen.UInt32Range(0o600, 0o777), // File permissions
        )
      }
      return gopter.CombineGens(fileGens...).Map(func(files []interface{}) map[TargetPath]FileInfo {
        result := make(map[TargetPath]FileInfo)
        for _, f := range files {
          parts := f.([]interface{})
          path := parts[0].(FilePath)
          content := parts[1].([]byte)
          perms := parts[2].(uint32)
          
          result[path] = FileInfo{
            Content:     content,
            Permissions: os.FileMode(perms),
          }
        }
        return result
      })
    }, reflect.TypeOf(map[TargetPath]FileInfo{}))
}

// Generate content hash for package
func genPackageHash() gopter.Gen {
  return gen.SliceOfN(32, gen.Byte()).
    Map(func(bytes []byte) string {
      return hex.EncodeToString(bytes)
    })
}

// Generate manifest with package state
func genManifest(packages []string) gopter.Gen {
  return gopter.CombineGens(
    gen.TimeRange(time.Now().Add(-30*24*time.Hour), time.Now()),
    genPackageInfoMap(packages),
    genHashMap(packages),
  ).Map(func(vals []interface{}) Manifest {
    return Manifest{
      Version:   "1.0",
      UpdatedAt: vals[0].(time.Time),
      Packages:  vals[1].(map[string]PackageInfo),
      Hashes:    vals[2].(map[string]string),
    }
  })
}
```

**Tasks**:
- [ ] Implement genFilesystemState() for complete filesystem state
- [ ] Implement genExistingFiles() with random content
- [ ] Implement genExistingDirs() for directory structures
- [ ] Implement genExistingLinks() with valid targets
- [ ] Implement genPackageHash() for content hashing
- [ ] Implement genManifest() for state tracking
- [ ] Add realistic file size and content distributions
- [ ] Write generator verification tests

**Testing**: Verify generated filesystem states are valid and realistic.

#### 16.2.5: Generator Composition Utilities

**File**: `tests/properties/generators_test.go`

```go
// Combine generators with constraints
func genConstrained[T any](base gopter.Gen, constraint func(T) bool) gopter.Gen {
  return base.SuchThat(func(v interface{}) bool {
    return constraint(v.(T))
  })
}

// Generate pairs with relationship constraints
func genRelatedPair[A, B any](
  genA gopter.Gen,
  genBFromA func(A) gopter.Gen,
) gopter.Gen {
  return genA.FlatMap(func(a interface{}) gopter.Gen {
    return genBFromA(a.(A)).Map(func(b interface{}) []interface{} {
      return []interface{}{a, b}
    })
  }, reflect.TypeOf([]interface{}{}))
}

// Generate list with uniqueness constraint
func genUniqueList[T comparable](gen gopter.Gen, minLen, maxLen int) gopter.Gen {
  return gen.IntRange(minLen, maxLen*2). // Generate more than needed
    FlatMap(func(count int) gopter.Gen {
      items := make([]gopter.Gen, count)
      for i := 0; i < count; i++ {
        items[i] = gen
      }
      return gopter.CombineGens(items...).Map(func(vals []interface{}) []T {
        // Deduplicate
        seen := make(map[T]bool)
        result := make([]T, 0, len(vals))
        for _, v := range vals {
          item := v.(T)
          if !seen[item] {
            seen[item] = true
            result = append(result, item)
          }
          if len(result) >= maxLen {
            break
          }
        }
        return result
      })
    }, reflect.TypeOf([]T{}))
}
```

**Tasks**:
- [ ] Implement genConstrained() for filtered generation
- [ ] Implement genRelatedPair() for dependent values
- [ ] Implement genUniqueList() with deduplication
- [ ] Implement genNonEmpty() ensuring non-empty collections
- [ ] Implement genSubset() for subset generation
- [ ] Add composition combinators
- [ ] Write composition utility tests

**Testing**: Verify generator utilities produce correct distributions.

---

### 16.3: Algebraic Law Verification

**Objective**: Verify mathematical properties of operations.

#### 16.3.1: Idempotence Laws

**File**: `tests/properties/laws_test.go`

```go
package properties_test

import (
  "testing"
  "github.com/leanovate/gopter"
  "github.com/leanovate/gopter/prop"
)

// Test: manage(manage(P)) = manage(P)
func TestManageIdempotence(t *testing.T) {
  properties := newPropertiesSuite()
  
  properties.Property("manage is idempotent", prop.ForAll(
    func(packages []string) bool {
      ctx := context.Background()
      fs := newTestFS()
      client := newTestClient(fs)
      
      setupPackages(fs, packages)
      
      // First manage
      if err := client.Manage(ctx, packages...); err != nil {
        return false
      }
      state1 := captureFilesystemState(fs)
      
      // Second manage (should be no-op)
      if err := client.Manage(ctx, packages...); err != nil {
        return false
      }
      state2 := captureFilesystemState(fs)
      
      // States should be identical
      return filesystemStatesEqual(state1, state2)
    },
    genPackageList(testStowDir, 1, 5),
  ))
  
  runProperties(t, properties, "ManageIdempotence")
}

// Test: remanage(remanage(P)) = remanage(P)
func TestRemanageIdempotence(t *testing.T) {
  properties := newPropertiesSuite()
  
  properties.Property("remanage is idempotent", prop.ForAll(
    func(packages []string) bool {
      ctx := context.Background()
      fs := newTestFS()
      client := newTestClient(fs)
      
      setupPackages(fs, packages)
      
      // Initial manage
      if err := client.Manage(ctx, packages...); err != nil {
        return false
      }
      
      // First remanage
      if err := client.Remanage(ctx, packages...); err != nil {
        return false
      }
      state1 := captureFilesystemState(fs)
      
      // Second remanage (should be no-op)
      if err := client.Remanage(ctx, packages...); err != nil {
        return false
      }
      state2 := captureFilesystemState(fs)
      
      return filesystemStatesEqual(state1, state2)
    },
    genPackageList(testStowDir, 1, 5),
  ))
  
  runProperties(t, properties, "RemanageIdempotence")
}

// Test: status(P) = status(status(P)) (query idempotence)
func TestStatusIdempotence(t *testing.T) {
  properties := newPropertiesSuite()
  
  properties.Property("status queries are idempotent", prop.ForAll(
    func(packages []string) bool {
      ctx := context.Background()
      fs := newTestFS()
      client := newTestClient(fs)
      
      setupAndManagePackages(fs, client, packages)
      
      // First status
      status1, err1 := client.Status(ctx, packages...)
      if err1 != nil {
        return false
      }
      
      // Second status (should be identical)
      status2, err2 := client.Status(ctx, packages...)
      if err2 != nil {
        return false
      }
      
      return statusesEqual(status1, status2)
    },
    genPackageList(testStowDir, 1, 5),
  ))
  
  runProperties(t, properties, "StatusIdempotence")
}
```

**Tasks**:
- [ ] Implement TestManageIdempotence for repeated manage operations
- [ ] Implement TestRemanageIdempotence for repeated remanage operations
- [ ] Implement TestStatusIdempotence for query operations
- [ ] Implement TestAdoptIdempotence for adoption operations
- [ ] Add idempotence verification helpers
- [ ] Write state comparison utilities
- [ ] Document idempotence guarantees

**Testing**: Verify all idempotence properties hold across 100+ iterations.

#### 16.3.2: Reversibility Laws

**File**: `tests/properties/laws_test.go`

```go
// Test: unmanage(manage(P)) = identity
func TestManageUnmanageReversibility(t *testing.T) {
  properties := newPropertiesSuite()
  
  properties.Property("manage is reversed by unmanage", prop.ForAll(
    func(packages []string) bool {
      ctx := context.Background()
      fs := newTestFS()
      client := newTestClient(fs)
      
      setupPackages(fs, packages)
      
      // Capture initial state
      initial := captureFilesystemState(fs)
      
      // Manage packages
      if err := client.Manage(ctx, packages...); err != nil {
        return false
      }
      
      // Unmanage packages
      if err := client.Unmanage(ctx, packages...); err != nil {
        return false
      }
      
      // Capture final state
      final := captureFilesystemState(fs)
      
      // Should return to initial state
      return filesystemStatesEqual(initial, final)
    },
    genPackageList(testStowDir, 1, 5),
  ))
  
  runProperties(t, properties, "ManageUnmanageReversibility")
}

// Test: unadopt restores original file location
func TestAdoptReversibility(t *testing.T) {
  properties := newPropertiesSuite()
  
  properties.Property("adopt can be reversed by moving files back", prop.ForAll(
    func(files []FileSpec) bool {
      ctx := context.Background()
      fs := newTestFS()
      client := newTestClient(fs)
      
      // Create files in target
      for _, f := range files {
        createFile(fs, f.Path, f.Content)
      }
      initial := captureFilesystemState(fs)
      
      // Adopt files into package
      pkg := "testpkg"
      paths := extractPaths(files)
      if err := client.Adopt(ctx, pkg, paths...); err != nil {
        return false
      }
      
      // Verify files moved and linked
      if !verifyAdoptionComplete(fs, pkg, paths) {
        return false
      }
      
      // Unmanage package
      if err := client.Unmanage(ctx, pkg); err != nil {
        return false
      }
      
      // Move files back from package
      for _, path := range paths {
        pkgPath := filepath.Join(testStowDir, pkg, path)
        if err := fs.Rename(ctx, pkgPath, path); err != nil {
          return false
        }
      }
      
      final := captureFilesystemState(fs)
      
      // Should restore original state
      return filesystemStatesEqual(initial, final)
    },
    genFileSpecList(testTargetDir, 3, 8),
  ))
  
  runProperties(t, properties, "AdoptReversibility")
}
```

**Tasks**:
- [ ] Implement TestManageUnmanageReversibility for install/remove cycles
- [ ] Implement TestAdoptReversibility for adoption reversal
- [ ] Implement TestRollbackReversibility for transaction rollback
- [ ] Add state capture and comparison utilities
- [ ] Add reversibility verification helpers
- [ ] Write filesystem diff utilities
- [ ] Document reversibility guarantees

**Testing**: Verify reversibility holds for various package configurations.

#### 16.3.3: Commutativity Laws

**File**: `tests/properties/laws_test.go`

```go
// Test: manage([A,B]) = manage([B,A])
func TestManageCommutativity(t *testing.T) {
  properties := newPropertiesSuite()
  
  properties.Property("manage order does not matter for non-conflicting packages", prop.ForAll(
    func(packages []string) bool {
      if len(packages) < 2 {
        return true // Trivial case
      }
      
      ctx := context.Background()
      
      // Manage in original order
      fs1 := newTestFS()
      client1 := newTestClient(fs1)
      setupPackages(fs1, packages)
      
      if err := client1.Manage(ctx, packages...); err != nil {
        return false
      }
      state1 := captureFilesystemState(fs1)
      
      // Manage in reversed order
      fs2 := newTestFS()
      client2 := newTestClient(fs2)
      reversed := reverseSlice(packages)
      setupPackages(fs2, reversed)
      
      if err := client2.Manage(ctx, reversed...); err != nil {
        return false
      }
      state2 := captureFilesystemState(fs2)
      
      // States should be equivalent (same links, may differ in creation order)
      return filesystemStatesEquivalent(state1, state2)
    },
    genNonConflictingPackages(testStowDir, 2, 5),
  ))
  
  runProperties(t, properties, "ManageCommutativity")
}

// Test: order independence for parallel operations
func TestOperationCommutativity(t *testing.T) {
  properties := newPropertiesSuite()
  
  properties.Property("independent operations commute", prop.ForAll(
    func(ops []Operation) bool {
      if len(ops) < 2 {
        return true
      }
      
      // Filter to independent operations only
      indepOps := filterIndependentOperations(ops)
      if len(indepOps) < 2 {
        return true
      }
      
      ctx := context.Background()
      
      // Execute in original order
      fs1 := newTestFS()
      executor1 := newTestExecutor(fs1)
      result1 := executeSequence(ctx, executor1, indepOps)
      state1 := captureFilesystemState(fs1)
      
      // Execute in different order
      fs2 := newTestFS()
      executor2 := newTestExecutor(fs2)
      shuffled := shuffleOperations(indepOps)
      result2 := executeSequence(ctx, executor2, shuffled)
      state2 := captureFilesystemState(fs2)
      
      // Results should be equivalent
      return result1.Success == result2.Success &&
             filesystemStatesEquivalent(state1, state2)
    },
    genOperationSequence(2, 10),
  ))
  
  runProperties(t, properties, "OperationCommutativity")
}
```

**Tasks**:
- [ ] Implement TestManageCommutativity for package order independence
- [ ] Implement TestOperationCommutativity for independent operations
- [ ] Implement genNonConflictingPackages() ensuring no file overlaps
- [ ] Add filterIndependentOperations() to extract commutative sets
- [ ] Add filesystemStatesEquivalent() for order-independent comparison
- [ ] Write operation independence detection
- [ ] Document commutativity limitations

**Testing**: Verify commutativity for non-conflicting operations.

#### 16.3.4: Associativity Laws

**File**: `tests/properties/laws_test.go`

```go
// Test: manage(A + B) = manage(A) + manage(B) for non-overlapping packages
func TestManageAssociativity(t *testing.T) {
  properties := newPropertiesSuite()
  
  properties.Property("manage is associative for non-overlapping packages", prop.ForAll(
    func(packages []string) bool {
      if len(packages) < 3 {
        return true // Need at least 3 for meaningful test
      }
      
      ctx := context.Background()
      
      // Split packages into two groups
      split := len(packages) / 2
      groupA := packages[:split]
      groupB := packages[split:]
      
      // Approach 1: manage all at once
      fs1 := newTestFS()
      client1 := newTestClient(fs1)
      setupPackages(fs1, packages)
      
      if err := client1.Manage(ctx, packages...); err != nil {
        return false
      }
      state1 := captureFilesystemState(fs1)
      
      // Approach 2: manage in two steps
      fs2 := newTestFS()
      client2 := newTestClient(fs2)
      setupPackages(fs2, packages)
      
      if err := client2.Manage(ctx, groupA...); err != nil {
        return false
      }
      if err := client2.Manage(ctx, groupB...); err != nil {
        return false
      }
      state2 := captureFilesystemState(fs2)
      
      // States should be equivalent
      return filesystemStatesEquivalent(state1, state2)
    },
    genNonConflictingPackages(testStowDir, 3, 6),
  ))
  
  runProperties(t, properties, "ManageAssociativity")
}
```

**Tasks**:
- [ ] Implement TestManageAssociativity for grouping independence
- [ ] Implement TestOperationAssociativity for operation batching
- [ ] Add package splitting utilities
- [ ] Add grouped execution verification
- [ ] Write associativity verification helpers
- [ ] Document associativity guarantees

**Testing**: Verify associativity for various package groupings.

#### 16.3.5: Conservation Laws

**File**: `tests/properties/laws_test.go`

```go
// Test: adopt preserves file content
func TestAdoptConservation(t *testing.T) {
  properties := newPropertiesSuite()
  
  properties.Property("adopt preserves file content", prop.ForAll(
    func(files []FileSpec) bool {
      if len(files) == 0 {
        return true
      }
      
      ctx := context.Background()
      fs := newTestFS()
      client := newTestClient(fs)
      
      // Create files with specific content
      contentMap := make(map[string][]byte)
      for _, f := range files {
        createFile(fs, f.Path, f.Content)
        contentMap[f.Path] = f.Content
      }
      
      // Adopt files into package
      pkg := "testpkg"
      paths := extractPaths(files)
      if err := client.Adopt(ctx, pkg, paths...); err != nil {
        return false
      }
      
      // Verify content unchanged (accessed via symlinks)
      for path, expectedContent := range contentMap {
        actualContent, err := fs.ReadFile(ctx, path)
        if err != nil {
          return false
        }
        if !bytes.Equal(expectedContent, actualContent) {
          return false
        }
      }
      
      return true
    },
    genFileSpecList(testTargetDir, 1, 10),
  ))
  
  runProperties(t, properties, "AdoptConservation")
}

// Test: link count conservation
func TestLinkCountConservation(t *testing.T) {
  properties := newPropertiesSuite()
  
  properties.Property("manage creates expected number of links", prop.ForAll(
    func(packages []Package) bool {
      ctx := context.Background()
      fs := newTestFS()
      client := newTestClient(fs)
      
      // Count files in packages
      expectedLinks := 0
      for _, pkg := range packages {
        expectedLinks += countFiles(pkg.Files)
      }
      
      // Setup and manage packages
      for _, pkg := range packages {
        setupPackage(fs, pkg)
      }
      
      pkgNames := extractPackageNames(packages)
      if err := client.Manage(ctx, pkgNames...); err != nil {
        return false
      }
      
      // Count created links
      actualLinks := countSymlinks(fs, testTargetDir)
      
      return actualLinks == expectedLinks
    },
    genPackageList(testStowDir, 1, 5),
  ))
  
  runProperties(t, properties, "LinkCountConservation")
}
```

**Tasks**:
- [ ] Implement TestAdoptConservation for content preservation
- [ ] Implement TestLinkCountConservation for link accounting
- [ ] Implement TestPermissionConservation for permission preservation
- [ ] Add content comparison utilities
- [ ] Add link counting utilities
- [ ] Write conservation verification helpers
- [ ] Document conservation guarantees

**Testing**: Verify conservation laws hold across operations.

---

### 16.4: Domain Invariant Verification

**Objective**: Verify system invariants always hold.

#### 16.4.1: Path Invariants

**File**: `tests/properties/invariants_test.go`

```go
package properties_test

// Test: phantom types prevent path mixing
func TestPathTypeInvariants(t *testing.T) {
  properties := newPropertiesSuite()
  
  properties.Property("stow paths never escape stow directory", prop.ForAll(
    func(stowDir StowPath, pkgName string) bool {
      pkgPath := stowDir.Join(pkgName)
      
      // Package path should be within stow directory
      return strings.HasPrefix(pkgPath.String(), stowDir.String())
    },
    genStowPath(),
    genPathSegment(),
  ))
  
  properties.Property("target paths never escape target directory", prop.ForAll(
    func(targetDir TargetPath, relPath string) bool {
      fullPath := targetDir.Join(relPath)
      
      // Should be within target directory (no traversal)
      cleaned := filepath.Clean(fullPath.String())
      return strings.HasPrefix(cleaned, targetDir.String())
    },
    genTargetPath(),
    genPathSegment(),
  ))
  
  properties.Property("paths are always absolute", prop.ForAll(
    func(path StowPath) bool {
      return filepath.IsAbs(path.String())
    },
    genStowPath(),
  ))
  
  runProperties(t, properties, "PathTypeInvariants")
}

// Test: path safety against traversal
func TestPathSafetyInvariants(t *testing.T) {
  properties := newPropertiesSuite()
  
  properties.Property("paths reject traversal attempts", prop.ForAll(
    func(base StowPath) bool {
      // Attempt to create path with traversal
      traversalAttempts := []string{
        "../etc/passwd",
        "../../etc/shadow",
        "./../../../etc/hosts",
        "pkg/../../../etc/sudoers",
      }
      
      for _, attempt := range traversalAttempts {
        result := base.SafeJoin(attempt)
        if result.IsOk() {
          resolved := result.MustUnwrap()
          // Should still be within base
          if !strings.HasPrefix(resolved.String(), base.String()) {
            return false
          }
        }
      }
      
      return true
    },
    genStowPath(),
  ))
  
  runProperties(t, properties, "PathSafetyInvariants")
}
```

**Tasks**:
- [ ] Implement TestPathTypeInvariants for phantom type safety
- [ ] Implement TestPathSafetyInvariants for traversal prevention
- [ ] Implement TestPathContainmentInvariants for directory boundaries
- [ ] Implement TestPathCleaningInvariants for normalization
- [ ] Add path validation verification
- [ ] Write path safety test utilities
- [ ] Document path safety guarantees

**Testing**: Verify path invariants prevent security issues.

#### 16.4.2: Graph Invariants

**File**: `tests/properties/invariants_test.go`

```go
// Test: dependency graphs are acyclic
func TestGraphAcyclicityInvariant(t *testing.T) {
  properties := newPropertiesSuite()
  
  properties.Property("dependency graphs never contain cycles", prop.ForAll(
    func(ops []Operation) bool {
      if len(ops) == 0 {
        return true
      }
      
      // Build dependency graph
      graph := BuildGraph(ops)
      
      // Verify no cycles
      cycle := graph.FindCycle()
      return cycle == nil
    },
    genOperationSequence(5, 20),
  ))
  
  properties.Property("topological sort produces valid order", prop.ForAll(
    func(ops []Operation) bool {
      if len(ops) == 0 {
        return true
      }
      
      graph := BuildGraph(ops)
      sorted, err := graph.TopologicalSort()
      if err != nil {
        return false
      }
      
      // Verify dependencies come before dependents
      seen := make(map[OperationID]bool)
      for _, op := range sorted {
        opID := op.ID()
        
        // All dependencies should have been seen already
        for _, depID := range op.Dependencies() {
          if !seen[depID] {
            return false
          }
        }
        
        seen[opID] = true
      }
      
      return true
    },
    genOperationSequence(5, 20),
  ))
  
  runProperties(t, properties, "GraphAcyclicityInvariant")
}

// Test: graph reachability properties
func TestGraphReachabilityInvariant(t *testing.T) {
  properties := newPropertiesSuite()
  
  properties.Property("all operations are reachable in dependency graph", prop.ForAll(
    func(ops []Operation) bool {
      if len(ops) == 0 {
        return true
      }
      
      graph := BuildGraph(ops)
      
      // Every node should be reachable from some root
      roots := graph.FindRoots() // Nodes with no dependencies
      reachable := make(map[OperationID]bool)
      
      for _, root := range roots {
        markReachable(graph, root.ID(), reachable)
      }
      
      // All operations should be reachable
      for _, op := range ops {
        if !reachable[op.ID()] {
          return false
        }
      }
      
      return true
    },
    genOperationSequence(5, 20),
  ))
  
  runProperties(t, properties, "GraphReachabilityInvariant")
}
```

**Tasks**:
- [ ] Implement TestGraphAcyclicityInvariant for cycle prevention
- [ ] Implement TestGraphReachabilityInvariant for connectivity
- [ ] Implement TestGraphOrderInvariant for topological ordering
- [ ] Implement TestGraphConsistencyInvariant for edge validity
- [ ] Add graph verification utilities
- [ ] Write cycle detection test helpers
- [ ] Document graph invariants

**Testing**: Verify graph invariants prevent invalid operation sequences.

#### 16.4.3: Manifest Invariants

**File**: `tests/properties/invariants_test.go`

```go
// Test: manifest consistency with filesystem
func TestManifestConsistencyInvariant(t *testing.T) {
  properties := newPropertiesSuite()
  
  properties.Property("manifest reflects actual filesystem state", prop.ForAll(
    func(packages []string) bool {
      ctx := context.Background()
      fs := newTestFS()
      client := newTestClient(fs)
      
      setupPackages(fs, packages)
      
      // Manage packages
      if err := client.Manage(ctx, packages...); err != nil {
        return false
      }
      
      // Load manifest
      manifest, err := loadManifest(fs, testTargetDir)
      if err != nil {
        return false
      }
      
      // Verify manifest matches filesystem
      for pkgName, pkgInfo := range manifest.Packages {
        // All links in manifest should exist on filesystem
        for _, linkPath := range pkgInfo.Links {
          if !fs.Exists(ctx, linkPath) {
            return false
          }
          if !fs.IsSymlink(ctx, linkPath) {
            return false
          }
        }
      }
      
      // All symlinks should be in manifest
      symlinks := findAllSymlinks(fs, testTargetDir)
      manifestLinks := extractAllLinks(manifest)
      if !setEqual(symlinks, manifestLinks) {
        return false
      }
      
      return true
    },
    genPackageList(testStowDir, 1, 5),
  ))
  
  runProperties(t, properties, "ManifestConsistencyInvariant")
}

// Test: manifest completeness
func TestManifestCompletenessInvariant(t *testing.T) {
  properties := newPropertiesSuite()
  
  properties.Property("manifest contains all managed packages", prop.ForAll(
    func(packages []string) bool {
      ctx := context.Background()
      fs := newTestFS()
      client := newTestClient(fs)
      
      setupPackages(fs, packages)
      
      // Manage packages
      if err := client.Manage(ctx, packages...); err != nil {
        return false
      }
      
      // Load manifest
      manifest, err := loadManifest(fs, testTargetDir)
      if err != nil {
        return false
      }
      
      // All managed packages should be in manifest
      for _, pkg := range packages {
        if _, exists := manifest.Packages[pkg]; !exists {
          return false
        }
      }
      
      // Manifest should not contain unmanaged packages
      for pkgName := range manifest.Packages {
        if !sliceContains(packages, pkgName) {
          return false
        }
      }
      
      return true
    },
    genPackageList(testStowDir, 1, 5),
  ))
  
  runProperties(t, properties, "ManifestCompletenessInvariant")
}

// Test: hash stability
func TestHashStabilityInvariant(t *testing.T) {
  properties := newPropertiesSuite()
  
  properties.Property("package hash is deterministic", prop.ForAll(
    func(pkg Package) bool {
      ctx := context.Background()
      fs := newTestFS()
      
      // Setup package
      setupPackage(fs, pkg)
      
      // Hash multiple times
      hasher := NewContentHasher()
      hash1 := hasher.HashPackage(ctx, fs, pkg.Path)
      hash2 := hasher.HashPackage(ctx, fs, pkg.Path)
      hash3 := hasher.HashPackage(ctx, fs, pkg.Path)
      
      // All hashes should be identical
      return hash1 == hash2 && hash2 == hash3
    },
    genPackage(testStowDir),
  ))
  
  properties.Property("hash changes when content changes", prop.ForAll(
    func(pkg Package) bool {
      ctx := context.Background()
      fs := newTestFS()
      
      setupPackage(fs, pkg)
      
      hasher := NewContentHasher()
      hash1 := hasher.HashPackage(ctx, fs, pkg.Path)
      
      // Modify package content
      if len(pkg.Files.Root.Children) > 0 {
        firstFile := pkg.Files.Root.Children[0]
        filePath := filepath.Join(pkg.Path.String(), firstFile.Name)
        appendToFile(fs, filePath, []byte("modified"))
      } else {
        return true // No files to modify
      }
      
      hash2 := hasher.HashPackage(ctx, fs, pkg.Path)
      
      // Hash should change
      return hash1 != hash2
    },
    genPackage(testStowDir),
  ))
  
  runProperties(t, properties, "HashStabilityInvariant")
}
```

**Tasks**:
- [ ] Implement TestManifestConsistencyInvariant for filesystem sync
- [ ] Implement TestManifestCompletenessInvariant for package tracking
- [ ] Implement TestHashStabilityInvariant for content hashing
- [ ] Implement TestManifestVersionInvariant for schema compatibility
- [ ] Add manifest verification utilities
- [ ] Write filesystem comparison helpers
- [ ] Document manifest invariants

**Testing**: Verify manifest correctly tracks system state.

#### 16.4.4: Operation Invariants

**File**: `tests/properties/invariants_test.go`

```go
// Test: operation validity
func TestOperationValidityInvariant(t *testing.T) {
  properties := newPropertiesSuite()
  
  properties.Property("all operations pass validation", prop.ForAll(
    func(op Operation) bool {
      // Generated operations should always be valid
      return op.Validate() == nil
    },
    genOperation(testStowDir, testTargetDir),
  ))
  
  properties.Property("operations have unique IDs", prop.ForAll(
    func(ops []Operation) bool {
      if len(ops) <= 1 {
        return true
      }
      
      seen := make(map[OperationID]bool)
      for _, op := range ops {
        id := op.ID()
        if seen[id] {
          return false // Duplicate ID
        }
        seen[id] = true
      }
      
      return true
    },
    genOperationSequence(5, 20),
  ))
  
  runProperties(t, properties, "OperationValidityInvariant")
}

// Test: operation atomicity
func TestOperationAtomicityInvariant(t *testing.T) {
  properties := newPropertiesSuite()
  
  properties.Property("operations are atomic (succeed or rollback completely)", prop.ForAll(
    func(ops []Operation) bool {
      if len(ops) == 0 {
        return true
      }
      
      ctx := context.Background()
      fs := newTestFS()
      executor := newTestExecutor(fs)
      
      initial := captureFilesystemState(fs)
      
      // Execute operations
      result := executor.Execute(ctx, Plan{Operations: ops})
      
      if result.IsOk() {
        // Success: state should have changed
        final := captureFilesystemState(fs)
        return !filesystemStatesEqual(initial, final) || len(ops) == 0
      } else {
        // Failure: state should be rolled back
        final := captureFilesystemState(fs)
        return filesystemStatesEqual(initial, final)
      }
    },
    genOperationSequence(3, 10),
  ))
  
  runProperties(t, properties, "OperationAtomicityInvariant")
}
```

**Tasks**:
- [ ] Implement TestOperationValidityInvariant for operation validation
- [ ] Implement TestOperationAtomicityInvariant for transaction semantics
- [ ] Implement TestOperationDependencyInvariant for dependency validity
- [ ] Implement TestOperationRollbackInvariant for rollback correctness
- [ ] Add operation verification utilities
- [ ] Write atomicity test helpers
- [ ] Document operation invariants

**Testing**: Verify operations maintain valid system states.

#### 16.4.5: Conflict Invariants

**File**: `tests/properties/invariants_test.go`

```go
// Test: conflict detection completeness
func TestConflictDetectionInvariant(t *testing.T) {
  properties := newPropertiesSuite()
  
  properties.Property("all actual conflicts are detected", prop.ForAll(
    func(packages []Package) bool {
      ctx := context.Background()
      fs := newTestFS()
      client := newTestClient(fs)
      
      // Setup packages
      for _, pkg := range packages {
        setupPackage(fs, pkg)
      }
      
      // Create pre-existing conflicting files
      conflicts := createRandomConflicts(fs, packages)
      
      // Plan management
      pkgNames := extractPackageNames(packages)
      plan, err := client.PlanManage(ctx, pkgNames...)
      if err != nil {
        return false
      }
      
      // Verify all conflicts were detected
      detectedPaths := extractConflictPaths(plan.Conflicts)
      expectedPaths := extractConflictPaths(conflicts)
      
      return setEqual(detectedPaths, expectedPaths)
    },
    genPackageList(testStowDir, 2, 4),
  ))
  
  properties.Property("no false positive conflicts", prop.ForAll(
    func(packages []Package) bool {
      ctx := context.Background()
      fs := newTestFS()
      client := newTestClient(fs)
      
      // Setup packages with NO conflicts
      for _, pkg := range packages {
        setupPackage(fs, pkg)
      }
      
      // Plan management
      pkgNames := extractPackageNames(packages)
      plan, err := client.PlanManage(ctx, pkgNames...)
      if err != nil {
        return false
      }
      
      // Should have no conflicts
      return len(plan.Conflicts) == 0
    },
    genNonConflictingPackages(testStowDir, 2, 4),
  ))
  
  runProperties(t, properties, "ConflictDetectionInvariant")
}
```

**Tasks**:
- [ ] Implement TestConflictDetectionInvariant for completeness
- [ ] Implement TestConflictFalsePositiveInvariant for precision
- [ ] Implement TestConflictCategorizationInvariant for type accuracy
- [ ] Add conflict generation utilities
- [ ] Write conflict verification helpers
- [ ] Document conflict detection guarantees

**Testing**: Verify conflict detection is complete and accurate.

---

### 16.5: Performance Properties

**Objective**: Verify performance characteristics and complexity bounds.

#### 16.5.1: Algorithmic Complexity

**File**: `tests/properties/performance_test.go`

```go
package properties_test

import (
  "testing"
  "time"
)

// Test: scanning scales linearly with file count
func TestScanningComplexity(t *testing.T) {
  sizes := []int{10, 100, 1000, 10000}
  times := make([]time.Duration, len(sizes))
  
  for i, size := range sizes {
    fs := newTestFS()
    pkg := generateLargePackage(fs, size)
    
    start := time.Now()
    _, err := scanPackage(context.Background(), fs, pkg.Path)
    times[i] = time.Since(start)
    
    if err != nil {
      t.Fatalf("Scan failed for size %d: %v", size, err)
    }
  }
  
  // Verify linear scaling (with tolerance)
  for i := 1; i < len(sizes); i++ {
    ratio := float64(times[i]) / float64(times[i-1])
    sizeRatio := float64(sizes[i]) / float64(sizes[i-1])
    
    // Allow 2x deviation from linear
    if ratio > sizeRatio*2 {
      t.Errorf("Scanning complexity exceeds linear: %d->%d files took %.2fx time (expected %.2fx)",
        sizes[i-1], sizes[i], ratio, sizeRatio)
    }
  }
}

// Test: topological sort is O(V + E)
func TestTopologicalSortComplexity(t *testing.T) {
  properties := newPropertiesSuite()
  
  properties.Property("topological sort completes in reasonable time", prop.ForAll(
    func(opCount int) bool {
      ops := generateOperationsWithDeps(opCount)
      
      start := time.Now()
      graph := BuildGraph(ops)
      _, err := graph.TopologicalSort()
      duration := time.Since(start)
      
      if err != nil {
        return false
      }
      
      // Should complete in O(V + E) ~ O(n) time
      // Allow 1ms per operation as reasonable bound
      maxTime := time.Duration(opCount) * time.Millisecond
      
      return duration < maxTime
    },
    gen.IntRange(10, 1000),
  ))
  
  runProperties(t, properties, "TopologicalSortComplexity")
}

// Test: plan generation scales with package size
func TestPlanningComplexity(t *testing.T) {
  properties := newPropertiesSuite()
  
  properties.Property("planning time grows sub-quadratically", prop.ForAll(
    func(fileCount int) bool {
      ctx := context.Background()
      fs := newTestFS()
      
      pkg := generatePackageWithFiles(fs, fileCount)
      
      start := time.Now()
      _, err := planManage(ctx, fs, []Package{pkg})
      duration := time.Since(start)
      
      if err != nil {
        return false
      }
      
      // Should be roughly O(n log n) or better
      // Allow 100μs per file
      maxTime := time.Duration(fileCount) * 100 * time.Microsecond
      
      return duration < maxTime
    },
    gen.IntRange(10, 500),
  ))
  
  runProperties(t, properties, "PlanningComplexity")
}
```

**Tasks**:
- [ ] Implement TestScanningComplexity for scan performance
- [ ] Implement TestTopologicalSortComplexity for graph algorithms
- [ ] Implement TestPlanningComplexity for planning performance
- [ ] Implement TestExecutionComplexity for execution performance
- [ ] Add complexity measurement utilities
- [ ] Write performance benchmarking helpers
- [ ] Document complexity guarantees

**Testing**: Verify performance scales as expected.

#### 16.5.2: Incremental Operation Performance

**File**: `tests/properties/performance_test.go`

```go
// Test: incremental remanage is faster than full remanage
func TestIncrementalRemanagePerformance(t *testing.T) {
  properties := newPropertiesSuite()
  
  properties.Property("incremental remanage is faster when few packages change", prop.ForAll(
    func(packages []string, changedIdx int) bool {
      if len(packages) < 5 {
        return true // Need enough packages for meaningful test
      }
      
      ctx := context.Background()
      
      // Full remanage
      fs1 := newTestFS()
      client1 := newTestClient(fs1)
      setupPackages(fs1, packages)
      client1.Manage(ctx, packages...)
      
      // Change one package
      modifyPackage(fs1, packages[changedIdx%len(packages)])
      
      start1 := time.Now()
      client1.Remanage(ctx, packages...)
      fullTime := time.Since(start1)
      
      // Incremental remanage with manifest
      fs2 := newTestFS()
      client2 := newTestClientWithManifest(fs2)
      setupPackages(fs2, packages)
      client2.Manage(ctx, packages...)
      
      // Change same package
      modifyPackage(fs2, packages[changedIdx%len(packages)])
      
      start2 := time.Now()
      client2.Remanage(ctx, packages...)
      incrementalTime := time.Since(start2)
      
      // Incremental should be faster (at least 2x)
      return incrementalTime < fullTime/2
    },
    genPackageList(testStowDir, 5, 10),
    gen.IntRange(0, 1000),
  ))
  
  runProperties(t, properties, "IncrementalRemanagePerformance")
}

// Test: manifest-based status is fast
func TestManifestStatusPerformance(t *testing.T) {
  properties := newPropertiesSuite()
  
  properties.Property("status query with manifest is fast", prop.ForAll(
    func(packages []string) bool {
      ctx := context.Background()
      fs := newTestFS()
      client := newTestClientWithManifest(fs)
      
      setupPackages(fs, packages)
      client.Manage(ctx, packages...)
      
      // Status should complete quickly
      start := time.Now()
      _, err := client.Status(ctx, packages...)
      duration := time.Since(start)
      
      if err != nil {
        return false
      }
      
      // Should be sub-millisecond for reasonable package count
      maxTime := time.Duration(len(packages)) * 100 * time.Microsecond
      
      return duration < maxTime
    },
    genPackageList(testStowDir, 1, 20),
  ))
  
  runProperties(t, properties, "ManifestStatusPerformance")
}
```

**Tasks**:
- [ ] Implement TestIncrementalRemanagePerformance for incremental ops
- [ ] Implement TestManifestStatusPerformance for manifest queries
- [ ] Implement TestHashingPerformance for content hashing
- [ ] Implement TestCachingPerformance for pattern caching
- [ ] Add performance measurement utilities
- [ ] Write timing comparison helpers
- [ ] Document performance expectations

**Testing**: Verify incremental operations provide expected speedups.

#### 16.5.3: Parallelization Correctness

**File**: `tests/properties/performance_test.go`

```go
// Test: parallel execution produces same result as sequential
func TestParallelExecutionCorrectness(t *testing.T) {
  properties := newPropertiesSuite()
  
  properties.Property("parallel execution equivalent to sequential", prop.ForAll(
    func(ops []Operation) bool {
      if len(ops) < 2 {
        return true
      }
      
      ctx := context.Background()
      
      // Sequential execution
      fs1 := newTestFS()
      executor1 := newTestExecutor(fs1)
      result1 := executor1.ExecuteSequential(ctx, Plan{Operations: ops})
      state1 := captureFilesystemState(fs1)
      
      // Parallel execution
      fs2 := newTestFS()
      executor2 := newTestExecutor(fs2)
      result2 := executor2.ExecuteParallel(ctx, Plan{Operations: ops})
      state2 := captureFilesystemState(fs2)
      
      // Results should be equivalent
      return result1.Success == result2.Success &&
             filesystemStatesEquivalent(state1, state2)
    },
    genOperationSequence(5, 15),
  ))
  
  runProperties(t, properties, "ParallelExecutionCorrectness")
}

// Test: parallel execution provides speedup
func TestParallelExecutionSpeedup(t *testing.T) {
  properties := newPropertiesSuite()
  
  properties.Property("parallel execution is faster for large operations", prop.ForAll(
    func(opCount int) bool {
      if opCount < 10 {
        return true // Need enough ops for parallelism benefit
      }
      
      ctx := context.Background()
      ops := generateIndependentOperations(opCount)
      
      // Sequential execution
      fs1 := newTestFS()
      executor1 := newTestExecutor(fs1)
      start1 := time.Now()
      executor1.ExecuteSequential(ctx, Plan{Operations: ops})
      seqTime := time.Since(start1)
      
      // Parallel execution
      fs2 := newTestFS()
      executor2 := newTestExecutor(fs2)
      start2 := time.Now()
      executor2.ExecuteParallel(ctx, Plan{Operations: ops})
      parTime := time.Since(start2)
      
      // Parallel should be faster (at least 1.5x)
      return parTime < seqTime*2/3
    },
    gen.IntRange(10, 100),
  ))
  
  runProperties(t, properties, "ParallelExecutionSpeedup")
}
```

**Tasks**:
- [ ] Implement TestParallelExecutionCorrectness for result equivalence
- [ ] Implement TestParallelExecutionSpeedup for performance gain
- [ ] Implement TestParallelScanningCorrectness for concurrent scanning
- [ ] Add parallel execution test utilities
- [ ] Write speedup measurement helpers
- [ ] Document parallelization guarantees

**Testing**: Verify parallelization is both correct and beneficial.

---

### 16.6: Error Handling Properties

**Objective**: Verify error handling completeness and correctness.

#### 16.6.1: Error Propagation

**File**: `tests/properties/errors_test.go`

```go
package properties_test

// Test: errors are never silently dropped
func TestErrorPropagationCompleteness(t *testing.T) {
  properties := newPropertiesSuite()
  
  properties.Property("all operation errors are propagated", prop.ForAll(
    func(ops []Operation, failureIdx int) bool {
      if len(ops) == 0 {
        return true
      }
      
      ctx := context.Background()
      fs := newTestFS()
      
      // Inject failure at specific operation
      injectFailure(fs, ops[failureIdx%len(ops)])
      
      executor := newTestExecutor(fs)
      result := executor.Execute(ctx, Plan{Operations: ops})
      
      // Error should be captured in result
      return !result.IsOk() && len(result.Errors) > 0
    },
    genOperationSequence(3, 10),
    gen.IntRange(0, 1000),
  ))
  
  properties.Property("multiple errors are collected", prop.ForAll(
    func(ops []Operation) bool {
      if len(ops) < 3 {
        return true
      }
      
      ctx := context.Background()
      fs := newTestFS()
      
      // Inject multiple failures
      failCount := min(3, len(ops))
      for i := 0; i < failCount; i++ {
        injectFailure(fs, ops[i])
      }
      
      executor := newTestExecutor(fs)
      result := executor.Execute(ctx, Plan{Operations: ops})
      
      // Should collect multiple errors
      return !result.IsOk() && len(result.Errors) >= failCount
    },
    genOperationSequence(5, 15),
  ))
  
  runProperties(t, properties, "ErrorPropagationCompleteness")
}

// Test: error context is preserved
func TestErrorContextPreservation(t *testing.T) {
  properties := newPropertiesSuite()
  
  properties.Property("errors contain operation context", prop.ForAll(
    func(ops []Operation, failureIdx int) bool {
      if len(ops) == 0 {
        return true
      }
      
      ctx := context.Background()
      fs := newTestFS()
      
      failOp := ops[failureIdx%len(ops)]
      injectFailure(fs, failOp)
      
      executor := newTestExecutor(fs)
      result := executor.Execute(ctx, Plan{Operations: ops})
      
      // Error should reference the failed operation
      if result.IsOk() {
        return false
      }
      
      // Check that error mentions operation details
      errStr := result.Errors[0].Error()
      return strings.Contains(errStr, failOp.ID().String())
    },
    genOperationSequence(3, 10),
    gen.IntRange(0, 1000),
  ))
  
  runProperties(t, properties, "ErrorContextPreservation")
}
```

**Tasks**:
- [ ] Implement TestErrorPropagationCompleteness for error capture
- [ ] Implement TestErrorContextPreservation for error details
- [ ] Implement TestErrorWrappingCorrectness for error chains
- [ ] Add error injection utilities
- [ ] Write error verification helpers
- [ ] Document error handling guarantees

**Testing**: Verify errors are properly propagated and contextualized.

#### 16.6.2: Rollback Correctness

**File**: `tests/properties/errors_test.go`

```go
// Test: rollback restores original state
func TestRollbackCorrectness(t *testing.T) {
  properties := newPropertiesSuite()
  
  properties.Property("rollback restores state after partial failure", prop.ForAll(
    func(ops []Operation, failureIdx int) bool {
      if len(ops) < 2 {
        return true
      }
      
      ctx := context.Background()
      fs := newTestFS()
      executor := newTestExecutor(fs)
      
      // Capture initial state
      initial := captureFilesystemState(fs)
      
      // Inject failure partway through
      failOp := ops[failureIdx%len(ops)]
      injectFailure(fs, failOp)
      
      // Execute with expected failure
      result := executor.Execute(ctx, Plan{Operations: ops})
      
      // Should have failed and rolled back
      if result.IsOk() {
        return false
      }
      
      // State should be restored
      final := captureFilesystemState(fs)
      return filesystemStatesEqual(initial, final)
    },
    genOperationSequence(3, 10),
    gen.IntRange(0, 1000),
  ))
  
  properties.Property("rollback is idempotent", prop.ForAll(
    func(ops []Operation) bool {
      if len(ops) == 0 {
        return true
      }
      
      ctx := context.Background()
      fs := newTestFS()
      executor := newTestExecutor(fs)
      
      // Execute operations
      result := executor.Execute(ctx, Plan{Operations: ops})
      if !result.IsOk() {
        return true // Already rolled back
      }
      
      // Manually rollback
      rollback1 := executor.Rollback(ctx, result.Executed)
      state1 := captureFilesystemState(fs)
      
      // Rollback again (should be no-op)
      rollback2 := executor.Rollback(ctx, result.Executed)
      state2 := captureFilesystemState(fs)
      
      // States should be identical
      return len(rollback2) == 0 && filesystemStatesEqual(state1, state2)
    },
    genOperationSequence(3, 10),
  ))
  
  runProperties(t, properties, "RollbackCorrectness")
}
```

**Tasks**:
- [ ] Implement TestRollbackCorrectness for state restoration
- [ ] Implement TestRollbackIdempotence for repeated rollback
- [ ] Implement TestRollbackOrdering for dependency-aware rollback
- [ ] Add rollback testing utilities
- [ ] Write state restoration verification
- [ ] Document rollback guarantees

**Testing**: Verify rollback correctly restores state.

#### 16.6.3: Validation Exhaustiveness

**File**: `tests/properties/errors_test.go`

```go
// Test: validation catches all invalid operations
func TestValidationExhaustiveness(t *testing.T) {
  properties := newPropertiesSuite()
  
  properties.Property("validation rejects invalid operations", prop.ForAll(
    func(op InvalidOperation) bool {
      // Generated invalid operation should fail validation
      return op.Validate() != nil
    },
    genInvalidOperation(),
  ))
  
  properties.Property("validation catches permission issues", prop.ForAll(
    func(ops []Operation) bool {
      ctx := context.Background()
      fs := newRestrictedFS() // Filesystem with permission restrictions
      executor := newTestExecutor(fs)
      
      // Validation should catch permission errors
      err := executor.Prepare(ctx, Plan{Operations: ops})
      
      // Should fail preparation for restricted filesystem
      return err != nil
    },
    genOperationSequence(3, 10),
  ))
  
  properties.Property("validation catches cyclic dependencies", prop.ForAll(
    func(ops []Operation) bool {
      // Inject cycle
      if len(ops) >= 2 {
        ops[0].SetDependencies([]OperationID{ops[len(ops)-1].ID()})
        ops[len(ops)-1].SetDependencies([]OperationID{ops[0].ID()})
      }
      
      graph := BuildGraph(ops)
      cycle := graph.FindCycle()
      
      // Should detect cycle
      return cycle != nil
    },
    genOperationSequence(2, 10),
  ))
  
  runProperties(t, properties, "ValidationExhaustiveness")
}
```

**Tasks**:
- [ ] Implement TestValidationExhaustiveness for invalid input detection
- [ ] Implement genInvalidOperation() for invalid operation generation
- [ ] Implement TestPermissionValidation for access checks
- [ ] Implement TestCyclicDependencyValidation for cycle detection
- [ ] Add validation testing utilities
- [ ] Write invalid input generators
- [ ] Document validation coverage

**Testing**: Verify validation catches all invalid conditions.

---

### 16.7: Integration and Documentation

**Objective**: Integrate property tests into development workflow.

#### 16.7.1: CI/CD Integration

**File**: `.github/workflows/properties.yml`

```yaml
name: Property Tests

on:
  pull_request:
    branches: [ main ]
  push:
    branches: [ main ]
  schedule:
    - cron: '0 2 * * *'  # Nightly at 2 AM

jobs:
  property-tests-standard:
    name: Property Tests (Standard)
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.25.1'
      
      - name: Run Standard Property Tests
        run: |
          go test -v -tags=properties \
            -propertyIterations=100 \
            -timeout=10m \
            ./tests/properties/...
      
      - name: Upload Test Results
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: property-test-results-standard
          path: tests/properties/*.log
  
  property-tests-extended:
    name: Property Tests (Extended)
    runs-on: ubuntu-latest
    if: github.event_name == 'schedule'
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.25.1'
      
      - name: Run Extended Property Tests
        run: |
          go test -v -tags=properties \
            -propertyIterations=1000 \
            -timeout=30m \
            ./tests/properties/...
      
      - name: Upload Test Results
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: property-test-results-extended
          path: tests/properties/*.log
  
  property-tests-stress:
    name: Property Tests (Stress)
    runs-on: ubuntu-latest
    if: github.event_name == 'schedule'
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.25.1'
      
      - name: Run Stress Property Tests
        run: |
          go test -v -tags=properties \
            -propertyIterations=10000 \
            -timeout=2h \
            ./tests/properties/...
      
      - name: Upload Test Results
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: property-test-results-stress
          path: tests/properties/*.log
      
      - name: Notify on Failure
        if: failure()
        uses: actions/github-script@v7
        with:
          script: |
            github.rest.issues.create({
              owner: context.repo.owner,
              repo: context.repo.repo,
              title: 'Property Tests Failed in Stress Run',
              body: 'Stress property tests failed. See workflow run for details.',
              labels: ['bug', 'property-tests']
            })
```

**Tasks**:
- [ ] Create property test CI workflow
- [ ] Configure standard (PR), extended (nightly), stress (weekly) runs
- [ ] Add test result artifact collection
- [ ] Implement failure notification
- [ ] Add test result reporting
- [ ] Document CI integration

**Testing**: Verify CI correctly runs property tests.

#### 16.7.2: Property Testing Guide

**File**: `docs/Property-Testing-Guide.md`

```markdown
# Property-Based Testing Guide

## Overview

This guide explains the property-based testing approach used in dot and how to write, run, and maintain property tests.

## What is Property-Based Testing?

Property-based testing verifies universal truths about the system by:
1. Generating random inputs
2. Executing operations
3. Asserting properties hold
4. Shrinking failures to minimal cases

## Writing Property Tests

### Basic Structure

[Content describing test structure, generators, properties, etc.]

### Property Categories

#### Algebraic Laws
[Documentation of algebraic properties]

#### Domain Invariants
[Documentation of system invariants]

#### Performance Properties
[Documentation of performance guarantees]

## Running Property Tests

### Local Development
```bash
# Run standard tests (100 iterations)
make test-properties

# Run extended tests (1000 iterations)
make test-properties-extended

# Run stress tests (10000 iterations)
make test-properties-stress
```

### CI Integration
[Documentation of CI workflow]

## Maintenance

### Adding New Properties
[Guidelines for adding properties]

### Debugging Failures
[Guide to investigating property test failures]

### Generator Maintenance
[Guide to maintaining and extending generators]
```

**Tasks**:
- [ ] Write comprehensive property testing guide
- [ ] Document property categories and examples
- [ ] Create property writing tutorials
- [ ] Document generator creation patterns
- [ ] Add troubleshooting guides
- [ ] Include property test maintenance procedures

**Testing**: Review documentation for completeness and clarity.

#### 16.7.3: Example Property Tests

**File**: `tests/properties/examples_test.go`

```go
package properties_test

import (
  "testing"
  "github.com/leanovate/gopter"
  "github.com/leanovate/gopter/gen"
  "github.com/leanovate/gopter/prop"
)

// Example: Simple idempotence property
func ExampleIdempotence() {
  properties := gopter.NewProperties(nil)
  
  properties.Property("repeated application has no effect", prop.ForAll(
    func(x int) bool {
      f := func(n int) int { return n * n }
      return f(f(x)) == f(x) || x != 0 // f(f(x)) = f(x) for x=0,1,-1
    },
    gen.Int(),
  ))
  
  properties.TestingRun(testing.TB(&testing.T{}))
}

// Example: Commutativity property
func ExampleCommutativity() {
  properties := gopter.NewProperties(nil)
  
  properties.Property("addition is commutative", prop.ForAll(
    func(a, b int) bool {
      return a+b == b+a
    },
    gen.Int(),
    gen.Int(),
  ))
  
  properties.TestingRun(testing.TB(&testing.T{}))
}

// Example: Conservation property
func ExampleConservation() {
  properties := gopter.NewProperties(nil)
  
  properties.Property("sort preserves elements", prop.ForAll(
    func(slice []int) bool {
      original := make([]int, len(slice))
      copy(original, slice)
      
      sort.Ints(slice)
      
      // Same elements, different order
      return sameElements(original, slice)
    },
    gen.SliceOf(gen.Int()),
  ))
  
  properties.TestingRun(testing.TB(&testing.T{}))
}
```

**Tasks**:
- [ ] Create example property tests for common patterns
- [ ] Document each example with explanations
- [ ] Create examples for all property categories
- [ ] Add examples to godoc
- [ ] Include examples in guide

**Testing**: Verify examples are clear and educational.

#### 16.7.4: Makefile Integration

**File**: `Makefile`

```makefile
# Property-based testing targets

.PHONY: test-properties
test-properties: ## Run standard property tests (100 iterations)
	@echo "Running standard property tests..."
	go test -v -tags=properties \
		-propertyIterations=100 \
		-timeout=10m \
		./tests/properties/...

.PHONY: test-properties-extended
test-properties-extended: ## Run extended property tests (1000 iterations)
	@echo "Running extended property tests..."
	go test -v -tags=properties \
		-propertyIterations=1000 \
		-timeout=30m \
		./tests/properties/...

.PHONY: test-properties-stress
test-properties-stress: ## Run stress property tests (10000 iterations)
	@echo "Running stress property tests..."
	go test -v -tags=properties \
		-propertyIterations=10000 \
		-timeout=2h \
		./tests/properties/...

.PHONY: test-properties-quick
test-properties-quick: ## Run quick property tests (10 iterations)
	@echo "Running quick property tests..."
	go test -v -tags=properties \
		-propertyIterations=10 \
		-timeout=2m \
		./tests/properties/...

.PHONY: test-all
test-all: test test-integration test-properties ## Run all tests
```

**Tasks**:
- [ ] Add property test targets to Makefile
- [ ] Create convenience targets for different iteration counts
- [ ] Integrate with existing test targets
- [ ] Document Makefile targets

**Testing**: Verify Makefile targets work correctly.

---

## Dependencies

### Phase Dependencies
- **Requires**: Phases 1-15 (domain model, core logic, CLI, integration tests)
- **Blocks**: Phase 20 (final release preparation)

### Package Dependencies
```
tests/properties/
├── Depends on: pkg/dot (public API)
├── Depends on: internal/api (client implementation)
├── Depends on: internal/adapters (test filesystem)
├── Depends on: internal/domain (domain types)
└── Depends on: github.com/leanovate/gopter (framework)
```

## File Structure

```
tests/properties/
├── framework_test.go         # Test infrastructure (400 lines)
├── generators_test.go         # Data generators (800 lines)
├── laws_test.go              # Algebraic law tests (600 lines)
├── invariants_test.go        # Domain invariant tests (700 lines)
├── performance_test.go       # Performance property tests (400 lines)
├── errors_test.go            # Error handling tests (400 lines)
├── helpers_test.go           # Test utilities (300 lines)
├── fixtures.go               # Test fixtures (200 lines)
└── examples_test.go          # Example tests (200 lines)

docs/
└── Property-Testing-Guide.md # Testing guide (2000 lines)

.github/workflows/
└── properties.yml            # CI workflow (150 lines)
```

**Total Estimated Lines**: ~6,150 lines

## Testing Strategy

### Property Test Coverage
1. **Algebraic Laws**: 5 properties × 3 variations = 15 tests
2. **Domain Invariants**: 5 categories × 4 tests = 20 tests
3. **Performance**: 3 categories × 3 tests = 9 tests
4. **Error Handling**: 3 categories × 3 tests = 9 tests

**Total**: ~53 property tests

### Verification Approach
1. Run standard tests (100 iterations) in development
2. Run extended tests (1000 iterations) in CI for PRs
3. Run stress tests (10000 iterations) nightly
4. Monitor shrinking for minimal failure cases
5. Add regression tests for discovered edge cases

## Success Criteria

### Functional Success
- [ ] All property tests pass with 100+ iterations
- [ ] Generators produce valid, diverse inputs
- [ ] Properties verify stated guarantees
- [ ] Shrinking produces minimal failure cases
- [ ] All algebraic laws verified
- [ ] All domain invariants verified
- [ ] Performance properties verified
- [ ] Error handling completeness verified

### Quality Metrics
- [ ] Property test coverage: 100% of major operations
- [ ] Generator coverage: All domain types
- [ ] CI integration: Standard, extended, stress runs
- [ ] Documentation: Complete property testing guide
- [ ] Zero false positives in property tests
- [ ] Fast shrinking to minimal cases (<10 steps)

### Non-Functional Success
- [ ] Standard tests complete in <10 minutes
- [ ] Extended tests complete in <30 minutes
- [ ] Stress tests complete in <2 hours
- [ ] Clear failure reports with context
- [ ] Reproducible test failures with seeds

## Risk Mitigation

### Technical Risks

**Risk**: Property tests too slow for CI
- **Mitigation**: Tiered testing (standard/extended/stress)
- **Mitigation**: Configurable iteration counts
- **Mitigation**: Parallel test execution

**Risk**: Generators produce unrealistic data
- **Mitigation**: Constraint validation on generators
- **Mitigation**: Generator verification tests
- **Mitigation**: Domain expert review of generators

**Risk**: False positive property failures
- **Mitigation**: Careful property formulation
- **Mitigation**: Property review and refinement
- **Mitigation**: Flakiness detection and elimination

**Risk**: Shrinking produces misleading minimal cases
- **Mitigation**: Custom shrinking strategies
- **Mitigation**: Manual verification of shrunken cases
- **Mitigation**: Multiple shrinking approaches

### Process Risks

**Risk**: Property tests not run consistently
- **Mitigation**: CI integration with blocking failures
- **Mitigation**: Pre-commit hooks for quick tests
- **Mitigation**: Visible test status in PRs

**Risk**: Property test maintenance burden
- **Mitigation**: Comprehensive documentation
- **Mitigation**: Example-driven learning
- **Mitigation**: Generator utilities for reuse

## Validation Plan

### Phase 16.1-16.2 Validation
```bash
# Verify framework integration
make test-properties-quick

# Verify generators produce valid data
go test -v -run TestGenerator ./tests/properties/

# Verify generator constraints
go test -v -run TestGeneratorConstraints ./tests/properties/
```

### Phase 16.3-16.4 Validation
```bash
# Verify algebraic laws
go test -v -run TestLaws ./tests/properties/

# Verify domain invariants
go test -v -run TestInvariants ./tests/properties/

# Run with extended iterations
make test-properties-extended
```

### Phase 16.5-16.6 Validation
```bash
# Verify performance properties
go test -v -run TestPerformance ./tests/properties/

# Verify error handling
go test -v -run TestErrors ./tests/properties/
```

### Phase 16.7 Validation
```bash
# Verify CI integration
git push origin feature/phase-16
# Check CI status

# Verify documentation
make docs
# Review generated docs

# Run full test suite
make test-all
```

## Timeline Estimate

### Hour Breakdown
- **16.1**: Test Infrastructure Setup (2 hours)
- **16.2**: Data Generators (4 hours)
- **16.3**: Algebraic Law Verification (3 hours)
- **16.4**: Domain Invariant Verification (3 hours)
- **16.5**: Performance Properties (2 hours)
- **16.6**: Error Handling Properties (2 hours)
- **16.7**: Integration and Documentation (4 hours)

**Total**: 20 hours

### Milestone Schedule
- **Week 1**: Infrastructure and generators (16.1-16.2)
- **Week 2**: Laws and invariants (16.3-16.4)
- **Week 3**: Performance and errors (16.5-16.6)
- **Week 4**: Integration and polish (16.7)

## Commit Strategy

### Atomic Commits
Each subsection becomes one atomic commit:

```
feat(properties): set up property testing infrastructure

feat(properties): implement data generators for domain types

feat(properties): verify idempotence algebraic laws

feat(properties): verify path safety invariants

feat(properties): verify algorithmic complexity bounds

feat(properties): verify error propagation completeness

feat(properties): integrate property tests into CI

docs(properties): add property testing guide
```

## Next Steps

After Phase 16 completion:
1. **Phase 17**: Integration Testing (end-to-end scenarios)
2. **Phase 18**: Performance Optimization (profiling and tuning)
3. **Phase 19**: Documentation (user and developer guides)
4. **Phase 20**: Release Preparation (final polish)

## References

- [gopter Documentation](https://pkg.go.dev/github.com/leanovate/gopter)
- Property-Based Testing Patterns (John Hughes)
- Architecture.md: System design and algebraic properties
- Features.md: Feature specifications for property coverage
- Implementation-Plan.md: Overall project structure
