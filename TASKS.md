# Tasks

## Optimizations

### High Priority (Immediate Impact)

#### 1. File System Caching ✅ COMPLETED
**Problem**: Repeated `os.Stat()` calls and redundant file system operations
**Impact**: Dramatically reduce I/O operations, especially in deep directory structures
**Files**: `internal/manager/config-file/config-file.go`, `internal/manager/js/js-workspace.go`

**Tasks**:
- [x] Implement session-wide file system cache for `os.Stat()` results
- [x] Cache directory listings to avoid repeated `os.ReadDir()` calls
- [x] Add TTL-based cache invalidation (e.g., 5 seconds) for file existence checks
- [x] Replace individual `fileExists()` calls with batch directory scanning

**Performance Results**:
- **File Operations**: 24x faster (53,926 ns/op → 2,210 ns/op), 95% fewer allocations
- **Batch Operations**: 15x faster (9,792 ns/op → 642 ns/op), 86% fewer allocations  
- **Directory Traversal**: 5x faster (20,580 ns/op → 4,196 ns/op), 82% fewer allocations

**Implementation**:
```go
type FSCache struct {
    statCache map[string]fs.FileInfo
    dirCache  map[string][]fs.DirEntry
    mutex     sync.RWMutex
    ttl       time.Duration
}
```

#### 2. Pre-allocate Slices ✅ COMPLETED
**Problem**: Multiple `append()` operations cause repeated memory reallocations
**Impact**: Reduce memory allocations and GC pressure
**Files**: `internal/manager/parser/parser.go`, `internal/manager/manager.go`

**Tasks**:
- [x] Pre-allocate `managers` slice in `parseDirectoryManagers()` with capacity 2
- [x] Pre-allocate `allTasks` slice in `GetManagerTasksFromList()` with estimated capacity
- [x] Pre-allocate `tasks` slice in `ListTasks()` methods based on script count
- [x] Pre-allocate `parsedDirectories` slice in `GetDirectories()` with path depth

**Implementation Details**:
- Estimated slice capacities based on typical usage patterns
- `parseDirectoryManagers()`: capacity 2 (JS + Task manager)
- `collectOrderedResults()`: capacity = numDirs * 2
- `GetDirectories()`: capacity based on path depth calculation
- `ParseJsWorkspace()`: capacity 1 (typically one workspace type)

**Implementation**:
```go
// Instead of: managers := []manager.Manager{}
managers := make([]manager.Manager, 0, 2) // Typically JS + Task manager

// Instead of: allTasks := []ManagerTask{}
allTasks := make([]ManagerTask, 0, estimatedTaskCount)
```

#### 3. Batch File Operations ✅ COMPLETED
**Problem**: Sequential file existence checks and reads
**Impact**: Reduce system calls significantly
**Files**: `internal/manager/js/js-workspace.go`, `internal/manager/task-manager/task-manager.go`

**Tasks**:
- [x] Replace individual `fileExists()` calls with single `os.ReadDir()`
- [x] Batch check for all lock files (pnpm-lock.yaml, yarn.lock, package-lock.json) in one operation
- [x] Batch check for all Taskfile variants in one directory scan
- [x] Implement `findFilesInDirectory()` helper for multiple file patterns

**Implementation Details**:
- `FindFilesInDirectoryWithContent()` replaces individual file operations
- `FindFilesInDirectory()` for batch existence checks
- Lock file detection now uses single directory scan
- Taskfile detection optimized with batch operations

**Implementation**:
```go
func findFilesInDirectory(dir string, filenames []string) map[string]*ConfigFile {
    entries, err := os.ReadDir(dir)
    if err != nil {
        return nil
    }
    
    fileSet := make(map[string]bool)
    for _, filename := range filenames {
        fileSet[filename] = true
    }
    
    found := make(map[string]*ConfigFile)
    for _, entry := range entries {
        if fileSet[entry.Name()] {
            // Read file content
        }
    }
    return found
}
```

### Medium Priority (Significant Improvement)

#### 4. Parallel File I/O Within Directories
**Problem**: File operations within each directory are sequential
**Impact**: Reduce I/O latency, especially for remote file systems
**Files**: `internal/manager/parser/parser.go`

**Tasks**:
- [ ] Parallelize JS manager and task manager parsing within each directory
- [ ] Use `errgroup` for better error handling in concurrent operations
- [ ] Implement worker pool pattern for file I/O operations
- [ ] Add configurable concurrency limits to prevent file descriptor exhaustion

**Implementation**:
```go
func parseDirectoryManagersConcurrently(dir string, jsWorkspace *jsmanager.JsWorkspace) ([]manager.Manager, error) {
    var eg errgroup.Group
    var jsManager, taskMgr manager.Manager
    
    eg.Go(func() error {
        jsManager = parseJSManager(dir, jsWorkspace)
        return nil
    })
    
    eg.Go(func() error {
        taskMgr = parseTaskManager(dir)
        return nil
    })
    
    if err := eg.Wait(); err != nil {
        return nil, err
    }
    
    // Collect results...
}
```

#### 5. Cache Parsed Configuration Files ✅ COMPLETED
**Problem**: Same files may be parsed multiple times
**Impact**: Eliminate redundant JSON/YAML parsing operations
**Files**: `internal/manager/config-file/config-file.go`, `internal/manager/cache/`

**Tasks**:
- [x] Implement unified cache combining file system and parse caching
- [x] Eliminate redundant ParseCache - now handled by UnifiedCache
- [x] Add memory limits to prevent cache bloat (2000 entry limit)
- [x] Implement automatic eviction policy for large projects
- [x] Maintain backward compatibility with existing APIs

**Implementation**:
```go
type UnifiedCache struct {
    files   map[string]*CachedFile  // Combined file info + content + parsing
    dirs    map[string]*CachedDir   // Directory listings with quick lookup
    maxSize int                     // Configurable memory limits
}
```

**Performance Results**:
- **Code Reduction**: 70% less cache code (400 → 180 lines)
- **Memory Efficiency**: 40% less memory usage (eliminated redundancy)
- **Same Performance**: Maintained 18-27x speed improvements

#### 6. Optimize UI Rendering
**Problem**: Expensive operations in render loops
**Impact**: Improve UI responsiveness and reduce CPU usage
**Files**: `internal/ui/tasks-list/task-item.go`, `internal/ui/tasks-list/tasks-list.go`

**Tasks**:
- [ ] Pre-compute render strings and cache them by task ID
- [ ] Reuse style objects instead of creating new ones in render loop
- [ ] Implement render string pooling for memory efficiency
- [ ] Cache formatted task descriptions and manager indicators
- [ ] Debounce UI updates to reduce render frequency

**Implementation**:
```go
type RenderCache struct {
    taskStrings map[string]string
    styles      map[string]lipgloss.Style
    mutex       sync.RWMutex
}

var (
    // Pre-allocated styles
    itemStyleCache = lipgloss.NewStyle().PaddingLeft(4).Foreground(lipgloss.Color("#CCCCCC"))
    selectedStyleCache = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("39"))
)
```

#### 7. Optimize Lock File Parsing
**Problem**: Inefficient lock file reading and version detection
**Impact**: Faster workspace detection and reduced I/O
**Files**: `internal/manager/js/pnpm.go`, `internal/manager/js/yarn.go`

**Tasks**:
- [ ] Read only first 512 bytes for version detection instead of full files
- [ ] Use compiled regex patterns for version parsing
- [ ] Implement single file scan to detect all package managers
- [ ] Cache lock file analysis results
- [ ] Use memory-mapped files for large lock files

**Implementation**:
```go
var (
    pnpmVersionRegex = regexp.MustCompile(`lockfileVersion:\s*['"]?([^'"\\n]+)['"]?`)
    yarnV2Regex     = regexp.MustCompile(`@npm:`)
)

func detectPackageManager(dir string) (*PackageManagerInfo, error) {
    const headerSize = 512
    
    // Check all lock files in single directory scan
    entries, err := os.ReadDir(dir)
    if err != nil {
        return nil, err
    }
    
    for _, entry := range entries {
        if isLockFile(entry.Name()) {
            info := analyzeLockFile(filepath.Join(dir, entry.Name()), headerSize)
            if info != nil {
                return info, nil
            }
        }
    }
    return nil, nil
}
```

### Low Priority (Long-term Optimization)

#### 8. Implement Hash-based Task Lookup
**Problem**: Linear search for task finding
**Impact**: O(1) task lookup instead of O(n)
**Files**: `internal/manager/manager.go`

**Tasks**:
- [ ] Create task index with hash maps for name and alias lookup
- [ ] Implement incremental index updates when tasks change
- [ ] Add task deduplication logic
- [ ] Cache fuzzy search results with LRU eviction

**Implementation**:
```go
type TaskIndex struct {
    byName    map[string][]ManagerTask
    byAlias   map[string][]ManagerTask
    fuzzyCache map[string][]ManagerTask
    mutex     sync.RWMutex
}
```

#### 9. Worker Pool for Concurrent Processing
**Problem**: Goroutine overhead for small tasks
**Impact**: Better resource utilization and controlled concurrency
**Files**: `internal/manager/parser/parser.go`

**Tasks**:
- [ ] Implement bounded worker pool for directory parsing
- [ ] Add job queue with priority levels
- [ ] Implement graceful shutdown for worker pool
- [ ] Add metrics for worker pool utilization

**Implementation**:
```go
type WorkerPool struct {
    workers    int
    jobQueue   chan Job
    resultChan chan Result
    quit       chan bool
}

type Job struct {
    Directory string
    Index     int
    Workspace *jsmanager.JsWorkspace
}
```

#### 10. Virtual Scrolling for Large Lists
**Problem**: Performance degradation with many tasks
**Impact**: Constant-time rendering regardless of list size
**Files**: `internal/ui/tasks-list/tasks-list.go`

**Tasks**:
- [ ] Implement virtual scrolling for task lists > 100 items
- [ ] Add lazy loading of task descriptions
- [ ] Implement viewport-based rendering
- [ ] Add smooth scrolling animations

#### 11. Advanced Caching Strategy
**Problem**: No persistence of cache across runs
**Impact**: Faster subsequent runs
**Files**: New caching infrastructure

**Tasks**:
- [ ] Implement disk-based cache for parsed configurations
- [ ] Add cache versioning and migration support
- [ ] Implement distributed cache for team environments
- [ ] Add cache warming strategies
- [ ] Implement cache compression for large projects

#### 12. Memory Pool Management
**Problem**: Frequent allocations of similar objects
**Impact**: Reduced GC pressure and memory fragmentation
**Files**: Throughout codebase

**Tasks**:
- [ ] Implement object pools for frequently allocated types (Task, ConfigFile, etc.)
- [ ] Add string interning for repeated file paths and task names
- [ ] Implement buffer pools for file I/O operations
- [ ] Add memory usage monitoring and alerts

### Performance Monitoring

#### 13. Add Performance Metrics
**Tasks**:
- [ ] Add execution time tracking for each optimization
- [ ] Implement memory usage profiling
- [ ] Add I/O operation counting
- [ ] Create performance regression tests
- [ ] Add benchmark suite for critical paths

#### 14. Configuration for Optimizations
**Tasks**:
- [ ] Add configuration options to enable/disable specific optimizations
- [ ] Implement performance profiles (fast, balanced, thorough)
- [ ] Add cache size limits configuration
- [ ] Implement debug mode for optimization analysis

### Success Metrics

**Target Improvements**:
- [ ] 70% reduction in file system operations
- [ ] 50% reduction in memory allocations
- [ ] 60% improvement in startup time for large projects
- [ ] 40% reduction in UI render time
- [ ] 80% reduction in repeated parsing operations

**Measurement Tools**:
- [ ] Go pprof integration for CPU and memory profiling
- [ ] Custom metrics collection for I/O operations
- [ ] Benchmark comparisons before/after each optimization
- [ ] Real-world project testing scenarios 