package main

import (
	"context"
	"fmt"
	"github.com/riemann/riemann-go-client"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os/exec"
	"strings"
	"time"
)

type Config struct {
	Riemann_host string
	Services     map[string][]string
}

func checkWindowsService(host string, service string, ch chan riemanngo.Event) {

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	command := fmt.Sprintf("sc \\\\%s query %s", host, service)
	cmd := exec.CommandContext(ctx, "cmd", "/C", command)

	out, err := cmd.Output()

	if ctx.Err() == context.DeadlineExceeded {
		ch <- riemanngo.Event{
			Service:     service,
			State:       "critical",
			Metric:      0,
			Description: fmt.Sprintf("Timed out querying service %s. Is host up?", service),
			Host:        host,
			Ttl:         300,
		}
	}

	if err != nil {
		ch <- riemanngo.Event{
			Service:     service,
			State:       "critical",
			Metric:      0,
			Description: fmt.Sprintf("%s service doesn't exist.", service),
			Host:        host,
			Ttl:         300,
		}
	}
	if strings.Contains(string(out), "STOPPED") {
		ch <- riemanngo.Event{
			Service:     service,
			State:       "warning",
			Metric:      0,
			Description: fmt.Sprintf("%s service is stopped.", service),
			Host:        host,
			Ttl:         300,
		}
	}
	if strings.Contains(string(out), "RUNNING") {
		ch <- riemanngo.Event{
			Service:     service,
			State:       "ok",
			Metric:      0,
			Description: fmt.Sprintf("%s service is running.", service),
			Host:        host,
			Ttl:         300,
		}
	}
}

func main() {
	data, err := ioutil.ReadFile("config.yaml")
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

	checkCount := 0
	for service, hosts := range t.Services {
		fmt.Printf("----%v\n", service)

		for _, host := range hosts {
			checkCount++
			go checkWindowsService(host, service, ch)
		}
	}

	for i := 0; i < checkCount; i++ {
		select {
		case result := <-ch:
			fmt.Printf("Got this over the channel %v\n", result)
			riemanngo.SendEvent(c, &result)
			i++
		}
	}
}
