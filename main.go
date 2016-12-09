package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Sirupsen/logrus"
)

const (
	BaseName = "peskar-index"
)

func main() {
	flag.Parse()
	if printVersion {
		fmt.Printf("%s %s\n", BaseName, Version)
		os.Exit(0)
	}

	if err := initConfig(); err != nil {
		logrus.Fatal(err.Error())
	}

	logrus.Infof("Starting %s", BaseName)

	hub := NewHub(BaseName, &config)
	if err := hub.redis.Check(); err != nil {
		logrus.Errorf("Error creating redis connection: %+v", err)
		os.Exit(1)
	}

	go hub.Run()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case sign := <-signalChan:
			logrus.Info(fmt.Sprintf("Captured %v. Exiting...", sign))
			os.Exit(0)
		}
	}
}
