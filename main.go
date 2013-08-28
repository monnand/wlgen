package main

import (
	"flag"
	"fmt"
	"os"
	"sync"
)

var argvFileName = flag.String("f", "example.yaml", "file name")
var argvOutput = flag.String("o", "", "output (default: stdout")
var argvIncludeBody = flag.Bool("i", false, "include the response body in the output")
var argvPrefix = flag.String("p", "#=#", "prefix for each response in the output")

func main() {
	flag.Parse()

	profile, err := ParseProfileFile(*argvFileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}

	output := os.Stdout
	if len(*argvOutput) > 0 {
		output, err = os.Create(*argvOutput)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			return
		}
	}

	ch := make(chan *ResponseInfo, 1024)

	dump := &Dumper{
		Prefix:      "#=",
		W:           output,
		IncludeBody: *argvIncludeBody,
		ResChan:     ch,
	}

	go dump.Dump()

	reqfac, err := profile.GenerateRequestFactory()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}
	wg := &sync.WaitGroup{}
	defer wg.Wait()

	for _, client := range profile.Clients {
		sleeper, err := client.GenerateSleeper()
		if err != nil {
			fmt.Fprintf(os.Stderr, "client %v: %v\n", client.Name, err)
			return
		}
		wlg := NewWorkLoadGenerator(client.MaxNrReq, client.MaxDuration, sleeper, reqfac)
		wg.Add(1)
		go func() {
			wlg.Start(profile.Method, profile.URL, ch)
			wg.Done()
		}()
	}
}
