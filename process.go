package parallel

import "sync"

// Operation types represent a single operation in a parallel process.
type Operation func(i int)

type finishedProcess struct{}

// Process types execute a specified number of operations on a given number of
// goroutines.
type Process struct {
	// The process' channel to be called when all operations complete.
	C chan struct{}

	// The number of goroutines the process should use when divvying up
	// operations.
	numRoutines int

	// The process' wait group to use when waiting for goroutines to
	// finish their execution.
	group sync.WaitGroup
}

// MARK: Initializers

// NewProcess creates and returns a new parallel process with the specified
// number of goroutines.
func NewProcess(numRoutines int) *Process {
	return &Process{
		C:           make(chan struct{}, 1),
		numRoutines: numRoutines,
	}
}

// MARK: Public methods

// Execute executes the parallel process for the specified number of operations.
func (p *Process) Execute(iterations int, operation Operation) {
	p.group.Add(p.numRoutines)
	for n := 0; n < p.numRoutines; n++ {
		go p.runRoutine(n, iterations, operation)
	}

	p.group.Wait()
	p.C <- finishedProcess{}
}

// MARK: Private methods

func (p *Process) runRoutine(start int, end int, operation Operation) {
	defer p.group.Done()

	for i := start; i < end; i += p.numRoutines {
		operation(i)
	}
}
