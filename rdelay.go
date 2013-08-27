package main

import (
	"code.google.com/p/probab/dst"
	"fmt"
	"strings"
	"time"
)

type RandomDelay struct {
	rgen func() int64 // random number generator
	unit string       // unit of time (default: ms)
}

func NewRandomDelay(unit, dist string, params ...float64) *RandomDelay {
	ret := new(RandomDelay)
	ret.unit = unit
	switch strings.ToLower(dist) {
	case "poisson":
		lambda := 500.0
		if len(params) > 0 {
			lambda := params[0]
		}
		ret.rgen = dst.Poisson(lambda)
	case "const":
		d := 500.0
		if len(params) > 0 {
			d = params[0]
		}
		ret.rgen = func() int64 {
			return int64(d)
		}
	}
	return ret
}

func (self *RandomDelay) Delay() error {
	t := int64(100)
	if self.rgen != nil {
		t = self.rgen()
	}
	if len(self.unit) == 0 {
		self.unit = "ms"
	}
	d, err := time.ParseDuration(fmt.Sprint("%v%v", t, self.unit))
	if err != nil {
		return err
	}
	time.Sleep(d)
	return nil
}
