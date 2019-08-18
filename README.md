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
// Create an optimized process with optimization interval, max goroutine count,
// and PID controller configuration.
c := NewControllerConfiguration(2.0, 0.0, 1.0, 0.1, 1.0)
p := parallel.NewOptimizedProcess(500 * time.Millisecond, 2 * runtime.NumCPU(), c, false)

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
p.SetMaxRoutines(50)

c := p.GetControllerConfiguration()
// Modify config...
p.SetControllerConfiguration(c)
```

#### Tuning the PID Controller
An optimized process has four [probes](https://www.github.com/colinc86/probes) that are activated by setting the `probeController` parameter to true upon initialization. They monitor and keep a record of the following values.
- CPU throughput
- PID input error
- PID output signal
- Number of goroutines

Examine any of the probes programmatically, or by saving a plot of the probe's signal to file.

```go
// Create an optimized process with probeController set to true.
c := NewControllerConfiguration(2.0, 0.0, 1.0, 0.1, 1.0)
p := parallel.NewOptimizedProcess(500 * time.Millisecond, 2 * runtime.NumCPU(), c, true)

// Execute 100 operations in parallel.
p.Execute(100, func(i int) {
  // Perform the ith operation.
})

// Examine the PID output signal programmatically
s := p.PIDProbe.Signal()

// Or save a plot of the signal to file
p.PIDProbe.WriteSignalToPNG("pid")
```
