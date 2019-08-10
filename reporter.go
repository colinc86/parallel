package parallel

//#include <time.h>
import "C"
import "time"

// reporter types report the amount of CPU usage between the current and last
// call to the usage method.
type reporter struct {
	lastTime time.Time
	lastTick C.clock_t
}

// MARK: Initializers

// newReporter creates and returns a new CPU reporter.
func newReporter() *reporter {
	return &Reporter {
		lastTime: time.Now(),
		lastTick: C.clock(),
	}
}

// MARK: Public methods

// usage returns the decimal percent of CPU usage used by the process. If this
// is the first time to call this method, then the usage reported will be
// calculated between this call and the last call to reset (or instantiation).
func (r *reporter) usage() (float64, float64) {
	nowClock := C.clock()
	nowActual := time.Now()
	
	clockSeconds := float64(nowClock - r.lastTick) / float64(C.CLOCKS_PER_SEC)
	r.lastTick = nowClock

	actualSeconds := nowActual.Sub(r.lastTime).Seconds()
	r.lastTime = nowActual

	return clockSeconds / actualSeconds, actualSeconds
}

// Reset resets the reporter's last time and tick.
func (r *reporter) reset() {
	r.lastTime = time.Now()
	r.lastTick = C.clock()
}
