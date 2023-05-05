package bufio

import "time"

//go:generate go run github.com/gptlocal/gosugar/codegen/errors -errmod=github.com/gptlocal/gosugar/errors

// ErrReadTimeout is an error that happens with IO timeout.
var ErrReadTimeout = newError("IO timeout")

// Reader extends io.Reader with MultiBuffer.
type Reader interface {
	// ReadMultiBuffer reads content from underlying reader, and put it into a MultiBuffer.
	ReadMultiBuffer() (MultiBuffer, error)
}

// TimeoutReader is a reader that returns error if Read() operation takes longer than the given timeout.
type TimeoutReader interface {
	ReadMultiBufferTimeout(time.Duration) (MultiBuffer, error)
}

// Writer extends io.Writer with MultiBuffer.
type Writer interface {
	// WriteMultiBuffer writes a MultiBuffer into underlying writer.
	WriteMultiBuffer(MultiBuffer) error
}
