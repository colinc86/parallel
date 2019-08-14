package parallel

import (
	"sync"
	"time"
)

// OptimizedProcess types execute a specified number of operations on a variable
// number of goroutines.
type OptimizedProcess struct {
	// The number of iterations between optimizations.
	optimizationInterval time.Duration

	// The process' wait group to use when waiting for goroutines to
	// finish their execution.
	group sync.WaitGroup

	// The ticker responsible for triggering an optimization.
	ticker *time.Ticker

	// Set to true when the ticker fires.
	needsOptimization bool

	// A mutex to provide read/write lock to needsOptimization.
	needsOptimizationMutex sync.Mutex

	// The number of goroutines the process should use when divvying up
	// operations.
	numRoutines safeInt

	// The maximum number of goroutines to use when optimizing.
	maxRoutines safeInt

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
func NewOptimizedProcess(interval time.Duration, maxRoutines int, kp float64, ki float64, kd float64) *OptimizedProcess {
	return &OptimizedProcess{
		optimizationInterval: interval,
		maxRoutines:          safeInt{value: maxRoutines},
		controller:           newController(kp, ki, kd),
	}
}

// MARK: Public methods

// Execute executes the parallel process for the specified number of operations
// while optimizing every interval iterations.
func (p *OptimizedProcess) Execute(iterations int, operation Operation) {
	p.reset()
	p.group.Add(1)

	go p.runRoutine(iterations, operation)
	go p.beginOptimizing()

	p.group.Wait()
	p.ticker.Stop()
}

// NumRoutines returns the number of routines that the optimized processes is
// currently using.
func (p *OptimizedProcess) NumRoutines() int {
	return p.numRoutines.get()
}

// GetOptimizationInterval returns the interval of the process' ticker.
func (p *OptimizedProcess) GetOptimizationInterval() time.Duration {
	return p.optimizationInterval
}

// SetOptimizationInterval sets the optimization interval and restarts the
// process' ticker.
func (p *OptimizedProcess) SetOptimizationInterval(interval time.Duration) {
	p.ticker.Stop()
	p.optimizationInterval = interval
	go p.beginOptimizing()
}

// GetMaxRoutines returns the maximum number of goroutins to use when
// optimizing.
func (p *OptimizedProcess) GetMaxRoutines() int {
	return p.maxRoutines.get()
}

// SetMaxRoutines sets the maximum number of goroutines to use when optimizing.
// Must be greater than 0.
func (p *OptimizedProcess) SetMaxRoutines(n int) {
	p.maxRoutines.set(n)
}

// GetOptimizationParameters gets the PID controller coefficients.
func (p *OptimizedProcess) GetOptimizationParameters() (float64, float64, float64) {
	p.controllerMutex.Lock()
	defer p.controllerMutex.Unlock()
	return p.controller.kp, p.controller.ki, p.controller.kd
}

// SetOptimizationParameters sets the PID controller coefficients.
func (p *OptimizedProcess) SetOptimizationParameters(kp float64, ki float64, kd float64) {
	p.controllerMutex.Lock()
	defer p.controllerMutex.Unlock()

	p.controller.kp = kp
	p.controller.ki = ki
	p.controller.kd = kd
}

// MARK: Private methods

// reset resets all of the process' properties to their initial state.
func (p *OptimizedProcess) reset() {
	p.needsOptimization = false
	p.numRoutines.set(1)
	p.count.set(0)
	p.numToRemove.set(0)
	p.controller.reset()
}

func (p *OptimizedProcess) beginOptimizing() {
	p.ticker = time.NewTicker(p.optimizationInterval)
	for range p.ticker.C {
		p.needsOptimizationMutex.Lock()
		p.needsOptimization = true
		p.needsOptimizationMutex.Unlock()
	}
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
	p.needsOptimizationMutex.Lock()
	if !p.needsOptimization {
		p.needsOptimizationMutex.Unlock()
		return
	}
	p.needsOptimization = false
	p.needsOptimizationMutex.Unlock()

	p.group.Add(1)
	n := p.nextAction(iteration)
	p.maxRoutines.mutex.Lock()
	if n > p.maxRoutines.value {
		n = p.maxRoutines.value
	}
	p.maxRoutines.mutex.Unlock()
	n -= p.numRoutines.get()

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
