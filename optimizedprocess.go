package parallel

import (
	"sync"
	"time"
)

// OptimizedProcess types execute a specified number of operations on a given number of
// goroutines.
type OptimizedProcess struct {
	// The number of iterations between optimizations.
	optimizationInterval int

	// The cutoff threshold for adding or removing goroutines.
	optimizationThreshold float64

	// The number of goroutines the process should use when divvying up
	// operations.
	numRoutines int

	// A mutex to protect against sumultaneous read/write to numRoutines.
	numRoutinesMutex sync.Mutex

	// The process' wait group to use when waiting for goroutines to
	// finish their execution.
	group sync.WaitGroup

	// The number of iterations in the current execution that have begun.
	count int

	// A mutex to protect against simultaneous read/write to count.
	countMutex sync.Mutex

	// The last error collected during optimization.
	previousError float64

	// The time of the last report collected during optimization.
	lastReportTime time.Time

	// A mutex to protect against sumultaneous read/write to lastReport.
	reportMutex sync.Mutex
}

// MARK: Initializers

// NewOptimizedProcess creates and returns a new parallel process with the specified
// number of goroutines.
func NewOptimizedProcess(interval int, threshold float64) *OptimizedProcess {
	return &OptimizedProcess{
		optimizationInterval:  interval,
		optimizationThreshold: threshold,
	}
}

// MARK: Public methods

// Execute executes the parallel process for the specified number of operations
// while optimizing every interval iterations.
func (p *OptimizedProcess) Execute(iterations int, operation Operation) {
	p.reset()
	p.group.Add(1)

	go p.runRoutine(iterations, operation)

	p.group.Wait()
}

// MARK: Private methods

// reset resets all of the process' properties to their initial state.
func (p *OptimizedProcess) reset() {
	p.numRoutines = 1
	p.count = 0
	p.previousError = 0.0
	p.lastReportTime = time.Now()
}

// runRoutine runs a new routine for the given number of iterations, picking up
// where other routines have left off.
func (p *OptimizedProcess) runRoutine(iterations int, operation Operation) {
	i := p.nextCount()
	for i < iterations {
		operation(i)

		if p.optimizeNumRoutines(i, iterations, operation) {
			break
		}

		i = p.nextCount()
	}

	p.group.Done()
}

// nextCount retrieves the next operation iteration to perform.
func (p *OptimizedProcess) nextCount() int {
	p.countMutex.Lock()
	defer p.countMutex.Unlock()

	p.count++
	return p.count
}

// optimizeNumRoutines optimized the number of routines to use for the parallel
// operation.
func (p *OptimizedProcess) optimizeNumRoutines(iteration int, iterations int, operation Operation) bool {
	if iteration%p.optimizationInterval != 0 || iteration == 0 {
		return false
	}

	p.group.Add(1)

	switch p.nextAction(iteration) {
	case routineActionNone:
		p.group.Done()
	case routineActionAdd:
		p.numRoutinesMutex.Lock()
		p.numRoutines++
		p.numRoutinesMutex.Unlock()
		go p.runRoutine(iterations, operation)
	case routineActionRemove:
		if p.numRoutines <= 1 {
			p.group.Done()
			break
		}

		p.numRoutinesMutex.Lock()
		p.numRoutines--
		p.numRoutinesMutex.Unlock()
		p.group.Done()
		return true
	}

	return false
}

// nextAction gets the next action the optimize method should use when
// optimizing the number of goroutines.
func (p *OptimizedProcess) nextAction(iteration int) routineAction {
	p.reportMutex.Lock()
	defer p.reportMutex.Unlock()

	now := time.Now()
	E := now.Sub(p.lastReportTime).Seconds()
	p.lastReportTime = now

	if iteration == p.optimizationInterval {
		p.previousError = E
		return routineActionNone
	}

	diff := (p.previousError - E) / p.previousError
	p.previousError = E

	if diff > p.optimizationThreshold {
		return routineActionAdd
	} else if diff < -p.optimizationThreshold {
		return routineActionRemove
	}

	return routineActionNone
}
