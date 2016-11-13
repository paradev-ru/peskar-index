package main

import (
	"encoding/json"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/leominov/peskar-index/lib"
)

const EventsChannel = "jobs"

type Hub struct {
	redis  *lib.RedisStore
	config *Config
}

func NewHub(config *Config) *Hub {
	redis := lib.NewRedis(config.RedisMaxIdle, config.RedisIdleTimeout, config.RedisAddr)

	return &Hub{
		redis:  redis,
		config: config,
	}
}

func (h *Hub) SuccessReceived(result []byte) error {
	var job Job

	if err := json.Unmarshal(result, &job); err != nil {
		return err
	}

	if job.State != "finished" {
		return nil
	}

	logrus.Infof("%s: Got a new job", job.ID)

	p := job.Directory()
	movieTarball := filepath.Join(h.config.ResultDir, p+".tar")
	if _, err := os.Stat(movieTarball); os.IsNotExist(err) {
		return err
	}

	err := job.SaveAsHTML(h.config.TemplateDir, h.config.ResultDir)
	if err != nil {
		return err
	}

	logrus.Infof("%s: Unpacking...", job.ID)
	err = Untar(movieTarball, path.Join(h.config.ResultDir, p))
	if err != nil {
		return err
	}
	logrus.Infof("%s: All done", job.ID)

	return nil
}

func (h *Hub) RetryingPolicy(attempts int, duration time.Duration) error {
	logrus.Debugf("Wait Redis for a 10 seconds (#%d, %v)", attempts, duration)
	time.Sleep(10 * time.Second)

	return nil
}

func (h *Hub) Run() error {
	sub := h.redis.NewSubscribe(EventsChannel)

	sub.SuccessReceivedCallback = h.SuccessReceived
	sub.RetryingPolicyCallback = h.RetryingPolicy

	return sub.Run()
}
