package main

import (
	"bytes"
	"io"
	"net/http"
	"sync/atomic"
	"text/template"
	"time"
)

type RequestBodyVars struct {
	Id  int32
	Now time.Time
}

type RequestFactory interface {
	NewRequest(method, url string, body io.Reader) (*http.Request, error)
}

type reqFactory struct {
	template *template.Template
	nextId   int32
}

func NewRequestFactory(tmpl string) (*reqFactory, error) {
	fac := new(reqFactory)
	if len(tmpl) > 0 {
		var err error
		fac.template, err = template.New("reqfactmpl").Parse(tmpl)
		if err != nil {
			return nil, err
		}
	}
	return fac, nil
}

func NewRequestFactoryFromFile(fn string) (RequestFactory, error) {
	fac := new(reqFactory)
	if len(fn) > 0 {
		var err error
		fac.template, err = template.New("reqfactmpl").ParseFiles(fn)
		if err != nil {
			return nil, err
		}
	}
	return fac, nil
}

func (self *reqFactory) NewRequest(method, url string, body io.Reader) (req *http.Request, err error) {
	if body == nil && self.template != nil {
		vars := new(RequestBodyVars)
		vars.Now = time.Now()
		vars.Id = atomic.AddInt32(&self.nextId, 1)

		buf := &bytes.Buffer{}
		err = self.template.Execute(buf, vars)
		if err != nil {
			return
		}
		body = bytes.NewBuffer(buf.Bytes())
	}
	return http.NewRequest(method, url, body)
}
