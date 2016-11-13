package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/leominov/peskar-index/lib"
)

const (
	EventsChannel = "jobs"
	IndexChannel  = "index"
)

type Hub struct {
	redis  *lib.RedisStore
	config *Config
}

type IndexLog struct {
	JobID   string `json:"job_id"`
	Message string `json:"message,omitempty"`
}

func NewHub(config *Config) *Hub {
	redis := lib.NewRedis(config.RedisMaxIdle, config.RedisIdleTimeout, config.RedisAddr)

	return &Hub{
		redis:  redis,
		config: config,
	}
}

func (h *Hub) Log(jobID, message string) error {
	l := IndexLog{
		JobID:   jobID,
		Message: message,
	}
	return h.redis.Send(IndexChannel, l)
}

func (h *Hub) SuccessReceived(result []byte) error {
	var job Job
	if err := json.Unmarshal(result, &job); err != nil {
		return fmt.Errorf("Unmarshal error: %v (%s)", err, string(result))
	}

	if job.State != "finished" {
		return nil
	}

	h.Log(job.ID, "Got a job")
	logrus.Infof("%s: Got a new job", job.ID)

	p := job.Directory()
	movieTarball := filepath.Join(h.config.ResultDir, p+".tar")
	if _, err := os.Stat(movieTarball); os.IsNotExist(err) {
		h.Log(job.ID, fmt.Sprintf("Error: %v", err))
		return err
	}

	err := job.SaveAsHTML(h.config.TemplateDir, h.config.ResultDir)
	if err != nil {
		h.Log(job.ID, fmt.Sprintf("Error: %v", err))
		return err
	}
	h.Log(job.ID, "HTML page created")

	logrus.Infof("%s: Unpacking...", job.ID)
	err = Untar(movieTarball, path.Join(h.config.ResultDir, p))
	if err != nil {
		h.Log(job.ID, fmt.Sprintf("Error: %v", err))
		return err
	}
	logrus.Infof("%s: All done", job.ID)
	h.Log(job.ID, "Tarball unarchived")

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
