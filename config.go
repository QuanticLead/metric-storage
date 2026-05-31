package main

import (
	"flag"
	"os"
	"strconv"
)

// getEnv returns the value of an environment variable or a fallback default string.
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// getEnvInt returns the value of an environment variable as an integer or a fallback default.
func getEnvInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		if val, err := strconv.Atoi(value); err == nil {
			return val
		}
	}
	return fallback
}

var (
	flagListenAddr  = flag.String("listenAddr", getEnv("LISTEN_ADDR", "localhost:4317"), "The listen address")
	flagMaxMsgSize  = flag.Int("maxReceiveMessageSize", getEnvInt("MAX_RECEIVE_MESSAGE_SIZE", 16777216), "The max message size in bytes")
	flagChannelSize = flag.Int("channelSize", getEnvInt("CHANNEL_SIZE", 1000), "Ingestion channel buffer size")
	flagWorkerCount = flag.Int("workerCount", getEnvInt("WORKER_COUNT", 4), "Number of concurrent background workers")
)

type Config struct {
	ListenAddr            string
	MaxReceiveMessageSize int
	ClickHouseAddr        string
	ClickHouseDB          string
	ClickHouseUser        string
	ClickHousePassword    string
	ChannelSize           int
	WorkerCount           int
}

func LoadConfig() *Config {
	if !flag.Parsed() {
		flag.Parse()
	}

	return &Config{
		ListenAddr:            *flagListenAddr,
		MaxReceiveMessageSize: *flagMaxMsgSize,
		ClickHouseAddr:        getEnv("CLICKHOUSE_ADDR", ""),
		ClickHouseDB:          getEnv("CLICKHOUSE_DB", "default"),
		ClickHouseUser:        getEnv("CLICKHOUSE_USER", "default"),
		ClickHousePassword:    getEnv("CLICKHOUSE_PASSWORD", ""),
		ChannelSize:           *flagChannelSize,
		WorkerCount:           *flagWorkerCount,
	}
}
