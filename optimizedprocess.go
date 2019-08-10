package parallel

import (
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

	// The number of routines to remove after optimizing.
	numToRemove int

	// A mutex to protect against simultaneous read/write to numToRemove.
	numToRemoveMutex sync.Mutex

	// A PID controller for controlling the number of goroutines.
	controller *controller

	// A mutex to protect against simultaneous read/write to controller variables.
	controllerMutex sync.Mutex
}

// MARK: Initializers

// NewOptimizedProcess creates and returns a new parallel process with the specified
// number of goroutines.
func NewOptimizedProcess(interval int, gain float64) *OptimizedProcess {
	return &OptimizedProcess{
		optimizationInterval: interval,
		controller:           newController(),
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

// NumRoutines returns the number of routines that the optimized processes is
// currently using.
func (p *OptimizedProcess) NumRoutines() int {
	return p.numRoutines
}

// MARK: Private methods

// reset resets all of the process' properties to their initial state.
func (p *OptimizedProcess) reset() {
	p.numRoutines = 1
	p.count = 0
	p.numToRemove = 0
	p.controller.reset()
}

// runRoutine runs a new routine for the given number of iterations, picking up
// where other routines have left off.
func (p *OptimizedProcess) runRoutine(iterations int, operation Operation) {
	i := p.nextCount()
	for i <= iterations {
		operation(i - 1)

		p.optimizeNumRoutines(i, iterations, operation)

		p.numToRemoveMutex.Lock()
		p.numRoutinesMutex.Lock()
		if p.numToRemove > 0 && p.numRoutines > 1 {
			p.numToRemove--
			p.numRoutines--
			p.numRoutinesMutex.Unlock()
			p.numToRemoveMutex.Unlock()
			break
		} else if p.numToRemove > 0 {
			p.numToRemove--
		}
		p.numRoutinesMutex.Unlock()
		p.numToRemoveMutex.Unlock()

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
		p.numToRemoveMutex.Lock()
		p.numToRemove = c
		p.numToRemoveMutex.Unlock()
		p.group.Done()
	}
}

// nextAction gets the next action the optimize method should use when
// optimizing the number of goroutines.
func (p *OptimizedProcess) nextAction(iteration int) (routineAction, int) {
	p.controllerMutex.Lock()
	defer p.controllerMutex.Unlock()

	r := p.controller.next()
	n := r - p.numRoutines

	if n > 0 {
		return routineActionAdd, n
	} else if n < 0 {
		return routineActionRemove, -n
	}

	return routineActionNone, 0
}
