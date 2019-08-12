package parallel

import "sync"

// OptimizedProcess types execute a specified number of operations on a variable
// number of goroutines.
type OptimizedProcess struct {
	// The number of iterations between optimizations.
	optimizationInterval int

	// The process' wait group to use when waiting for goroutines to
	// finish their execution.
	group sync.WaitGroup

	// The number of goroutines the process should use when divvying up
	// operations.
	numRoutines safeInt

	// The number of iterations in the current execution that have begun.
	count safeInt

	// The number of routines to remove after optimizing.
	numToRemove safeInt

	// A PID controller for controlling the number of goroutines.
	controller *controller

	// A mutex to protect against simultaneous read/write to controller variables.
	controllerMutex sync.Mutex
}

// MARK: Initializers

// NewOptimizedProcess creates and returns a new parallel process with the
// specified optimization interval.
func NewOptimizedProcess(interval int) *OptimizedProcess {
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
	return p.numRoutines.get()
}

// MARK: Private methods

// reset resets all of the process' properties to their initial state.
func (p *OptimizedProcess) reset() {
	p.numRoutines.set(1)
	p.count.set(0)
	p.numToRemove.set(0)
	p.controller.reset()
}

// runRoutine runs a new routine for the given number of iterations, picking up
// where other routines have left off.
func (p *OptimizedProcess) runRoutine(iterations int, operation Operation) {
	i := p.count.get()
	for i < iterations {
		operation(i)

		p.optimizeNumRoutines(i, iterations, operation)

		// Lock/unlock together to keep vars in sync
		p.numToRemove.mutex.Lock()
		p.numRoutines.mutex.Lock()
		if p.numToRemove.value > 0 && p.numRoutines.value > 1 {
			p.numToRemove.value--
			p.numRoutines.value--
			p.numToRemove.mutex.Unlock()
			p.numRoutines.mutex.Unlock()
			break
		} else if p.numToRemove.value > 0 {
			p.numToRemove.value--
		}
		p.numRoutines.mutex.Unlock()
		p.numToRemove.mutex.Unlock()

		i = p.count.add(1)
	}

	p.group.Done()
}

// optimizeNumRoutines optimized the number of routines to use for the parallel
// operation.
func (p *OptimizedProcess) optimizeNumRoutines(iteration int, iterations int, operation Operation) {
	if iteration%p.optimizationInterval != 0 || iteration == 0 {
		return
	}

	p.group.Add(1)
	n := p.nextAction(iteration)

	if n == 0 {
		p.group.Done()
	} else if n > 0 {
		if n > 1 {
			p.group.Add(n - 1)
		}

		p.numRoutines.add(n)
		for i := 0; i < n; i++ {
			go p.runRoutine(iterations, operation)
		}
	} else if n < 0 {
		p.numToRemove.set(n)
		p.group.Done()
	}
}

// nextAction gets the next action the optimize method should use when
// optimizing the number of goroutines.
func (p *OptimizedProcess) nextAction(iteration int) int {
	p.controllerMutex.Lock()
	defer p.controllerMutex.Unlock()

	return p.controller.next() - p.numRoutines.get()
}
