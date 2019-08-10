package parallel

import "sync"

// SyncedProcess types execute a specified number of operations on a given number of
// goroutines.
type SyncedProcess struct {
	// The number of goroutines the process should use when divvying up
	// operations.
	numRoutines int

	// The process' wait group to use when waiting for goroutines to
	// finish their execution.
	group sync.WaitGroup

	// The number of iterations in the current execution that have begun.
	count int

	// A mutex to protect against simultaneous read/writes to count.
	countMutex sync.Mutex
}

// MARK: Initializers

// NewSyncedProcess creates and returns a new parallel process with the specified
// number of goroutines.
func NewSyncedProcess(numRoutines int) *SyncedProcess {
	return &SyncedProcess{
		numRoutines: numRoutines,
	}
}

// MARK: Public methods

// Execute executes the parallel process for the specified number of operations.
func (p *SyncedProcess) Execute(iterations int, operation Operation) {
	p.count = 0
	p.group.Add(p.numRoutines)
	for n := 0; n < p.numRoutines; n++ {
		go p.runRoutine(iterations, operation)
	}

	p.group.Wait()
}

// NumRoutines returns the number of routines that the synced processes was
// initialized with.
func (p *SyncedProcess) NumRoutines() int {
	return p.numRoutines
}

// MARK: Private methods

func (p *SyncedProcess) runRoutine(iterations int, operation Operation) {
	defer p.group.Done()

	i := p.nextCount()
	for i <= iterations {
		operation(i - 1)
		i = p.nextCount()
	}
}

func (p *SyncedProcess) nextCount() int {
	p.countMutex.Lock()
	defer p.countMutex.Unlock()

	p.count++
	return p.count
}
