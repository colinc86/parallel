package parallel

import "sync"

// safeInt wraps an integer type and exposes methods to safely read/write to the
// integer value from multiple threads.
type safeInt struct {
	value int
	mutex sync.Mutex
}

// get gets the integer value.
func (s *safeInt) get() int {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.value
}

// set sets the integer value and returns the result.
func (s *safeInt) set(n int) int {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.value = n
	return s.value
}

// add adds the input parameter to the integer value and returns the result.
func (s *safeInt) add(n int) int {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.value += n
	return s.value
}

// subtract subtracts the input parameter from the integer value and returns the
// result.
func (s *safeInt) subtract(n int) int {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.value -= n
	return s.value
}
