package parallel

import (
	"math"
	"runtime"
	"sync"
)

// OptimizedProcess types execute a specified number of operations on a given number of
// goroutines.
type OptimizedProcess struct {
	// The number of iterations between optimizations.
	optimizationInterval int

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

	// The total error used to calculated the integral of the PID control loop.
	totalError float64

	// The number of routines to remove after optimizing.
	removeCount int

	// A mutex to protect against simultaneous read/write to removeCount.
	removeCountMutex sync.Mutex

	// The number of CPUs available to the runtime.
	cpuCount int

	// Reporter to collect CPU usage.
	reporter *Reporter

	// A mutex to protect against simultaneous read/write when calculating a CPU
	// usage report.
	reportMutex sync.Mutex
}

// MARK: Initializers

// NewOptimizedProcess creates and returns a new parallel process with the specified
// number of goroutines.
func NewOptimizedProcess(interval int, gain float64) *OptimizedProcess {
	return &OptimizedProcess{
		optimizationInterval: interval,
		cpuCount:             runtime.NumCPU(),
		reporter:             NewReporter(),
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
	p.removeCount = 0
	p.previousError = 0.0
	p.totalError = 0.0
	p.reporter.Reset()
}

// runRoutine runs a new routine for the given number of iterations, picking up
// where other routines have left off.
func (p *OptimizedProcess) runRoutine(iterations int, operation Operation) {
	i := p.nextCount()
	for i <= iterations {
		operation(i - 1)

		p.optimizeNumRoutines(i, iterations, operation)

		p.removeCountMutex.Lock()
		p.numRoutinesMutex.Lock()
		if p.removeCount > 0 && p.numRoutines > 1 {
			p.removeCount--
			p.numRoutines--
			p.numRoutinesMutex.Unlock()
			p.removeCountMutex.Unlock()
			break
		} else if p.removeCount > 0 {
			p.removeCount--
		}
		p.numRoutinesMutex.Unlock()
		p.removeCountMutex.Unlock()

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
func (p *OptimizedProcess) optimizeNumRoutines(iteration int, iterations int, operation Operation) {
	if iteration%p.optimizationInterval != 0 || iteration == 0 {
		return
	}

	p.group.Add(1)
	nextAction, c := p.nextAction(iteration)

	switch nextAction {
	case routineActionNone:
		p.group.Done()
	case routineActionAdd:
		if c == 0 {
			p.group.Done()
		} else if c > 0 {
			if c > 1 {
				p.group.Add(c - 1)
			}

			for i := 0; i < c; i++ {
				p.numRoutinesMutex.Lock()
				p.numRoutines++
				p.numRoutinesMutex.Unlock()
				go p.runRoutine(iterations, operation)
			}
		}
	case routineActionRemove:
		p.removeCountMutex.Lock()
		p.removeCount = c
		p.removeCountMutex.Unlock()
		p.group.Done()
	}
}

// nextAction gets the next action the optimize method should use when
// optimizing the number of goroutines.
func (p *OptimizedProcess) nextAction(iteration int) (routineAction, int) {
	p.reportMutex.Lock()
	defer p.reportMutex.Unlock()

	usage, t := p.reporter.Usage()

	P := 1.0 - (usage / float64(p.cpuCount))
	P = (0.01 * P) + (0.99 * p.previousError)
	I := p.totalError + P
	D := (P - p.previousError) / t
	U := 20.0*P + 0.1*I - 3.0*D
	N := int(math.Round(U)) - p.numRoutines

	p.previousError = P
	p.totalError += P

	usage /= float64(p.cpuCount)
	if N > 0 {
		return routineActionAdd, N
	} else if N < 0 {
		return routineActionRemove, -N
	}

	return routineActionNone, 0
}
