package pubsub_test

import (
	"testing"

	. "github.com/gptlocal/gosugar/concurrent/pubsub"
)

func TestPubSub(t *testing.T) {
	topicService := NewTopicService()

	sub := topicService.Subscribe("a")
	topicService.Publish("a", 1)

	select {
	case v := <-sub.Wait():
		if v != 1 {
			t.Error("expected subscribed value 1, but got ", v)
		}
	default:
		t.Fail()
	}

	sub.Close()
	topicService.Publish("a", 2)

	select {
	case <-sub.Wait():
		t.Fail()
	default:
	}

	topicService.Cleanup()
}
