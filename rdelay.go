package main

import (
	"code.google.com/p/probab/dst"
	"fmt"
	"strings"
	"time"
)

type Sleeper interface {
	Sleep() error
}

type randomDelay struct {
	rgen func() float64 // random number generator
	unit string         // unit of time (default: second)
}

func NewRandomDelay(unit, dist string, params ...float64) (Sleeper, error) {
	ret := new(randomDelay)
	ret.unit = unit
	switch strings.ToLower(dist) {
	case "poisson":
		fallthrough
	case "exp":
		lambda := 0.5
		if len(params) > 0 {
			lambda = params[0]
		}
		ret.rgen = dst.Exponential(lambda)
	case "const":
		d := 0.5
		if len(params) > 0 {
			d = params[0]
		}
		ret.rgen = func() float64 {
			return d
		}
	default:
		return nil, fmt.Errorf("unknown distribution type")
	}
	return ret, nil
}

func (self *randomDelay) Sleep() error {
	t := float64(500.0)
	if self.rgen != nil {
		t = self.rgen()
	}
	if len(self.unit) == 0 {
		self.unit = "s"
	}
	d, err := time.ParseDuration(fmt.Sprintf("%v%v", t, self.unit))
	if err != nil {
		return err
	}
	time.Sleep(d)
	return nil
}
