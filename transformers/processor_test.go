package transformers_test

import (
	"github.com/drborges/rivers"
	"github.com/drborges/rivers/stream"
	"github.com/drborges/rivers/transformers"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestProcessor(t *testing.T) {
	evensFilter := func(d stream.T, out stream.Writable) {
		if d.(int)%2 == 0 {
			out <- d
		}
	}

	Convey("Given I have a context", t, func() {
		context := rivers.NewContext()

		Convey("And a stream of data", func() {
			in, out := stream.New(2)
			out <- 1
			out <- 2
			close(out)

			Convey("When I apply the transformer to the stream", func() {
				transformed := transformers.New(context).ProcessWith(evensFilter).Transform(in)

				Convey("Then a transformed stream is returned", func() {
					So(transformed.Read(), ShouldResemble, []stream.T{2})
				})
			})

			Convey("When I close the context", func() {
				context.Close()

				Convey("And I apply the transformer to the stream", func() {
					next := transformers.New(context).ProcessWith(evensFilter).Transform(in)

					Convey("Then no item is sent to the next stage", func() {
						So(next.Read(), ShouldBeEmpty)
					})
				})
			})
		})
	})
}
