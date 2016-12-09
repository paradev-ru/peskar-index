package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/leominov/peskar-hub/peskar"
	"github.com/leominov/peskar-index/lib"
)

type Hub struct {
	Name   string
	redis  *lib.RedisStore
	config *Config
}

func NewHub(name string, config *Config) *Hub {
	redis := lib.NewRedis(config.RedisMaxIdle, config.RedisIdleTimeout, config.RedisAddr)
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "na"
	}
	return &Hub{
		Name:   fmt.Sprintf("%s-%s", name, hostname),
		redis:  redis,
		config: config,
	}
}

func (h *Hub) Log(jobID, message string) error {
	l := peskar.LogItem{
		Initiator: h.Name,
		JobID:     jobID,
		Message:   message,
	}
	logrus.Infof("%s: %s", jobID, message)
	return h.redis.Send(peskar.JobLogChannel, l)
}

func (h *Hub) SuccessReceived(result []byte) error {
	var job peskar.Job
	if err := json.Unmarshal(result, &job); err != nil {
		return fmt.Errorf("Unmarshal error: %v (%s)", err, string(result))
	}
	if job.State != "finished" {
		return nil
	}

	h.Log(job.ID, "Got a job")
	p := job.Directory()
	movieTarball := filepath.Join(h.config.ResultDir, p+".tar")
	if _, err := os.Stat(movieTarball); os.IsNotExist(err) {
		h.Log(job.ID, fmt.Sprintf("Error: %v", err))
		return err
	}

	h.Log(job.ID, "Parsing and creating HTML page...")
	err := SaveAsHTML(job, h.config.TemplateDir, h.config.ResultDir)
	if err != nil {
		h.Log(job.ID, fmt.Sprintf("Error: %v", err))
		return err
	}
	h.Log(job.ID, "HTML page created")

	h.Log(job.ID, "Extracting files from an archive...")
	err = Untar(movieTarball, path.Join(h.config.ResultDir, p))
	if err != nil {
		h.Log(job.ID, fmt.Sprintf("Error extracting: %v", err))
		return err
	}
	h.Log(job.ID, "Files extracted")

	if err := os.Remove(movieTarball); err != nil {
		h.Log(job.ID, fmt.Sprintf("Can't delete archive: %v", err))
	} else {
		h.Log(job.ID, "Archive deleted")
	}

	h.Log(job.ID, "Done")

	return nil
}

func (h *Hub) RetryingPolicy(attempts int, duration time.Duration) error {
	logrus.Debugf("Wait Redis for a 10 seconds (#%d, %v)", attempts, duration)
	time.Sleep(10 * time.Second)
	return nil
}

func (h *Hub) Run() error {
	logrus.Info("Waiting for incoming events...")
	sub := h.redis.NewSubscribe(peskar.JobEventsChannel)
	sub.SuccessReceivedCallback = h.SuccessReceived
	sub.RetryingPolicyCallback = h.RetryingPolicy
	return sub.Run()
}
