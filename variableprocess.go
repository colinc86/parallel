package parallel

import (
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/colinc86/probes"
)

// VariableProcess types execute a specified number of operations on a variable
// number of goroutines.
type VariableProcess struct {
	// The number of iterations between optimizations.
	optimizationInterval time.Duration

	// The process' wait group to use when waiting for goroutines to
	// finish their execution.
	group sync.WaitGroup

	// The ticker responsible for triggering an optimization.
	ticker *time.Ticker

	// The number of goroutines the process should use when divvying up
	// operations.
	numRoutines int64

	// The maximum number of goroutines to use when optimizing.
	maxRoutines safeInt

	// The number of iterations in the current execution that have begun.
	iteration safeInt

	// The total number of iterations specified by the last call to Execute.
	iterations int

	// The operation function called for each iteration of the process.
	operation Operation

	// The number of routines to remove after optimizing.
	numToRemove int64

	// The CPU reporter used to calculate CPU throughput.
	reporter *reporter

	// A PID controller for controlling the number of goroutines.
	controller *controller

	// A mutex to protect against simultaneous read/write to controller variables.
	controllerMutex sync.Mutex

	// Whether or not the controller should be probed.
	probeController bool

	// The CPU probe.
	CPUProbe *probes.Probe

	// The error probe.
	ErrorProbe *probes.Probe

	// The PID output probe.
	PIDProbe *probes.Probe

	// The routine probe.
	RoutineProbe *probes.Probe
}

// MARK: Initializers

// NewVariableProcess creates and returns a new parallel process with the
// specified optimization interval.
func NewVariableProcess(interval time.Duration, maxRoutines int, controllerConfiguration *ControllerConfiguration, probeController bool) *VariableProcess {
	p := &VariableProcess{
		optimizationInterval: interval,
		maxRoutines:          safeInt{value: maxRoutines},
		reporter:             newReporter(),
		controller:           newController(controllerConfiguration),
		probeController:      probeController,
	}

	if probeController {
		p.CPUProbe = probes.NewProbe()
		p.ErrorProbe = probes.NewProbe()
		p.PIDProbe = probes.NewProbe()
		p.RoutineProbe = probes.NewProbe()
	}

	return p
}

// MARK: Public methods

// Execute executes the parallel process for the specified number of operations
// while optimizing every interval iterations.
func (p *VariableProcess) Execute(iterations int, operation Operation) {
	if p.probeController {
		p.CPUProbe.Activate()
		p.ErrorProbe.Activate()
		p.PIDProbe.Activate()
		p.RoutineProbe.Activate()
	}

	p.iterations = iterations
	p.operation = operation
	p.reset()
	p.group.Add(1)

	go p.runRoutine()
	go p.beginOptimizing()

	p.group.Wait()
	p.ticker.Stop()

	if p.probeController {
		p.CPUProbe.Flush()
		p.ErrorProbe.Flush()
		p.PIDProbe.Flush()
		p.RoutineProbe.Flush()

		p.CPUProbe.Deactivate()
		p.ErrorProbe.Deactivate()
		p.PIDProbe.Deactivate()
		p.RoutineProbe.Deactivate()
	}
}

// NumRoutines returns the number of routines that the variable processes is
// currently using.
func (p *VariableProcess) NumRoutines() int {
	return int(atomic.LoadInt64(&p.numRoutines))
}

// GetOptimizationInterval returns the interval of the process' ticker.
func (p *VariableProcess) GetOptimizationInterval() time.Duration {
	return p.optimizationInterval
}

// SetOptimizationInterval sets the optimization interval and restarts the
// process' ticker.
func (p *VariableProcess) SetOptimizationInterval(interval time.Duration) {
	p.ticker.Stop()
	p.optimizationInterval = interval
	go p.beginOptimizing()
}

// GetMaxRoutines returns the maximum number of goroutines to use when
// optimizing.
func (p *VariableProcess) GetMaxRoutines() int {
	return p.maxRoutines.get()
}

// SetMaxRoutines sets the maximum number of goroutines to use when optimizing.
// Must be greater than 0.
func (p *VariableProcess) SetMaxRoutines(n int) {
	p.maxRoutines.set(n)
}

// GetControllerConfiguration gets the PID controller configuration.
func (p *VariableProcess) GetControllerConfiguration() *ControllerConfiguration {
	p.controllerMutex.Lock()
	defer p.controllerMutex.Unlock()
	return p.controller.configuration.Copy()
}

// SetControllerConfiguration sets the PID controller coefficients.
func (p *VariableProcess) SetControllerConfiguration(configuration *ControllerConfiguration) {
	p.controllerMutex.Lock()
	defer p.controllerMutex.Unlock()

	p.controller.configuration = configuration
}

// MARK: Private methods

// reset resets all of the process' properties to their initial state.
func (p *VariableProcess) reset() {
	if p.probeController {
		p.PIDProbe.ClearSignal()
		p.CPUProbe.ClearSignal()
		p.ErrorProbe.ClearSignal()
		p.RoutineProbe.ClearSignal()
	}

	p.numRoutines = 1
	p.iteration.set(0)
	p.numToRemove = 0
	p.controller.reset()
	p.reporter.reset()
}

// beginOptimizing begins optimizing by calling optimizeNumRoutines each time
// the process' ticker fires.
func (p *VariableProcess) beginOptimizing() {
	p.ticker = time.NewTicker(p.optimizationInterval)
	for range p.ticker.C {
		p.optimizeNumRoutines()
	}
}

// runRoutine runs a new routine for the given number of iterations, picking up
// where other routines have left off.
func (p *VariableProcess) runRoutine() {
	i := p.iteration.get()
	for i < p.iterations {
		p.operation(i)

		n := atomic.LoadInt64(&p.numToRemove)
		if n > 0 && atomic.LoadInt64(&p.numRoutines) > 1 {
			atomic.AddInt64(&p.numToRemove, -1)
			atomic.AddInt64(&p.numRoutines, -1)
			break
		} else if n > 0 {
			atomic.AddInt64(&p.numToRemove, -1)
		}

		i = p.iteration.add(1)
	}

	p.group.Done()
}

// optimizeNumRoutines variable the number of routines to use for the parallel
// operation.
func (p *VariableProcess) optimizeNumRoutines() {
	p.group.Add(1)

	p.controllerMutex.Lock()
	usage := p.reporter.usage()
	u, e := p.controller.next(usage)
	p.controllerMutex.Unlock()

	m := int(math.Ceil(u))
	p.maxRoutines.mutex.Lock()
	if m > p.maxRoutines.value {
		m = p.maxRoutines.value
	}
	p.maxRoutines.mutex.Unlock()

	routines := int(atomic.LoadInt64(&p.numRoutines))
	n := m - routines

	if p.probeController {
		p.CPUProbe.C <- usage
		p.PIDProbe.C <- u
		p.ErrorProbe.C <- e
		p.RoutineProbe.C <- float64(m)
	}

	if n == 0 {
		p.group.Done()
	} else if n > 0 {
		if n > 1 {
			p.group.Add(n - 1)
		}

		atomic.AddInt64(&p.numRoutines, int64(n))

		for i := 0; i < n; i++ {
			go p.runRoutine()
		}
	} else if n < 0 {
		if routines > 1 {
			atomic.StoreInt64(&p.numToRemove, -1*int64(n))
		}
		p.group.Done()
	}
}
