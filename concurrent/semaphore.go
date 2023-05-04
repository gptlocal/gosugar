package concurrent

// Semaphore is an implementation of semaphore.
type Semaphore struct {
	token chan struct{}
}

// NewSemaphore create a new Semaphore with n permits.
func NewSemaphore(n int) *Semaphore {
	s := &Semaphore{
		token: make(chan struct{}, n),
	}
	for i := 0; i < n; i++ {
		s.token <- struct{}{}
	}
	return s
}

// Signal releases a permit into the semaphore.
func (s *Semaphore) Signal() {
	s.token <- struct{}{}
}

// Wait returns a channel for acquiring a permit.
func (s *Semaphore) Wait() <-chan struct{} {
	return s.token
}
