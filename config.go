package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
)

const (
	DefaultTemplateDir      = "/opt/peskar/template"
	DefaultResultDir        = "/var/www/peskar"
	DefaultRedisAddr        = "redis://localhost:6379/0"
	DefaultRedisIdleTimeout = 240 * time.Second
	DefaultRedisMaxIdle     = 3
)

var (
	templatedir      string
	resultdir        string
	redisAddr        string
	redisIdleTimeout time.Duration
	redisMaxIdle     int
	logLevel         string
	config           Config
	printVersion     bool
)

type Config struct {
	RedisAddr        string
	RedisIdleTimeout time.Duration
	RedisMaxIdle     int
	LogLevel         string
	TemplateDir      string
	ResultDir        string
}

func init() {
	flag.StringVar(&templatedir, "templatedir", "", "template directory")
	flag.StringVar(&resultdir, "resultdir", "", "result directory")
	flag.StringVar(&redisAddr, "redis-addr", "", "Redis server URL")
	flag.DurationVar(&redisIdleTimeout, "redis-idle-timeout", 0*time.Second, "close Redis connections after remaining idle for this duration")
	flag.IntVar(&redisMaxIdle, "redis-max-idle", 0, "Maximum number of idle connections in the Redis pool")
	flag.StringVar(&logLevel, "log-level", "", "level which confd should log messages")
	flag.BoolVar(&printVersion, "version", false, "print version and exit")
}

func initConfig() error {
	config = Config{
		RedisAddr:        DefaultRedisAddr,
		RedisIdleTimeout: DefaultRedisIdleTimeout,
		RedisMaxIdle:     DefaultRedisMaxIdle,
		TemplateDir:      DefaultTemplateDir,
		ResultDir:        DefaultResultDir,
	}

	processEnv()

	processFlags()

	if config.LogLevel != "" {
		level, err := logrus.ParseLevel(config.LogLevel)
		if err != nil {
			return err
		}
		logrus.SetLevel(level)
	}

	if config.RedisAddr == "" {
		return errors.New("Must specify Redis server URL using -redis-addr")
	}

	if config.RedisIdleTimeout == 0*time.Second {
		return errors.New("Must specify Redis idle timeout using -redis-idle-timeout")
	}

	if config.RedisMaxIdle == 0 {
		return errors.New("Must specify Redis max idle using -redis-max-idle")
	}

	if config.TemplateDir == "" {
		return errors.New("Must specify template directory using -templatedir")
	}

	if _, err := os.Stat(config.TemplateDir); os.IsNotExist(err) {
		return fmt.Errorf("Template directory '%s' does not exist", config.TemplateDir)
	}

	if config.ResultDir == "" {
		return errors.New("Must specify result directory using -resultdir")
	}

	if _, err := os.Stat(config.ResultDir); os.IsNotExist(err) {
		return fmt.Errorf("Result directory '%s' does not exist", config.ResultDir)
	}

	return nil
}

func processEnv() {
	redisAddrEnv := os.Getenv("PESKAR_REDIS_ADDR")
	if len(redisAddrEnv) > 0 {
		config.RedisAddr = redisAddrEnv
	}
	templatedirEnv := os.Getenv("PESKAR_TMPDIR")
	if len(templatedirEnv) > 0 {
		config.TemplateDir = templatedirEnv
	}
	resultdirEnv := os.Getenv("PESKAR_RESULTDIR")
	if len(resultdirEnv) > 0 {
		config.ResultDir = resultdirEnv
	}
}

func processFlags() {
	flag.Visit(setConfigFromFlag)
}

func setConfigFromFlag(f *flag.Flag) {
	switch f.Name {
	case "templatedir":
		config.TemplateDir = templatedir
	case "resultdir":
		config.ResultDir = resultdir
	case "redis-addr":
		config.RedisAddr = redisAddr
	case "redis-idle-timeout":
		config.RedisIdleTimeout = redisIdleTimeout
	case "redis-max-idle":
		config.RedisMaxIdle = redisMaxIdle
	case "log-level":
		config.LogLevel = logLevel
	}
}
