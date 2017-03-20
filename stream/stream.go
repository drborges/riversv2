package stream

import "github.com/drborges/rivers/ctxtree"

// T data type flowing through rivers streams
type T interface{}

// Readable a channel which one may read data from
type Readable <-chan T

// Writable a channel which one may write data to
type Writable chan<- T

// Reader provides means to read from a readable stream as well as signal the
// termination of the stream, gracefully or not.
type Reader interface {
	// Close closes the stream indicating that no further data will
	// be read from this reader. If no error is provided, the stream
	// is then gracefully closed. Closing with an error, indicates a
	// failure.
	Close(error)

	// Read provides a readable stream from which data can be read
	Read() Readable

	// Done returns a read-only channel which when closed, signals to all tasks
	// bound to the underlying context should halt their jobs.
	Done() <-chan struct{}

	// NewDownstream creates the components reader and writer of a
	// new downstream, which is bound to this stream. This
	// relationship dictates how stream cancellation is propagated:
	// 1. A stream is only closed if all its downstreams are also
	// closed.
	// 2. Closing a downstream, propagates the cancellation signal
	// to the upstream, which then checks whether or not it should
	// close itself.
	NewDownstream() (Reader, Writer)
}

// Writer provides means to write data to a writable stream as well as signal the
// termination of the stream, gracefully or not. Closing also closes the underlying
// writable stream.
type Writer interface {
	// Close closes the stream indicating that no further data will
	// be written to the stream. If no error is provided, the stream
	// is then gracefully closed. Closing with an error, indicates a
	// failure.
	Close(error)

	// Write writes the given items to the underlying writable stream, returning an
	// error in case of a failure.
	Write(items ...T) error

	// Writable exposes the underlying writable channel.
	Writable() Writable
}

// Empty represents an empty readable stream which has already ceased producing
// data.
var Empty = func() Reader {
	r, w := New(ctxtree.New())
	w.Close(nil)
	return r
}()

// New Creates the Reader and Writer components of a rivers stream with the default
// configuration.
func New(ctx ctxtree.Context) (Reader, Writer) {
	ch := make(chan T, ctx.Config().BufferSize)
	return &reader{ctx, ch}, &writer{ctx, ch}
}

type reader struct {
	ctx ctxtree.Context
	ch  chan T
}

func (reader *reader) Read() Readable {
	return reader.ch
}

func (reader *reader) Done() <-chan struct{} {
	return reader.ctx.Done()
}

func (reader *reader) Close(err error) {
	reader.ctx.Close(err)
}

func (reader *reader) NewDownstream() (Reader, Writer) {
	return New(reader.ctx.NewChild())
}

type writer struct {
	ctx ctxtree.Context
	ch  chan T
}

func (writer *writer) write(data T) error {
	select {
	case <-writer.ctx.Done():
		return writer.ctx.Err()
	default:
		select {
		case writer.ch <- data:
		case <-writer.ctx.Done(): // Eventually times out
			return writer.ctx.Err()
		}
	}
	return nil
}

func (writer *writer) Write(items ...T) error {
	for _, item := range items {
		if err := writer.write(item); err != nil {
			return err
		}
	}

	return nil
}

func (writer *writer) Close(err error) {
	defer close(writer.ch)
	writer.ctx.Close(err)
}

func (writer *writer) Writable() Writable {
	return writer.ch
}
