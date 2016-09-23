package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

var clients []http.Client

type info map[string][]interface{}

type result struct {
	e      error
	tag    string
	result interface{}
}

type url struct {
	url string
	tag string
}

var (
	timeOut        = flag.Int64("t", 10, "time val")
	parallels      = flag.Int("p", 1, "parallels range 1..4")
	beaconInterval = flag.Int64("b", 10, "beacon interval")
	debug          = flag.Bool("debug", false, "enable debug mode")
)

func initialize() {
	for i := 0; i < *parallels; i++ {
		clients = append(clients, http.Client{Timeout: time.Duration(*timeOut) * time.Second})
	}
}

func fetchURL(wg *sync.WaitGroup, warkerNum int, urlQueue <-chan url, r chan<- result) {
	defer wg.Done()
	for {
		url, ok := <-urlQueue

		if !ok {
			return
		}
		resp, err := clients[warkerNum].Get(url.url)
		if err != nil {
			r <- result{
				err,
				url.tag,
				nil,
			}
			continue
		}
		body, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			r <- result{
				err,
				url.tag,
				nil,
			}
			continue
		}
		r <- result{
			nil,
			url.tag,
			string(body),
		}
	}
}

func GetInfo() (*info, error) {
	var wg sync.WaitGroup
	var ret info
	ret = map[string][]interface{}{}
	r := make(chan result, 256)
	urls := make(chan url, 256)
	for i := 0; i < *parallels; i++ {
		wg.Add(1)
		go fetchURL(&wg, i, urls, r)
	}
	//urls <- url{"http://httpbin.org/ip", "url"}
	//urls <- url{"http://httpbin.org/ip", "url"}
	if !*debug {
		urls <- url{"http://169.254.169.254/latest/meta-data/placement/availability-zone", "az"}
		urls <- url{"http://169.254.169.254/latest/meta-data/instance-id", "instance-id"}
	} else {
		r <- result{
			nil,
			"az",
			"ap-northeast-1c",
		}

		r <- result{
			nil,
			"instance-id",
			"i-03324aba380510e93",
		}
	}

	close(urls)

	wg.Wait()

	if len(r) == 0 {
		return nil, errors.New("no info")
	}

	for len(r) > 0 {
		res := <-r
		ret[res.tag] = append(ret[res.tag], res.result)
	}

	return &ret, nil
}

func main() {
	flag.Parse()
	initialize()
	info, _ := GetInfo()
	bin, _ := json.Marshal(info)
	fmt.Println(string(bin))
}
