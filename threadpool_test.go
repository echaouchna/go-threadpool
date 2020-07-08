package concurrent

import (
	"testing"
	"time"
)

func TestPausePlayStop(t *testing.T) {
	channel := make(chan Action, 1)
	defer close(channel)
	jobs := make(map[string]JobFunc)
	jobs["dumb"] = func(id int, value interface{}) {}
	play, pause, stopJobs := RunWorkers(channel, jobs, 0)

	for i := 0; i < 10; i++ {
		channel <- Action{"dumb", 0}
		// fmt.Println(i)
	}

	time.Sleep(time.Second)

	length := len(channel)

	if length != 0 {
		t.Errorf("len(channel) = %v, want %v", length, 0)
	}

	pause()

	time.Sleep(time.Second)

	channel <- Action{"dumb", 0}

	length = len(channel)

	if length != 1 {
		t.Errorf("pause: len(channel) = %v, want %v", length, 1)
	}

	play()

	time.Sleep(time.Second)

	length = len(channel)

	if length != 0 {
		t.Errorf("play: len(channel) = %v, want %v", length, 0)
	}

	stopJobs()

	time.Sleep(time.Second)

	channel <- Action{"dumb", 0}

	length = len(channel)

	if length != 1 {
		t.Errorf("pause: len(channel) = %v, want %v", length, 1)
	}
}
