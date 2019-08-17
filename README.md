# parallel
A parallel processing package for Go.

## Usage
Two structures implement the `Process` interface; `SyncedProcess` and `OptimizedProcess`. Both perform a set of parallel operations on either a fixed or varying number of goroutines.

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
`OptimizedProcess` types execute their set of operations on a variable number of goroutines by utilizing a PID control loop to maximize CPU throughput. You configure the PID controller by creating a `ControllerConfiguration` struct and passing to the `NewOptimizedProcess` function.

```go
// Create an optimized process with optimiztaion interval, max goroutine count,
// and PID controller configuration.
c := NewControllerConfiguration(2.0, 0.0, 1.0, 0.1, 1.0)
p := parallel.NewOptimizedProcess(500 * time.Millisecond, 2 * runtime.NumCPU(), c)

// Execute 100 operations in parallel.
p.Execute(100, func(i int) {
  // Perform the ith operation.
})
```

You can't change the number of goroutines directly in an optimized process, but you can modify its optimization parameters while it's executing.

```go
// Execute the process on its own goroutine
go p.Execute(100, func(i int) {
  // Perform the ith operation.
})

// Modify optimization parameters while the process is executing...
p.SetOptimizationInterval(time.Second)
p.SetOptimizationCoefficients(1.0, 0.01, 1.0)
p.SetMaxRoutines(50)
```
