package main

import (
	"time"

	"github.com/256dpi/max-go"
)

type instance struct {
	sigIn1   *max.Inlet
	sigIn2   *max.Inlet
	anyIn    *max.Inlet
	floatIn  *max.Inlet
	intIn    *max.Inlet
	sigOut1  *max.Outlet
	sigOut2  *max.Outlet
	anyOut   *max.Outlet
	floatOut *max.Outlet
	bangOut  *max.Outlet
	tick     *time.Ticker
	bench    bool
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
	i.sigIn1 = obj.Inlet(max.Signal, "signal 1", false)
	i.sigIn2 = obj.Inlet(max.Signal, "signal 2", false)
	i.anyIn = obj.Inlet(max.Any, "any", true)
	i.floatIn = obj.Inlet(max.Float, "float", false)
	i.intIn = obj.Inlet(max.Int, "int", false)

	// declare outlets
	i.sigOut1 = obj.Outlet(max.Signal, "signal 1")
	i.sigOut2 = obj.Outlet(max.Signal, "signal 2")
	i.anyOut = obj.Outlet(max.Any, "any")
	i.floatOut = obj.Outlet(max.Float, "float")
	i.bangOut = obj.Outlet(max.Bang, "bang")

	// bang second outlet from a timer
	if !i.bench {
		// create timer
		i.tick = time.NewTicker(1 * time.Second)

		// send a bang for every tick
		go func() {
			var j int
			for range i.tick.C {
				max.Pretty("tick", max.IsMainThread())

				// bang immediately or defer
				if j++; j%2 == 0 {
					i.bangOut.Bang()
				} else {
					max.Defer(func() {
						max.Pretty("defer", max.IsMainThread())
						i.bangOut.Bang()
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

	// echo, double or triple
	switch inlet {
	case 0, 1:
		// signals
	case 2:
		i.anyOut.Any(msg, data)
	case 3:
		i.floatOut.Float(data[0].(float64) * 2)
	case 4:
		i.floatOut.Float(float64(data[0].(int64) * 3))
	}
}

func (i *instance) Process(input, output [][]float64) {
	// log
	max.Pretty("process", len(input), len(output))

	// scale output
	for i := range input {
		for j := range input[i] {
			output[i][j] = input[i][j] / 2.0
		}
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

func main() {
	// initialize Max class
	max.Register("maxgo", &instance{})
}
