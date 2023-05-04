package task_test

import (
	"testing"
	"time"

	. "github.com/gptlocal/gosugar/concurrent/task"
	"github.com/gptlocal/gosugar/lang"
)

func TestPeriodicTaskStop(t *testing.T) {
	value := 0
	task := &Periodic{
		Interval: time.Second * 2,
		Execute: func() error {
			value++
			return nil
		},
	}
	lang.Must(task.Start())
	time.Sleep(time.Second * 5)
	lang.Must(task.Close())
	if value != 3 {
		t.Fatal("expected 3, but got ", value)
	}
	time.Sleep(time.Second * 4)
	if value != 3 {
		t.Fatal("expected 3, but got ", value)
	}
	lang.Must(task.Start())
	time.Sleep(time.Second * 3)
	if value != 5 {
		t.Fatal("Expected 5, but ", value)
	}
	lang.Must(task.Close())
}
