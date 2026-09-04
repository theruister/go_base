package main

import (
	"os"
	"os/signal"
	"syscall"

	svc "bitbucket.org/theruister/goBase/internal"
)

func main() {

	// cleanup on stop signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sigs
		svc.ShutdownSvr()
		os.Exit(1)
	}()

	svc.InitLogger()

	svc.RunRPC()
	svc.ShutdownSvr()
}
