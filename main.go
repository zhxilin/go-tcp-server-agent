package main

import (
	"go-tcp-server-agent/core"
	"go-tcp-server-agent/core/model"
	"os"
	"os/signal"
)

func main() {
	cfg := &model.Config{
		Id:   1000,
		Host: "127.0.0.1",
		Port: 1922,
	}

	if agent := core.NewTcpServer(cfg); agent != nil {
		agent.Run()
		defer agent.Shutdown()
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
}
