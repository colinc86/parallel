package parallel

import "sync"

// AlternatingProcess types execute a specified number of operations on a given number of
// goroutines.
type AlternatingProcess struct {
	// The number of goroutines the process should use when divvying up
	// operations.
	numRoutines int

	// The process' wait group to use when waiting for goroutines to
	// finish their execution.
	group sync.WaitGroup
}

// MARK: Initializers

// NewAlternatingProcess creates and returns a new parallel process with the specified
// number of goroutines.
func NewAlternatingProcess(numRoutines int) *AlternatingProcess {
	return &AlternatingProcess{
		numRoutines: numRoutines,
	}
}

// MARK: Public methods

// Execute executes the parallel process for the specified number of operations.
func (p *AlternatingProcess) Execute(iterations int, operation Operation) {
	p.group.Add(p.numRoutines)
	for n := 0; n < p.numRoutines; n++ {
		go p.runRoutine(n, iterations, operation)
	}

	p.group.Wait()
}

// MARK: Private methods

func (p *AlternatingProcess) runRoutine(start int, end int, operation Operation) {
	defer p.group.Done()

	for i := start; i < end; i += p.numRoutines {
		operation(i)
	}
}
