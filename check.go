package main

import (
	"fmt"
	"github.com/riemann/riemann-go-client"
	"context"
	"time"
	"os/exec"
	"strings"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

type Config struct {
	Riemann_host string
	Services struct {
		Clustermonitor []string
	}
}


func checkWindowsService(host string, service string, ch chan riemanngo.Event) {
	fmt.Println("checking!!")

	ctx, cancel := context.WithTimeout(context.Background(), 45 * time.Second)
	defer cancel()

	command := fmt.Sprintf("sc \\\\%s query %s", host, service)
	cmd := exec.CommandContext(ctx, "cmd", "/C", command)

	out, err := cmd.Output()

	if ctx.Err() == context.DeadlineExceeded {
		ch <- riemanngo.Event{
			Service: "clustermonitor",
			State: "critical",
			Metric: 0,
			Description: "Timed out, is host up?",
			Host: host,
			Ttl: 300,
		}
	}

	if err != nil {
		ch <- riemanngo.Event{
			Service: "clustermonitor",
			State: "critical",
			Metric: 0,
			Description: "Clustermonitor service doesn't exist",
			Host: host,
			Ttl: 300,
		}
	}
	if strings.Contains(string(out), "STOPPED") {
		ch <- riemanngo.Event{
			Service: "clustermonitor",
			State: "warning",
			Metric: 0,
			Description: "Clustermonitor service is stopped",
			Host: host,
			Ttl: 300,
		}
	}
	if strings.Contains(string(out), "RUNNING") {
		ch <- riemanngo.Event{
			Service:     "clustermonitor",
			State:       "ok",
			Metric:      0,
			Description: "Clustermonitor service is running",
			Host:        host,
			Ttl:         300,
		}
	}
}

func main() {
	fmt.Println("hello!")
	data, err := ioutil.ReadFile("hmp.yaml")
	if err != nil {
		log.Fatal(err)
	}

	t := Config{}
	err1 := yaml.Unmarshal(data, &t)
	if err1 != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Println(t.Riemann_host)
	c := riemanngo.NewTcpClient(t.Riemann_host)
	err2 := c.Connect(5)
	if err2 != nil {
		panic(err)
	}

	defer c.Close()

	ch := make(chan riemanngo.Event)



	fmt.Printf("--- t:\n%v\n\n", t.Services.Clustermonitor)



	for _, host := range t.Services.Clustermonitor {
		go checkWindowsService(host, "clustermonitor", ch)
	}


	for i := 0; i < len(t.Services.Clustermonitor); {
		select {
		case result := <-ch:
			fmt.Printf("Got this over the channel %v\n", result)
			riemanngo.SendEvent(c, &result)
			i++
		}
	}
}