package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

type ServerHealthCheck struct {
	interval                time.Duration
	servers                 []string
	client                  *http.Client
	healthCheckEndpoint     string
	healthCheckEndpointPort string
}

type HealthCheckResult struct {
	server  string
	port    string
	status  int
	latency time.Duration
}

func (hc *ServerHealthCheck) Start() {
	ticker := time.NewTicker(hc.interval)
	defer ticker.Stop()
	for {
		hc.printServersStatus()
		//time.Sleep(hc.interval)
		<-ticker.C
	}
}

func (hc *ServerHealthCheck) printServersStatus() {
	var wg sync.WaitGroup
	healthCheckResults := make(chan *HealthCheckResult)
	for _, server := range hc.servers {
		wg.Add(1)
		go func() {
			wg.Done()
			hc.getServerStaus(server, healthCheckResults)
		}()
	}

	wg.Wait()
	close(healthCheckResults)

	for result := range healthCheckResults {
		fmt.Println(result)
	}
}

func (hc *ServerHealthCheck) getServerStaus(server string, healthCheckResults chan *HealthCheckResult) {
	u := &url.URL{
		Host:   server + ":" + string(hc.healthCheckEndpointPort),
		Scheme: "http",
		Path:   hc.healthCheckEndpoint,
	}
	_, err := url.Parse(u.String())
	if err != nil {
		err := fmt.Errorf("invalid server hostname: %s", server)
		log.Println(err)
		return
	}

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		err := fmt.Errorf("error creating healthcheck request for server hostname: %s", server)
		log.Println(err)
		return
	}

	tStart := time.Now()
	resp, err := hc.client.Do(req)
	tend := time.Now()
	if err != nil {
		err := fmt.Errorf("error sending healthcheck request for server hostname: %s, err: %v",
			server, err)
		log.Println(err)
		return
	}
	defer resp.Body.Close()

	//read resp body and discard to free up socket buffer and release connection
	_, err = io.Copy(io.Discard, resp.Body)
	if err != nil {
		err := fmt.Errorf("error discarding the resp body, err: %v", err)
		log.Println(err)
	}

	result := &HealthCheckResult{
		server:  server,
		port:    req.URL.Port(),
		status:  resp.StatusCode,
		latency: tend.Sub(tStart),
	}

	healthCheckResults <- result
}

func main() {
	// host1.api.com:8081
	// host2.api.com:8080
	// read servers from file
	data, err := os.ReadFile("servers_data")
	if err != nil {
		log.Fatalf("unable to read servers file, error: %v", err)
	}
	servers := strings.Split(string(data), "\n")
	//fmt.Printf("... %v, %d", servers, len(servers))
	timeout := 5 * time.Second
	client := &http.Client{Timeout: timeout}
	interval := 5 * time.Minute
	healthCheckEndpoint := "/healthcheck"
	port := "8080"
	healthCheck := &ServerHealthCheck{interval: interval, servers: servers, client: client,
		healthCheckEndpoint: healthCheckEndpoint, healthCheckEndpointPort: port}
	healthCheck.Start()
}
