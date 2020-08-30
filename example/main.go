package main

import (
	"time"

	"github.com/256dpi/maxgo"
)

type instance struct {
	in1   *maxgo.Inlet
	in2   *maxgo.Inlet
	out1  *maxgo.Outlet
	out2  *maxgo.Outlet
	tick  *time.Ticker
	bench bool
}

func (i *instance) Init(obj *maxgo.Object, args []maxgo.Atom) bool {
	// check if benchmark
	if len(args) > 0 && args[0] == "bench" {
		i.bench = true
	}

	// check bench
	if !i.bench {
		maxgo.Pretty("init", args)
	}

	// declare inlets
	i.in1 = obj.Inlet(maxgo.Any, "example inlet 1", true)
	i.in2 = obj.Inlet(maxgo.Float, "example inlet 2", false)

	// declare outlets
	i.out1 = obj.Outlet(maxgo.Any, "example outlet 1")
	i.out2 = obj.Outlet(maxgo.Bang, "example outlet 2")

	// bang second outlet from a timer
	if !i.bench {
		// create timer
		i.tick = time.NewTicker(time.Second)

		// send a bang for every tick
		go func() {
			var j int
			for range i.tick.C {
				maxgo.Pretty("tick", maxgo.IsMainThread())

				// bang immediately or defer
				if j++; j%2 == 0 {
					i.out2.Bang()
				} else {
					maxgo.Defer(func() {
						maxgo.Pretty("defer", maxgo.IsMainThread())
						i.out2.Bang()
					})
				}
			}
		}()
	}

	return true
}

func (i *instance) Handle(msg string, inlet int, data []maxgo.Atom) {
	// check bench
	if !i.bench {
		maxgo.Pretty("handle", msg, inlet, data)
	}

	// send to first outlet
	i.out1.Any(msg, data)
}

func (i *instance) Free() {
	// check bench
	if !i.bench {
		maxgo.Pretty("free")

		// stop ticker
		i.tick.Stop()
	}
}

func init() {
	// initialize Max class
	maxgo.Register("maxgo", &instance{})
}

func main() {
	// not called
}
