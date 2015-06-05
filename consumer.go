package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/jroimartin/rpcmq"
)

func main() {
	s := rpcmq.NewServer("amqp://localhost:5672",
		"rcp-queue", "rpc-exchange", "fanout")
	if err := s.Register("httpBruter", httpBruter); err != nil {
		log.Fatalf("Register: %v", err)
	}
	if err := s.Init(); err != nil {
		log.Fatalf("Init: %v", err)
	}
	defer s.Shutdown()

	<-time.After(60 * time.Second)
}

func httpBruter(id string, data []byte) ([]byte, error) {
	var wg sync.WaitGroup
	type JsonUrls struct {
		Urls []string
	}
	// This will receive a url array in json format: { "urls": [ "url1", "url2" ] }
	var urlsS JsonUrls
	json.Unmarshal(data, &urlsS)
	urls := urlsS.Urls
	log.Printf("Received (%v): %v urls.\n", id, len(urls))
	// results := []string{}
	var tmpResults JsonUrls
	// Iterate over the received urls
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	var simult = make(chan string, 100)
	for _, u := range urls {
		simult <- u
		go func() {
			wg.Add(1)
			fmt.Println("Querying: ", u)
			req, err := http.NewRequest("GET", u, nil)
			req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/43.0.2357.65 Safari/537.36")
			response, err := client.Do(req)
			if err != nil {
				log.Println(err)
			} else {
				if response.StatusCode == 200 {
					tmpResults.Urls = append(tmpResults.Urls, u)
				}
			}
			<-simult
			defer wg.Done()
		}()
	}
	// This must return an array of 200'd urls
	wg.Wait()
	fmt.Println("Returning " + string(len(tmpResults.Urls)) + " urls.")
	results, _ := json.Marshal(tmpResults)
	return []byte(results), nil
}
