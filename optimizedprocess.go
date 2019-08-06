package parallel

import (
	"log"
	"sync"
	"time"
)

type routineAction int

const (
	routineActionNone   routineAction = 0
	routineActionAdd    routineAction = 1
	routineActionRemove routineAction = 2
)

// OptimizedProcess types execute a specified number of operations on a given number of
// goroutines.
type OptimizedProcess struct {
	// The number of iterations between optimizations.
	OptimizationInterval int

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

	// The last report collected during optimization.
	lastReport float64

	// The time of the last report collected during optimization.
	lastReportTime time.Time

	// A mutex to protect against sumultaneous read/write to lastReport.
	reportMutex sync.Mutex
}

// MARK: Initializers

// NewOptimizedProcess creates and returns a new parallel process with the specified
// number of goroutines.
func NewOptimizedProcess(interval int) *OptimizedProcess {
	return &OptimizedProcess{
		OptimizationInterval: interval,
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
	log.Println(rCount)
}

// MARK: Private methods

// reset resets all of the process' properties to their initial state.
func (p *OptimizedProcess) reset() {
	totalError = 0.0
	previousError = 0.0
	baselineDelta = 0.0
	added = 0
	removed = 0
	max = 0
	rCount = nil
	changed = false
	wasPositive = true

	p.numRoutines = 1
	p.count = 0
	p.lastReport = 0.0
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
	if iteration%p.OptimizationInterval != 0 || iteration == 0 {
		return false
	}

	defer p.updateRCount()
	p.group.Add(1)

	switch p.nextAction(iteration) {
	case routineActionNone:
		p.group.Done()
	case routineActionAdd:
		added++
		p.numRoutinesMutex.Lock()
		p.numRoutines++
		if p.numRoutines > max {
			max = p.numRoutines
		}
		p.numRoutinesMutex.Unlock()
		go p.runRoutine(iterations, operation)
	case routineActionRemove:
		if p.numRoutines <= 1 {
			p.group.Done()
			break
		}

		removed++
		p.numRoutinesMutex.Lock()
		p.numRoutines--
		p.numRoutinesMutex.Unlock()
		p.group.Done()
		return true
	}

	return false
}

func (p *OptimizedProcess) updateRCount() {
	p.numRoutinesMutex.Lock()
	rCount = append(rCount, p.numRoutines)
	p.numRoutinesMutex.Unlock()
}

var totalError float64
var previousError float64
var baselineDelta float64
var previousDelta float64

var added int
var removed int
var max int
var rCount []int
var changed bool
var wasPositive bool

// nextAction gets the next action the optimize method should use when
// optimizing the number of goroutines to use.
func (p *OptimizedProcess) nextAction(iteration int) routineAction {
	p.reportMutex.Lock()
	defer p.reportMutex.Unlock()

	now := time.Now()
	delta := now.Sub(p.lastReportTime).Seconds()
	p.lastReportTime = now

	if iteration == p.OptimizationInterval {
		previousDelta = delta
		return routineActionAdd
	}

	e := delta - previousDelta
	log.Println(e)
	cutoff := 0.01
	if e < -cutoff {
		return routineActionAdd
	} else if e > cutoff {
		return routineActionRemove
	}

	return routineActionNone
}
