package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type ServerHealthCheck struct {
	interval                time.Duration
	servers                 []string
	healthCheckResults      sync.Map
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
	for {
		hc.printServersStatus()
		time.Sleep(hc.interval)
		// time.Ticker
	}
}

func (hc *ServerHealthCheck) printServersStatus() {
	var wg *sync.WaitGroup
	// n, g
	// for n/g {
	wg.Add(len(hc.servers))
	for _, server := range hc.servers {
		go func() {
			defer wg.Done()
			hc.getServerStaus(server)
		}()
	}
	wg.Wait()
	// }

	//print
}

func (hc *ServerHealthCheck) getServerStaus(server string) {
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
	}
	defer resp.Body.Close()

	//read resp body

	result := &HealthCheckResult{
		server:  server,
		port:    req.URL.Port(),
		status:  resp.StatusCode,
		latency: tend.Sub(tStart),
	}

	hc.healthCheckResults.Store(result, struct{}{})
}

func main() {
	// host1.api.com, host2.api.com:8080
	// read servers from file
	timeout := 5 * time.Second
	client := &http.Client{Timeout: timeout}
	interval := 5 * time.Minute
	healthCheckEndpoint := "/healthcheck"
	port := "8080"
	servers := []string{}
	healthCheck := &ServerHealthCheck{interval: interval, servers: servers, client: client,
		healthCheckEndpoint: healthCheckEndpoint, healthCheckEndpointPort: port}
	healthCheck.Start()
}
