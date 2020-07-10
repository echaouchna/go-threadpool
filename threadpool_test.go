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
	handler := RunWorkers(channel, jobs, 0)

	for i := 0; i < 10; i++ {
		channel <- Action{"dumb", 0}
	}

	sleep()

	length := len(channel)

	if length != 0 {
		t.Errorf("start: len(channel) = %v, want %v", length, 0)
	}

	handler.Pause()

	sleep()

	channel <- Action{"dumb", 0}

	length = len(channel)

	if length != 1 {
		t.Errorf("pause: len(channel) = %v, want %v", length, 1)
	}

	handler.Play()

	sleep()

	length = len(channel)

	if length != 0 {
		t.Errorf("play: len(channel) = %v, want %v", length, 0)
	}

	handler.Quit()

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
	handler := RunWorkers(channel, jobs, 0)

	for i := 0; i < 10; i++ {
		channel <- Action{"dumb", 0}
	}

	sleep()

	length := len(channel)

	if length != 0 {
		t.Errorf("start: len(channel) = %v, want %v", length, 0)
	}

	handler.Pause()

	sleep()

	blocked := true

	go func() {
		channel <- Action{"dumb", 0}
		blocked = false
	}()

	sleep()

	if handler.GetStatus().String() != "paused" {
		t.Errorf("pause: getStatus().String() = %v, want %v", handler.GetStatus().String(), "paused")
	}

	if !blocked {
		t.Errorf("pause: blocked = %v, want %v", blocked, true)
	}

	handler.Play()

	channel <- Action{"dumb", 0}

	sleep()

	if handler.GetStatus().String() != "running" {
		t.Errorf("play: getStatus().String() = %v, want %v", handler.GetStatus().String(), "running")
	}

	if blocked {
		t.Errorf("play: blocked = %v, want %v", blocked, false)
	}

	length = len(channel)

	if length != 0 {
		t.Errorf("play: len(channel) = %v, want %v", length, 0)
	}

	handler.Quit()

	sleep()

	if handler.GetStatus().String() != "stopped" {
		t.Errorf("stop: getStatus().String() = %v, want %v", handler.GetStatus().String(), "stopped")
	}
}
