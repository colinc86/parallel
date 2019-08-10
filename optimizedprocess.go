package parallel

//#include <time.h>
import "C"

import (
	"fmt"
	"math"
	"runtime"
	"sync"
	"time"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

// OptimizedProcess types execute a specified number of operations on a given number of
// goroutines.
type OptimizedProcess struct {
	// The number of iterations between optimizations.
	optimizationInterval int

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

	// The last error collected during optimization.
	previousError float64

	baselineError float64

	totalError float64

	removeCount int

	removeCountMutex sync.Mutex

	// The time of the last report collected during optimization.
	lastReportTime time.Time

	// A mutex to protect against sumultaneous read/write to lastReport.
	reportMutex sync.Mutex

	cpuCount int

	reporter *Reporter
}

// MARK: Initializers

// NewOptimizedProcess creates and returns a new parallel process with the specified
// number of goroutines.
func NewOptimizedProcess(interval int, gain float64) *OptimizedProcess {
	return &OptimizedProcess{
		optimizationInterval: interval,
		cpuCount:             runtime.NumCPU(),
		reporter:             NewReporter(),
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

	var errors []float64
	var controller []float64
	var routines []int
	for _, v := range reports {
		errors = append(errors, v.e)
		controller = append(controller, v.u)
		routines = append(routines, v.n)
	}

	// log.Println(diffs)

	PlotSignal(errors, "Error", "Error/Optimization", "Optimization Number", "Time Since Last Operation", "error.png", nil)
	PlotSignal(controller, "Controller", "Controller/Optimization", "Optimization Number", "Time Since Last Operation", "controller.png", nil)
	PlotSignalI(routines, "Routines", "Routines/Optimization", "Optimization Number", "Number of Routines", "routines.png", nil)
}

// MARK: Private methods

// reset resets all of the process' properties to their initial state.
func (p *OptimizedProcess) reset() {
	reports = nil
	p.numRoutines = 1
	p.count = 0
	p.removeCount = 0
	p.previousError = 0.0
	p.totalError = 0.0
	p.reporter.Reset()
	p.lastReportTime = time.Now()
}

// runRoutine runs a new routine for the given number of iterations, picking up
// where other routines have left off.
func (p *OptimizedProcess) runRoutine(iterations int, operation Operation) {
	i := p.nextCount()
	for i <= iterations {
		operation(i-1)

		p.optimizeNumRoutines(i, iterations, operation)

		p.removeCountMutex.Lock()
		p.numRoutinesMutex.Lock()
		if p.removeCount > 0 && p.numRoutines > 1 {
			p.removeCount--
			p.numRoutines--
			p.numRoutinesMutex.Unlock()
			p.removeCountMutex.Unlock()
			break
		} else if p.removeCount > 0 {
			p.removeCount--
		}
		p.numRoutinesMutex.Unlock()
		p.removeCountMutex.Unlock()

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
func (p *OptimizedProcess) optimizeNumRoutines(iteration int, iterations int, operation Operation) {
	if iteration%p.optimizationInterval != 0 || iteration == 0 {
		return
	}

	p.group.Add(1)
	nextAction, c, u, e := p.nextAction(iteration)

	finished := func() {
		reports = append(reports, report{e, u, p.numRoutines})
	}
	defer finished()

	switch nextAction {
	case routineActionNone:
		p.group.Done()
	case routineActionAdd:
		if c == 0 {
			p.group.Done()
		} else if c > 0 {
			if c > 1 {
				p.group.Add(c - 1)
			}

			for i := 0; i < c; i++ {
				p.numRoutinesMutex.Lock()
				p.numRoutines++
				p.numRoutinesMutex.Unlock()
				go p.runRoutine(iterations, operation)
			}
		}
	case routineActionRemove:
		p.removeCountMutex.Lock()
		p.removeCount = c
		p.removeCountMutex.Unlock()
		p.group.Done()
	}
}

type report struct {
	e float64
	u float64
	n int
}

var reports []report
var diffs []float64

// nextAction gets the next action the optimize method should use when
// optimizing the number of goroutines.
func (p *OptimizedProcess) nextAction(iteration int) (routineAction, int, float64, float64) {
	p.reportMutex.Lock()
	defer p.reportMutex.Unlock()

	usage, t := p.reporter.Usage()
	// usage = math.Round(usage*100.0) / 100.0

	P := 1.0 - (usage / float64(p.cpuCount))
	P = (0.01 * P) + (0.99 * p.previousError)
	I := p.totalError + P
	D := (P - p.previousError) / t
	U := 20.0*P + 0.01*I - 3.0*D
	N := int(math.Round(U)) - p.numRoutines

	p.previousError = P
	p.totalError += P

	usage /= float64(p.cpuCount)
	if N > 0 {
		return routineActionAdd, N, usage, P
	} else if N < 0 {
		return routineActionRemove, -N, usage, P
	}

	return routineActionNone, 0, usage, P
}

// PlotSignal plots a signal and saves the image to a file.
func PlotSignal(signal []float64, series string, title string, xAxis string, yAxis string, file string, horizontalLines []float64) {
	pe, err := plot.New()
	if err != nil {
		panic(err)
	}

	pe.Title.Text = title
	pe.X.Label.Text = xAxis
	pe.Y.Label.Text = yAxis

	errorValues := make(plotter.XYs, len(signal))

	var horizontalLinePoints []plotter.XYs
	for range horizontalLines {
		horizontalLinePoints = append(horizontalLinePoints, make(plotter.XYs, len(signal)))
	}

	for i, v := range signal {
		errorValues[i].X = float64(i)
		errorValues[i].Y = v

		for j := range horizontalLinePoints {
			horizontalLinePoints[j][i].X = float64(i)
			horizontalLinePoints[j][i].Y = horizontalLines[j]
		}
	}

	err = plotutil.AddLines(pe,
		series, errorValues,
	)

	for i, v := range horizontalLinePoints {
		_ = plotutil.AddLines(pe, fmt.Sprintf("price volume line %d", i), v)
	}

	if err != nil {
		panic(err)
	}

	// Save the plot to a PNG file.
	if err := pe.Save(16*vg.Inch, 8*vg.Inch, file); err != nil {
		panic(err)
	}
}

// PlotSignalI plots a signal and saves the image to a file.
func PlotSignalI(signal []int, series string, title string, xAxis string, yAxis string, file string, horizontalLines []float64) {
	pe, err := plot.New()
	if err != nil {
		panic(err)
	}

	pe.Title.Text = title
	pe.X.Label.Text = xAxis
	pe.Y.Label.Text = yAxis

	errorValues := make(plotter.XYs, len(signal))

	var horizontalLinePoints []plotter.XYs
	for range horizontalLines {
		horizontalLinePoints = append(horizontalLinePoints, make(plotter.XYs, len(signal)))
	}

	for i, v := range signal {
		errorValues[i].X = float64(i)
		errorValues[i].Y = float64(v)

		for j := range horizontalLinePoints {
			horizontalLinePoints[j][i].X = float64(i)
			horizontalLinePoints[j][i].Y = horizontalLines[j]
		}
	}

	err = plotutil.AddLines(pe,
		series, errorValues,
	)

	for i, v := range horizontalLinePoints {
		_ = plotutil.AddLines(pe, fmt.Sprintf("price volume line %d", i), v)
	}

	if err != nil {
		panic(err)
	}

	// Save the plot to a PNG file.
	if err := pe.Save(16*vg.Inch, 8*vg.Inch, file); err != nil {
		panic(err)
	}
}
