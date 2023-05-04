package concurrent_test

import (
	"testing"

	. "github.com/gptlocal/gosugar/concurrent"
)

func TestSemaphoreSignal(t *testing.T) {
	n := 10
	s := NewSemaphore(n)

	for i := 0; i < n; i++ {
		s.Wait()
		go func() {
			s.Signal()
		}()
	}

	for i := 0; i < n; i++ {
		w := s.Wait()
		select {
		case <-w:
		default:
			t.Log("Could not acquire semaphore")
			t.Fail()
		}
	}
}
