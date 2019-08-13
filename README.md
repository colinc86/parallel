# parallel
A parallel processing package for Go.

## Examples
The two structures that implement the `Process` interface, `SyncedProcess` and `OptimizedProcess`, both perform a set of parallel operations on either a fixed or varying number of goroutines.

### SyncedProcess
`SyncedProcess` types execute their set of operations on a fixed number of goroutines specified upon initialization.

```go
// Create a process with two goroutines.
p := parallel.NewSyncedProcess(2)

// Execute 100 operations in parallel.
p.Execute(100, func (i int) {
  // Perform the ith operation.
})
```

### OptimizedProcess
`OptimizedProcess` types execute their set of operations on a variable number of goroutines by utilizing a PID control loop to maximize CPU throughput.

```go
// Create an optimized process with optimiztaion interval
// and PID coefficients.
p := parallel.NewOptimizedProcess(500 * time.Millisecond, 3.0, 0.0, 1.0)

// Execute 100 operations in parallel.
p.Execute(100, func(i int) {
  // Perform the ith operation.
})
```
