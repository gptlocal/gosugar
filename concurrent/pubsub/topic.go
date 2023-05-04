package pubsub

import (
	"sync"
	"time"

	"github.com/gptlocal/gosugar/concurrent/done"
	"github.com/gptlocal/gosugar/concurrent/task"
	"github.com/gptlocal/gosugar/errors"
	"github.com/gptlocal/gosugar/lang"
)

type TopicService struct {
	sync.RWMutex
	subscribers map[string][]*Subscriber
	periodic    *task.Periodic
}

func NewTopicService() *TopicService {
	s := &TopicService{
		subscribers: make(map[string][]*Subscriber),
	}
	s.periodic = &task.Periodic{
		Execute:  s.Cleanup,
		Interval: time.Second * 30,
	}
	return s
}

// Cleanup cleans up internal caches of subscribers.
// Visible for testing only.
func (t *TopicService) Cleanup() error {
	t.Lock()
	defer t.Unlock()

	if len(t.subscribers) == 0 {
		return errors.New("nothing to do")
	}

	for name, subs := range t.subscribers {
		newSub := make([]*Subscriber, 0, len(t.subscribers))
		for _, sub := range subs {
			if !sub.IsClosed() {
				newSub = append(newSub, sub)
			}
		}
		if len(newSub) == 0 {
			delete(t.subscribers, name)
		} else {
			t.subscribers[name] = newSub
		}
	}

	if len(t.subscribers) == 0 {
		t.subscribers = make(map[string][]*Subscriber)
	}
	return nil
}

func (t *TopicService) Subscribe(topic string) *Subscriber {
	subscriber := &Subscriber{
		buffer: make(chan interface{}, 16),
		done:   done.New(),
	}
	t.Lock()
	t.subscribers[topic] = append(t.subscribers[topic], subscriber)
	t.Unlock()
	lang.Must(t.periodic.Start())
	return subscriber
}

func (t *TopicService) Publish(topic string, message interface{}) {
	t.RLock()
	defer t.RUnlock()

	for _, subscriber := range t.subscribers[topic] {
		if !subscriber.IsClosed() {
			subscriber.push(message)
		}
	}
}
