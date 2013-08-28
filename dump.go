package main

import (
	"fmt"
	"io"
)

type Dumper struct {
	Prefix      string
	W           io.Writer
	IncludeBody bool
	ResChan     <-chan *ResponseInfo
}

func (self *Dumper) Dump() {
	for res := range self.ResChan {
		if res.Error != nil {
			fmt.Fprintf(self.W, "%v Error: %v\n", self.Prefix, res.Error)
			if res.Response != nil && res.Response.Body != nil {
				res.Response.Body.Close()
			}
			continue
		}
		fmt.Fprintf(self.W, "%v Duration: %v\n", self.Prefix, res.Duration)
		if self.IncludeBody {
			io.Copy(self.W, res.Response.Body)
			res.Response.Body.Close()
		}
	}
}
