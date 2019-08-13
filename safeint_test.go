package parallel

import "testing"

func TestGetSafeIntValue(t *testing.T) {
	var s safeInt
	if s.get() != 0 {
		t.Errorf("Value, %d, should be 0.", s.value)
	}
}

func TestSetSafeIntValue(t *testing.T) {
	var s safeInt
	s.set(1)

	if s.value != 1 {
		t.Errorf("Value, %d, should be 1.", s.value)
	}
}

func TestAddSafeIntValue(t *testing.T) {
	var s safeInt
	s.add(2)

	if s.value != 2 {
		t.Errorf("Value, %d, should be 2.", s.value)
	}
}

func TestSubtractSafeIntValue(t *testing.T) {
	var s safeInt
	s.subtract(2)

	if s.value != -2 {
		t.Errorf("Value, %d, should be -2.", s.value)
	}
}
