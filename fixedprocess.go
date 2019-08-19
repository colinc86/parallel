package parallel

import "sync"

// FixedProcess types execute a specified number of operations on a given
// number of goroutines.
type FixedProcess struct {
	// The number of goroutines the process should use when divvying up
	// operations.
	numRoutines int

	// The process' wait group to use when waiting for goroutines to finish their
	// execution.
	group sync.WaitGroup

	// The number of iterations in the current execution that have begun.
	count safeInt
}

// MARK: Initializers

// NewFixedProcess creates and returns a new parallel process with the
// specified number of goroutines.
func NewFixedProcess(numRoutines int) *FixedProcess {
	return &FixedProcess{
		numRoutines: numRoutines,
	}
}

// MARK: Public methods

// Execute executes the synced process for the specified number of operations.
func (p *FixedProcess) Execute(iterations int, operation Operation) {
	p.count.set(0)
	p.group.Add(p.numRoutines)
	for n := 0; n < p.numRoutines; n++ {
		go p.runRoutine(iterations, operation)
	}

	p.group.Wait()
}

// NumRoutines returns the number of routines that the synced processes was
// initialized with.
func (p *FixedProcess) NumRoutines() int {
	return p.numRoutines
}

// MARK: Private methods

func (p *FixedProcess) runRoutine(iterations int, operation Operation) {
	defer p.group.Done()

	i := p.count.get()
	for i < iterations {
		operation(i)
		i = p.count.add(1)
	}
}
