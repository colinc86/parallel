package parallel

import "testing"
import "time"

func TestReporterReset(t *testing.T) {
	r := newReporter()
	nowTime := time.Now()

	if r.lastTime.After(nowTime) || r.lastTime.Equal(time.Unix(0, 0)) {
		t.Errorf("The initial report time should be non-zero and less than the current time.")
	}

	r.reset()

	nowTime = time.Now()

	if r.lastTime.After(nowTime) || r.lastTime.Equal(time.Unix(0, 0)) {
		t.Errorf("The last report time should be non-zero and less than the current time.")
	}
}

func TestReporterUsage(t *testing.T) {
	r := newReporter()
	u := r.usage()

	if u <= 0.0 {
		t.Errorf("CPU usage, %f, should be greater than 0.0.", u)
	}
}
