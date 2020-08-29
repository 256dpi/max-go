package main

import (
	"time"

	"github.com/256dpi/maxgo"
	"github.com/256dpi/maxgo/max"
)

type instance struct {
	in1  *max.Inlet
	in2  *max.Inlet
	out1 *max.Outlet
	out2 *max.Outlet
	tick *time.Ticker
}

func (i *instance) Init(obj *max.Object, args []max.Atom) {
	max.Pretty("init", args)

	// declare inlets
	i.in1 = obj.Inlet(max.Any, "example inlet 1", true)
	i.in2 = obj.Inlet(max.Float, "example inlet 2", false)

	// declare outlets
	i.out1 = obj.Outlet(max.Any, "example outlet 1")
	i.out2 = obj.Outlet(max.Bang, "example outlet 2")

	// create timer
	i.tick = time.NewTicker(time.Second)

	// send a bang for every tick
	go func() {
		var j int
		for range i.tick.C {
			max.Pretty("tick", max.IsMainThread(), max.IsTimerThread())

			// bang immediately or defer
			if j++; j%2 == 0 {
				i.out2.Bang()
			} else {
				max.Defer(func() {
					max.Pretty("defer", max.IsMainThread(), max.IsTimerThread())
					i.out2.Bang()
				})
			}
		}
	}()
}

func (i *instance) Handle(msg string, inlet int, data []max.Atom) {
	max.Pretty("handle", msg, inlet, data)

	// send to first outlet
	i.out1.Any(msg, data)
}

func (i *instance) Free() {
	max.Pretty("free")

	// stop ticker
	i.tick.Stop()
}

func init() {
	// initialize Max class
	maxgo.Init("maxgo", &instance{})
}

func main() {
	// not called
}
