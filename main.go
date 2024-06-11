package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"

	"github.com/ardnew/cimon/server"
	"github.com/ardnew/cimon/socket"
)

func main() {

	log.SetPrefix("-·- ")
	log.SetFlags(log.Lmicroseconds|log.Lmsgprefix)

	ctx, cancel := context.WithCancel(context.Background())
	group, ctx := errgroup.WithContext(ctx)

	group.Go(func() error { return server.Run(ctx, socket.New()) })

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigs:
		log.Printf("caught %[1]v (%#[1]v): closing all connections", sig)
		cancel()
	case <-ctx.Done():
	}

	err := group.Wait()
	if err != nil {
		log.Print(err)
	}
}
