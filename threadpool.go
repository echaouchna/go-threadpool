package concurrent

import (
	"runtime"
	"sync"

	log "github.com/sirupsen/logrus"
)

// JobFunc the function type to be run by the jobs
type JobFunc func(int, interface{})

// Action data used by the jobs
// * Name the type of action, should match a name of a function in the jobFunctions map
// will used to select the function to run
// Data the data to pass to the jobFunc
type Action struct {
	Name string
	Data interface{}
}

// Status type
type Status uint32

// Convert the Status to a string. E.g. Running becomes "running".
func (status Status) String() (strStatus string) {
	strStatus = "undefined"
	switch status {
	case Running:
		strStatus = "running"
	case Paused:
		strStatus = "paused"
	case Stopped:
		strStatus = "stopped"
	}
	return
}

// AllStatus is a constant exposing all possible Status
var AllStatus = []Status{
	Running,
	Paused,
	Stopped,
}

const (
	// Running means all jobs are running
	Running Status = iota
	// Paused means all jobs are paused
	Paused
	// Stopped means all jobs are stopped
	Stopped
)

func threadMain(id int, queue chan Action, wg *sync.WaitGroup, jobs map[string]JobFunc) (playCommand chan bool, pauseCommand chan bool, quitCommand chan bool) {
	quitCommand = make(chan bool, 1)
	pauseCommand = make(chan bool, 1)
	playCommand = make(chan bool, 1)
	go func() {
		wg.Add(1)
		defer wg.Done()
		defer close(quitCommand)
		defer close(pauseCommand)
		defer close(playCommand)
		channel := queue
		for {
			select {
			case action := <-channel:
				log.Debugf("Thread %d running", id)
				if job, ok := jobs[action.Name]; ok {
					job(id, action.Data)
				}
			case <-quitCommand:
				log.Debugf("Thread %d quitting", id)
				return
			case <-pauseCommand:
				channel = nil
				log.Debugf("Thread %d pausing", id)
			case <-playCommand:
				log.Debugf("Thread %d playing", id)
				channel = queue
			}
		}
	}()
	return
}

// RunWorkers create and run jobs
// * queue channel of Action type, jobs will listen to this queue
// * jobFunctions containing the jobFunc to be used by the jobs and the action nmaes used to select
// one of these functions
// * workersNumber the number of jobs
//   - if workersNumber <= 0 or workersNumber > 2*cpuCount ==> runtime.NumCPU() will be used
//   - otherwise workersNumber will be used
func RunWorkers(queue chan Action, jobFunctions map[string]JobFunc, workersNumber int) (play func(), pause func(), quit func(), getStatus func() Status) {
	var wg sync.WaitGroup
	cpuCount := runtime.NumCPU()
	jobCount := cpuCount
	if workersNumber > 0 && workersNumber <= 2*cpuCount {
		jobCount = workersNumber
	}
	runtime.GOMAXPROCS(cpuCount)

	log.Infof("Running %d workers", jobCount)

	playCommands := make([]chan bool, jobCount)
	pauseCommands := make([]chan bool, jobCount)
	quitCommands := make([]chan bool, jobCount)
	currentStatus := Stopped
	for i := 0; i < jobCount; i++ {
		playCommands[i], pauseCommands[i], quitCommands[i] = threadMain(i+1, queue, &wg, jobFunctions)
	}
	pause = func() {
		for _, pauseCommand := range pauseCommands {
			pauseCommand <- true
		}
		currentStatus = Paused
	}
	play = func() {
		for _, playCommand := range playCommands {
			playCommand <- true
		}
		currentStatus = Running
	}
	quit = func() {
		for _, quitCommand := range quitCommands {
			quitCommand <- true
		}
		wg.Wait()
		currentStatus = Stopped
	}
	getStatus = func() Status {
		return currentStatus
	}
	return
}
