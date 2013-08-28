package main

import (
	"github.com/mreiferson/go-httpclient"
	"net/http"
	"sync"
	"time"
)

type WorkLoadGenerator struct {
	maxNrReq    int
	maxDuration time.Duration
	sleeper     Sleeper
	reqfac      RequestFactory
}

func NewWorkLoadGenerator(maxNrReq int,
	maxDuration time.Duration,
	sleeper Sleeper,
	reqfac RequestFactory) *WorkLoadGenerator {

	ret := &WorkLoadGenerator{
		maxNrReq:    maxNrReq,
		maxDuration: maxDuration,
		sleeper:     sleeper,
		reqfac:      reqfac,
	}
	return ret
}

type ResponseInfo struct {
	Response *http.Response
	Duration time.Duration
	Error    error
}

func (self *WorkLoadGenerator) Start(method, url string, resChan chan<- *ResponseInfo) {
	var deadline time.Time
	if self.maxDuration > 0*time.Second {
		deadline = time.Now().Add(self.maxDuration)
	}

	transport := &httpclient.Transport{
		ConnectTimeout: 10 * time.Second,
	}
	defer transport.Close()

	client := &http.Client{Transport: transport}

	wg := &sync.WaitGroup{}
	for i := 0; i < self.maxNrReq || self.maxNrReq <= 0; i++ {
		self.sleeper.Sleep()
		if !deadline.IsZero() && time.Now().After(deadline) {
			break
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			respInfo := new(ResponseInfo)
			req, err := self.reqfac.NewRequest(method, url, nil)
			if err != nil {
				respInfo.Error = err
			} else {
				respInfo.Response, respInfo.Error = client.Do(req)
			}
			resChan <- respInfo
		}()
	}
	wg.Wait()
}
