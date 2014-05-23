package main

import (
	"flag"
	"fmt"
	"github.com/koofr/go-serviceregistry"
	"os"
	"os/signal"
	"syscall"
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s register service protocol server\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "       %s get service protocol\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {
	var zkServers = flag.String("zk", "localhost:2181", "zookeeper servers")

	flag.Usage = usage
	flag.Parse()

	r, err := serviceregistry.NewZkRegistry(*zkServers)

	if err != nil {
		fmt.Fprintln(os.Stderr, "Registry connection: %s\n", err)
		os.Exit(1)
	}

	defer r.Close()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)
	go func() {
		for _ = range c {
			r.Close()
			os.Exit(1)
		}
	}()

	args := flag.Args()

	if len(args) < 1 {
		usage()
	}

	action := args[0]

	switch action {
	case "register":
		if len(args) < 4 {
			fmt.Fprintln(os.Stderr, "Wrong number of parameters.")
			os.Exit(1)
		}

		service := args[1]
		protocol := args[2]
		server := args[3]

		err := r.Register(service, protocol, server)

		if err != nil {
			fmt.Fprintln(os.Stderr, "Registration failed: %s\n", err)
			os.Exit(1)
		}

		fmt.Fprintln(os.Stderr, "Registered...")

		<-make(chan int)
	case "get":
		if len(args) < 3 {
			fmt.Fprintln(os.Stderr, "Wrong number of parameters.")
			os.Exit(1)
		}

		service := args[1]
		protocol := args[2]

		servers, err := r.Get(service, protocol)

		if err != nil {
			fmt.Fprintln(os.Stderr, "Lookup failed: %s\n", err)
			os.Exit(1)
		}

		for _, server := range servers {
			fmt.Println(server)
		}
	default:
		fmt.Fprintln(os.Stderr, "Invalid action")
		os.Exit(1)
	}
}
