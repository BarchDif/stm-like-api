package main

import (
	"context"
	"github.com/BarchDif/stm-like-api/internal/app/retranslator"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	rootContext := context.Background()
	cancelCtx, cancelFunc := context.WithCancel(rootContext)
	sigs := make(chan os.Signal, 1)

	cfg := retranslator.Config{
		ChannelSize:   512,
		ConsumerCount: 2,
		ConsumeSize:   10,
		ProducerCount: 28,
		WorkerCount:   2,
	}

	retranslator := retranslator.NewRetranslator(cfg)
	retranslator.Start(cancelCtx)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs
	cancelFunc()

	shutdownCtx, _ := context.WithTimeout(rootContext, time.Second)
	select {
	case <-retranslator.Stopped():
	case <-shutdownCtx.Done():
	}

}
