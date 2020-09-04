package main

import (
	"time"

	"github.com/256dpi/max-go"
)

type instance struct {
	in1   *max.Inlet
	in2   *max.Inlet
	in3   *max.Inlet
	out1  *max.Outlet
	out2  *max.Outlet
	out3  *max.Outlet
	tick  *time.Ticker
	bench bool
}

func (i *instance) Init(obj *max.Object, args []max.Atom) bool {
	// check if doomed
	if len(args) > 0 && args[0] == "fail" {
		return false
	}

	// check if benchmark
	if len(args) > 0 && args[0] == "bench" {
		i.bench = true
	}

	// check bench
	if !i.bench {
		max.Pretty("init", args)
	}

	// declare inlets
	i.in1 = obj.Inlet(max.Any, "example inlet 1", true)
	i.in2 = obj.Inlet(max.Float, "example inlet 2", false)
	i.in3 = obj.Inlet(max.Int, "example inlet 3", false)

	// declare outlets
	i.out1 = obj.Outlet(max.Any, "example outlet 1")
	i.out2 = obj.Outlet(max.Float, "example outlet 2")
	i.out3 = obj.Outlet(max.Bang, "example outlet 3")

	// bang second outlet from a timer
	if !i.bench {
		// create timer
		i.tick = time.NewTicker(time.Second)

		// send a bang for every tick
		go func() {
			var j int
			for range i.tick.C {
				max.Pretty("tick", max.IsMainThread())

				// bang immediately or defer
				if j++; j%2 == 0 {
					i.out3.Bang()
				} else {
					max.Defer(func() {
						max.Pretty("defer", max.IsMainThread())
						i.out3.Bang()
					})
				}
			}
		}()
	}

	return true
}

func (i *instance) Handle(inlet int, msg string, data []max.Atom) {
	// check bench
	if !i.bench {
		max.Pretty("handle", inlet, msg, data)
	}

	// echo or double
	switch inlet {
	case 0:
		i.out1.Any(msg, data)
	case 1:
		i.out2.Float(data[0].(float64) * 2)
	case 2:
		i.out2.Float(float64(data[0].(int64) * 3))
	}
}

func (i *instance) Loaded() {
	// check bench
	if !i.bench {
		max.Pretty("loaded")
	}
}

func (i *instance) DoubleClicked() {
	// check bench
	if !i.bench {
		max.Pretty("double clicked")
	}
}

func (i *instance) Free() {
	// check bench
	if !i.bench {
		max.Pretty("free")

		// stop ticker
		i.tick.Stop()
	}
}

func init() {
	// initialize Max class
	max.Register("maxgo", &instance{})
}

func main() {
	// not called
}
