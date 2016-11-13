package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Sirupsen/logrus"
)

func main() {
	flag.Parse()
	if printVersion {
		fmt.Printf("peskar-index %s\n", Version)
		os.Exit(0)
	}

	if err := initConfig(); err != nil {
		logrus.Fatal(err.Error())
	}

	logrus.Info("Starting peskar-index")

	hub := NewHub(&config)
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
