package context

import (
	goContext "context"
	"time"
)

var (
	// DefaultConfig the default configuration used to create a context.
	DefaultConfig = Config{
		Timeout:    5 * time.Second,
		BufferSize: 1000,
	}
)

// Config configuration used to adjust the behavior of the context.
type Config struct {
	// Timeout the timeout which after the context is automatically closed.
	Timeout time.Duration
	// BufferSize the size of the buffer used by workers bound to this context.
	BufferSize int
}

// Context implements the golang context.Context interface, with support to
// Context Trees, providing a different semanthics for cancellation propagation.
// In a Context Tree, cancellation is propagated from leaf nodes up to their
// parent. A parent node can only be canceled if all its children are canceled.
// For example, given the following Context Tree:
//
//  ┏━━━> c2
// c1 ━━> c3 ━━> c5
//  ┗━━━> c4
//
// The context c1, con only be canceld if all its children (c2, c3, and c4) are
// already canceld. Similarly, c3 can only be canceled if c5 is canceled,
// threfore, c1 depends on c5 being canceled befre it may be canceled.
//
// This abstraction enables the creation of stream pipelines, wheren downstreams
// can signal to their upstream when they are done consuming data, freeing the
// upstream to cease its work when no more data is required by its downstreams.
type Context interface {
	// Implements the golang context.Context interface.
	goContext.Context
	// Config returns the configuration used to create the context.
	Config() Config
	// Close attempts to close the context. If the context still has opened
	// children, this operation will be a noop.
	Close()
	// NewChild creates a new child Context.
	NewChild() Context
}

// New creates a new Context.
func New() Context {
	return FromStdContext(goContext.Background())
}

// FromStdContext creates a new Context from the standard golang context.
func FromStdContext(stdCtx goContext.Context) Context {
	ctx, cancel := goContext.WithCancel(stdCtx)
	return &context{ctx, cancel, make([]Context, 0), DefaultConfig}
}

type context struct {
	goContext.Context
	cancel   goContext.CancelFunc
	children []Context
	config   Config
}

func (ctx *context) Config() Config {
	return ctx.config
}

func (ctx *context) Close() {
	for _, child := range ctx.children {
		select {
		case <-child.Done():
		default:
			// Parent can only close the context when all children have closed theirs.
			return
		}
	}
	ctx.cancel()
}

func (parent *context) NewChild() Context {
	ctx, cancel := goContext.WithCancel(parent.Context)
	cancelWithPropagation := func() {
		cancel()
		parent.Close()
	}
	child := &context{ctx, cancelWithPropagation, make([]Context, 0), parent.config}
	parent.children = append(parent.children, child)
	return child
}
