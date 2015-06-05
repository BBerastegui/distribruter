package main

import (
	"crypto/rand"
	"encoding/json"
	"flag"
	"log"
	"os"
	"regexp"
	"time"

	"github.com/jroimartin/rpcmq"
)

var TargetUrl string

func main() {
	// Command line options
	urlPtr := flag.String("u", "", "Target url.")
	flag.Parse()
	if *urlPtr == "" {
		panic("[!] Empty url.")
		os.Exit(1)
	}
	TargetUrl = *urlPtr
	// END Command line options

	c := rpcmq.NewClient("amqp://localhost:5672",
		"rcp-queue", "rpc-client", "rpc-exchange", "fanout")
	if err := c.Init(); err != nil {
		log.Fatalf("Init: %v", err)
	}
	defer c.Shutdown()

	// Keep getting results
	go func() {
		for r := range c.Results() {
			if r.Err != "" {
				log.Printf("Received error: %v (%v)", r.Err, r.UUID)
				continue
			}
			log.Printf("Received: %v (%v)\n", string(r.Data), r.UUID)
		}
	}()

	go feedHttpBruter(*c)

	<-time.After(600 * time.Second)
}

func feedHttpBruter(c rpcmq.Client) {
	// Get the data needed to call httpBruter
	data := generateUrls(TargetUrl)
	// Call to httpBruter()
	uuid, err := c.Call("httpBruter", data, 0)
	if err != nil {
		log.Println("Call:", err)
	}
	log.Printf("Sent: httpBruter(%v) (%v)\n", string(data), uuid)
	//<-time.After(500 * time.Millisecond)
}

func generateUrls(urlTemplate string) []byte {
	// urlTemplate string must be : http://localhost/admin/*/example
	// This must return []byte(`{"urls":["http://"]}`)
	type JsonUrls struct {
		Urls []string
	}
	var urlsS JsonUrls
	// Detect the injection point:
	injRegex := regexp.MustCompile("(.*)\\*(.*)")
	match := injRegex.FindStringSubmatch(urlTemplate)

	// Generate URLs (Random from pattern)
	for _, u := range randomString("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz1234567890", 44, 1000) {
		urlsS.Urls = append(urlsS.Urls, match[1]+u+match[2])
	}
	results, _ := json.Marshal(urlsS)
	return []byte(results)
}

func randomString(charset string, size int, num int) []string {
	var results []string
	for i := 0; i < num; i++ {
		var bytes = make([]byte, size)
		rand.Read(bytes)
		for k, v := range bytes {
			bytes[k] = charset[v%byte(len(charset))]
		}
		// Append results
		results = append(results, string(bytes))
	}
	return results
}
