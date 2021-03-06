package main

import (
	"fmt"
	"github.com/kylelemons/go-gypsy/yaml"
	"strconv"
	"time"
)

type ClientProfile struct {
	Name         string
	Distribution string
	Parameters   []float64
	MaxNrReq     int
	MaxDuration  time.Duration
	FirstWait    time.Duration
}

type Profile struct {
	URL      string
	Method   string
	Template string

	Clients []*ClientProfile
}

func (self *ClientProfile) GenerateSleeper() (Sleeper, error) {
	return NewRandomDelay("s", self.Distribution, self.Parameters...)
}

func (self *Profile) GenerateRequestFactory() (RequestFactory, error) {
	return NewRequestFactory(self.Template)
}

func parseDuration(node yaml.Node) (t time.Duration, err error) {
	if scalar, ok := node.(yaml.Scalar); ok {
		t, err = time.ParseDuration(string(scalar))
	} else {
		err = fmt.Errorf("duration should be a scalar")
	}
	return
}

func parseInt(node yaml.Node) (n int, err error) {
	if scalar, ok := node.(yaml.Scalar); ok {
		str := string(scalar)
		n, err = strconv.Atoi(str)
	} else {
		err = fmt.Errorf("Not a scalar")
	}
	return
}

func parseString(node yaml.Node) (str string, err error) {
	if node == nil {
		str = ""
		return
	}
	if scalar, ok := node.(yaml.Scalar); ok {
		str = string(scalar)
	} else {
		err = fmt.Errorf("not a scalar")
	}
	return
}

func parseFloatList(node yaml.Node) (l []float64, err error) {
	if node == nil {
		l = nil
		return
	}
	if list, ok := node.(yaml.List); ok {
		var scalar string
		var f float64
		l = make([]float64, 0, len(list))
		for i, n := range list {
			scalar, err = parseString(n)
			if err != nil {
				err = fmt.Errorf("element %v is not a scalar", i)
				return
			}
			f, err = strconv.ParseFloat(scalar, 64)
			if err != nil {
				err = fmt.Errorf("element %v is not a float64", i)
				return
			}
			l = append(l, f)
		}
	} else {
		err = fmt.Errorf("not a list")
	}
	return
}

func parseClientProfile(clientName string, node yaml.Node) (*ClientProfile, error) {
	if kv, ok := node.(yaml.Map); ok {
		var err error
		prof := new(ClientProfile)
		if dist, ok := kv["distribution"]; ok {
			prof.Distribution, err = parseString(dist)
			if err != nil {
				return nil, fmt.Errorf("client %v's distribution is %v", clientName, err)
			}
		} else {
			return nil, fmt.Errorf("client %v should have a distribution of inter arrival time", clientName)
		}
		if params, ok := kv["parameters"]; ok {
			prof.Parameters, err = parseFloatList(params)
			if err != nil {
				return nil, fmt.Errorf("client %v's parameter error: %v", clientName, err)
			}
		}

		if n, ok := kv["max-number-of-request"]; ok {
			prof.MaxNrReq, err = parseInt(n)
			if err != nil {
				return nil, fmt.Errorf("client %v's max number of request error: %v", clientName, err)
			}
		}

		if d, ok := kv["max-duration"]; ok {
			prof.MaxDuration, err = parseDuration(d)
			if err != nil {
				return nil, fmt.Errorf("client %v's max duration error: %v", clientName, err)
			}
		}
		if w, ok := kv["wait-for"]; ok {
			prof.FirstWait, err = parseDuration(w)
			if err != nil {
				return nil, fmt.Errorf("client %v's wait-for error: %v", clientName, err)
			}
		}
		return prof, nil
	}
	return nil, fmt.Errorf("client %v's profile should be a map", clientName)
}

func parseStringFromMap(kv yaml.Map, key string) (string, error) {
	if str, ok := kv[key]; ok {
		s, err := parseString(str)
		if err != nil {
			return "", fmt.Errorf("%v should be a string", key)
		}
		return s, nil
	}
	return "", fmt.Errorf("cannot find %v", key)
}

func ParseProfile(node yaml.Node) (p *Profile, err error) {
	if kv, ok := node.(yaml.Map); ok {
		var client *ClientProfile
		p = new(Profile)
		p.Clients = make([]*ClientProfile, 0, len(kv)-1)
		p.URL, err = parseStringFromMap(kv, "url")
		if err != nil {
			return
		}

		p.Method, err = parseStringFromMap(kv, "method")
		if err != nil {
			p.Method = "GET"
		}

		p.Template, err = parseStringFromMap(kv, "template")
		if err != nil {
			p.Template = ""
		}
		for k, v := range kv {
			if k == "url" || k == "method" || k == "template" {
				continue
			}
			client, err = parseClientProfile(k, v)
			if err != nil {
				return
			}
			p.Clients = append(p.Clients, client)
		}
	} else {
		return nil, fmt.Errorf("profile should be a map")
	}

	if len(p.Clients) == 0 {
		err = fmt.Errorf("no client")
	}
	return
}

func ParseProfileFile(filename string) (*Profile, error) {
	file, err := yaml.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	root := file.Root
	return ParseProfile(root)
}
