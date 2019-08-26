# parallel
[![GoDoc](https://godoc.org/github.com/colinc86/parallel?status.svg)](https://godoc.org/github.com/colinc86/parallel)

A parallel processing package for Go.

## Usage
Two structures implement the `Process` interface; `FixedProcess` and `VariableProcess`. Both perform a set of parallel operations on either a fixed or varying number of goroutines.

### FixedProcess
`FixedProcess` types execute their set of operations on a fixed number of goroutines specified upon initialization.

```go
// Create a process with two goroutines.
p := parallel.NewFixedProcess(2)

// Execute 100 operations in parallel.
p.Execute(100, func (i int) {
  // Perform the ith operation.
})
```

### VariableProcess
`VariableProcess` types execute their set of operations on a variable number of goroutines by utilizing a PID control loop to maximize CPU throughput. You configure the PID controller by creating a `ControllerConfiguration` struct and passing to the `NewVariableProcess` function.

```go
// Create a variable process with optimization interval, max goroutine count,
// and PID controller configuration.
c := NewControllerConfiguration(2.0, 0.0, 1.0, 0.1, 1.0)
p := parallel.NewVariableProcess(500 * time.Millisecond, 2 * runtime.NumCPU(), c, false)

// Execute 100 operations in parallel.
p.Execute(100, func(i int) {
  // Perform the ith operation.
})
```

You can't change the number of goroutines directly in an variable process, but you can modify its optimization parameters while it's executing.

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
A variable process has four [probes](https://www.github.com/colinc86/probes) that are activated by setting the `probeController` parameter to true upon initialization. They monitor and keep a record of the following values.
- CPU throughput
- PID input error
- PID output signal
- Number of goroutines

```go
// Create a variable process with probeController set to true.
c := NewControllerConfiguration(2.0, 0.0, 1.0, 0.1, 1.0)
p := parallel.NewVariableProcess(500 * time.Millisecond, 2 * runtime.NumCPU(), c, true)

// Execute 100 operations in parallel.
p.Execute(100, func(i int) {
  // Perform the ith operation.
})

// Examine the PID output signal programmatically
s := p.PIDProbe.Signal()
```

### Stopping a Process
A process can be stopped at any time by calling the `Stop()` method. The process will stop after any operations that have already begun finish executing.

```go
p.Execute(100, func(i int) {
  // Perform the ith operation with an error condition...
  if err != nil {
    p.Stop()
  }
})
