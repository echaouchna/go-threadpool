package concurrent

import (
	"testing"
	"time"
)

func sleep() {
	time.Sleep(time.Second)
}

func TestPausePlayStopBuffered(t *testing.T) {
	channel := make(chan Action, 1)
	defer close(channel)
	jobs := make(map[string]JobFunc)
	jobs["dumb"] = func(id int, value interface{}) {}
	play, pause, stop, _ := RunWorkers(channel, jobs, 0)

	for i := 0; i < 10; i++ {
		channel <- Action{"dumb", 0}
	}

	sleep()

	length := len(channel)

	if length != 0 {
		t.Errorf("start: len(channel) = %v, want %v", length, 0)
	}

	pause()

	sleep()

	channel <- Action{"dumb", 0}

	length = len(channel)

	if length != 1 {
		t.Errorf("pause: len(channel) = %v, want %v", length, 1)
	}

	play()

	sleep()

	length = len(channel)

	if length != 0 {
		t.Errorf("play: len(channel) = %v, want %v", length, 0)
	}

	stop()

	sleep()

	channel <- Action{"dumb", 0}

	length = len(channel)

	if length != 1 {
		t.Errorf("stop: len(channel) = %v, want %v", length, 1)
	}
}

func TestPausePlayStopUnbuffered(t *testing.T) {
	channel := make(chan Action)
	defer close(channel)
	jobs := make(map[string]JobFunc)
	jobs["dumb"] = func(id int, value interface{}) {}
	play, pause, stopJobs, getStatus := RunWorkers(channel, jobs, 0)

	for i := 0; i < 10; i++ {
		channel <- Action{"dumb", 0}
	}

	sleep()

	length := len(channel)

	if length != 0 {
		t.Errorf("start: len(channel) = %v, want %v", length, 0)
	}

	pause()

	sleep()

	blocked := true

	go func() {
		channel <- Action{"dumb", 0}
		blocked = false
	}()

	sleep()

	if getStatus().String() != "paused" {
		t.Errorf("pause: getStatus().String() = %v, want %v", getStatus().String(), "paused")
	}

	if !blocked {
		t.Errorf("pause: blocked = %v, want %v", blocked, true)
	}

	play()

	channel <- Action{"dumb", 0}

	sleep()

	if getStatus().String() != "running" {
		t.Errorf("play: getStatus().String() = %v, want %v", getStatus().String(), "running")
	}

	if blocked {
		t.Errorf("play: blocked = %v, want %v", blocked, false)
	}

	length = len(channel)

	if length != 0 {
		t.Errorf("play: len(channel) = %v, want %v", length, 0)
	}

	stopJobs()

	sleep()

	if getStatus().String() != "stopped" {
		t.Errorf("stop: getStatus().String() = %v, want %v", getStatus().String(), "stopped")
	}
}
