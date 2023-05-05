package pipe_test

import (
	"errors"
	"io"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/gptlocal/gosugar/bufio"
	"github.com/gptlocal/gosugar/lang"
	. "github.com/gptlocal/gosugar/pipe"
	"golang.org/x/sync/errgroup"
)

func TestPipeReadWrite(t *testing.T) {
	pReader, pWriter := New(WithSizeLimit(1024))

	b := bufio.New()
	b.WriteString("abcd")
	lang.Must(pWriter.WriteMultiBuffer(bufio.MultiBuffer{b}))

	b2 := bufio.New()
	b2.WriteString("efg")
	lang.Must(pWriter.WriteMultiBuffer(bufio.MultiBuffer{b2}))

	rb, err := pReader.ReadMultiBuffer()
	lang.Must(err)
	if r := cmp.Diff(rb.String(), "abcdefg"); r != "" {
		t.Error(r)
	}
}

func TestPipeInterrupt(t *testing.T) {
	pReader, pWriter := New(WithSizeLimit(1024))
	payload := []byte{'a', 'b', 'c', 'd'}
	b := bufio.New()
	b.Write(payload)
	lang.Must(pWriter.WriteMultiBuffer(bufio.MultiBuffer{b}))
	pWriter.Interrupt()

	rb, err := pReader.ReadMultiBuffer()
	if err != io.ErrClosedPipe {
		t.Fatal("expect io.ErrClosePipe, but got ", err)
	}
	if !rb.IsEmpty() {
		t.Fatal("expect empty buffer, but got ", rb.Len())
	}
}

func TestPipeClose(t *testing.T) {
	pReader, pWriter := New(WithSizeLimit(1024))
	payload := []byte{'a', 'b', 'c', 'd'}
	b := bufio.New()
	lang.Must2(b.Write(payload))
	lang.Must(pWriter.WriteMultiBuffer(bufio.MultiBuffer{b}))
	lang.Must(pWriter.Close())

	rb, err := pReader.ReadMultiBuffer()
	lang.Must(err)
	if rb.String() != string(payload) {
		t.Fatal("expect content ", string(payload), " but actually ", rb.String())
	}

	rb, err = pReader.ReadMultiBuffer()
	if err != io.EOF {
		t.Fatal("expected EOF, but got ", err)
	}
	if !rb.IsEmpty() {
		t.Fatal("expect empty buffer, but got ", rb.String())
	}
}

func TestPipeLimitZero(t *testing.T) {
	pReader, pWriter := New(WithSizeLimit(0))
	bb := bufio.New()
	lang.Must2(bb.Write([]byte{'a', 'b'}))
	lang.Must(pWriter.WriteMultiBuffer(bufio.MultiBuffer{bb}))

	var errg errgroup.Group
	errg.Go(func() error {
		b := bufio.New()
		b.Write([]byte{'c', 'd'})
		return pWriter.WriteMultiBuffer(bufio.MultiBuffer{b})
	})
	errg.Go(func() error {
		time.Sleep(time.Second)

		var container bufio.MultiBufferContainer
		if err := bufio.Copy(pReader, &container); err != nil {
			return err
		}

		if r := cmp.Diff(container.String(), "abcd"); r != "" {
			return errors.New(r)
		}
		return nil
	})
	errg.Go(func() error {
		time.Sleep(time.Second * 2)
		return pWriter.Close()
	})
	if err := errg.Wait(); err != nil {
		t.Error(err)
	}
}

func TestPipeWriteMultiThread(t *testing.T) {
	pReader, pWriter := New(WithSizeLimit(0))

	var errg errgroup.Group
	for i := 0; i < 10; i++ {
		errg.Go(func() error {
			b := bufio.New()
			b.WriteString("abcd")
			return pWriter.WriteMultiBuffer(bufio.MultiBuffer{b})
		})
	}
	time.Sleep(time.Millisecond * 100)
	pWriter.Close()
	errg.Wait()

	b, err := pReader.ReadMultiBuffer()
	lang.Must(err)
	if r := cmp.Diff(b[0].Bytes(), []byte{'a', 'b', 'c', 'd'}); r != "" {
		t.Error(r)
	}
}

func TestInterfaces(t *testing.T) {
	_ = (bufio.Reader)(new(Reader))
	_ = (bufio.TimeoutReader)(new(Reader))

	_ = (lang.Interruptible)(new(Reader))
	_ = (lang.Interruptible)(new(Writer))
	_ = (lang.Closable)(new(Writer))
}

func BenchmarkPipeReadWrite(b *testing.B) {
	reader, writer := New(WithoutSizeLimit())
	a := bufio.New()
	a.Extend(bufio.Size)
	c := bufio.MultiBuffer{a}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lang.Must(writer.WriteMultiBuffer(c))
		d, err := reader.ReadMultiBuffer()
		lang.Must(err)
		c = d
	}
}
