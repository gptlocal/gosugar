package concurrent_test

import (
	"testing"

	. "github.com/gptlocal/gosugar/concurrent"
)

func TestNotifierSignal(t *testing.T) {
	n := NewNotifier()

	w := n.Wait()
	n.Signal()

	select {
	case <-w:
	default:
		t.Fail()
	}
}
