package pipe

import (
	"errors"
	"github.com/gptlocal/gosugar/bufio"
	"github.com/gptlocal/gosugar/concurrent"
	"github.com/gptlocal/gosugar/concurrent/done"
	"github.com/gptlocal/gosugar/lang"
	"io"
	"runtime"
	"sync"
	"time"
)

type pipe struct {
	sync.Mutex
	data        bufio.MultiBuffer
	readSignal  *concurrent.Notifier
	writeSignal *concurrent.Notifier
	done        *done.Instance
	errChan     chan error
	option      pipeOption
	state       state
}

var (
	errBufferFull = errors.New("buffer full")
	errSlowDown   = errors.New("slow down")
)

func (p *pipe) getState(forRead bool) error {
	switch p.state {
	case opened:
		if !forRead && p.option.isFull(p.data.Len()) {
			return errBufferFull
		}
		return nil
	case closed:
		if !forRead {
			return io.ErrClosedPipe
		}
		if !p.data.IsEmpty() {
			return nil
		}
		return io.EOF
	case errord:
		return io.ErrClosedPipe
	default:
		panic("impossible case")
	}
}

func (p *pipe) readMultiBufferInternal() (bufio.MultiBuffer, error) {
	p.Lock()
	defer p.Unlock()

	if err := p.getState(true); err != nil {
		return nil, err
	}

	data := p.data
	p.data = nil
	return data, nil
}

func (p *pipe) ReadMultiBuffer() (bufio.MultiBuffer, error) {
	for {
		data, err := p.readMultiBufferInternal()
		if data != nil || err != nil {
			p.writeSignal.Signal()
			return data, err
		}

		select {
		case <-p.readSignal.Wait():
		case <-p.done.Wait():
		case err = <-p.errChan:
			return nil, err
		}
	}
}

func (p *pipe) ReadMultiBufferTimeout(d time.Duration) (bufio.MultiBuffer, error) {
	timer := time.NewTimer(d)
	defer timer.Stop()

	for {
		data, err := p.readMultiBufferInternal()
		if data != nil || err != nil {
			p.writeSignal.Signal()
			return data, err
		}

		select {
		case <-p.readSignal.Wait():
		case <-p.done.Wait():
		case <-timer.C:
			return nil, bufio.ErrReadTimeout
		}
	}
}

func (p *pipe) writeMultiBufferInternal(mb bufio.MultiBuffer) error {
	p.Lock()
	defer p.Unlock()

	if err := p.getState(false); err != nil {
		return err
	}

	if p.data == nil {
		p.data = mb
		return nil
	}

	p.data, _ = bufio.MergeMulti(p.data, mb)
	return errSlowDown
}

func (p *pipe) WriteMultiBuffer(mb bufio.MultiBuffer) error {
	if mb.IsEmpty() {
		return nil
	}

	if p.option.onTransmission != nil {
		mb = p.option.onTransmission(mb)
	}

	for {
		err := p.writeMultiBufferInternal(mb)
		if err == nil {
			p.readSignal.Signal()
			return nil
		}

		if err == errSlowDown {
			p.readSignal.Signal()

			// Yield current goroutine. Hopefully the reading counterpart can pick up the payload.
			runtime.Gosched()
			return nil
		}

		if err == errBufferFull && p.option.discardOverflow {
			bufio.ReleaseMulti(mb)
			return nil
		}

		if err != errBufferFull {
			bufio.ReleaseMulti(mb)
			p.readSignal.Signal()
			return err
		}

		select {
		case <-p.writeSignal.Wait():
		case <-p.done.Wait():
			return io.ErrClosedPipe
		}
	}
}

func (p *pipe) Close() error {
	p.Lock()
	defer p.Unlock()

	if p.state == closed || p.state == errord {
		return nil
	}

	p.state = closed
	lang.Must(p.done.Close())
	return nil
}

// Interrupt implements common.Interruptible.
func (p *pipe) Interrupt() {
	p.Lock()
	defer p.Unlock()

	if p.state == closed || p.state == errord {
		return
	}

	p.state = errord

	if !p.data.IsEmpty() {
		bufio.ReleaseMulti(p.data)
		p.data = nil
	}

	lang.Must(p.done.Close())
}
