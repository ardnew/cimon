package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/ardnew/cimon/server"
	"github.com/ardnew/cimon/socket"
	"github.com/ardnew/cimon/socket/config"
	"golang.org/x/sync/errgroup"
)

const version = "0.1.0"

func splitExe() (dir, base string) {
	exe, err := os.Executable()
	if err != nil {
		return "", ""
	}
	res, err := filepath.Abs(exe)
	if err != nil {
		res = filepath.Clean(exe)
	}
	return filepath.Dir(res), filepath.Base(res)
}

func main() {
	log.SetPrefix("-·- ")
	log.SetFlags(log.Lmicroseconds | log.Lmsgprefix)

	_, exeBase := splitExe()

	flags := config.New(log.Writer(), exeBase, version)

	if err := flags.Parse(os.Args[1:]); err != nil {
		log.Fatal(err)
	}

	group, ctx := errgroup.WithContext(context.Background())
	ctx, cancel := context.WithCancel(ctx)

	group.Go(func() error {
		return server.Run(ctx, cancel, socket.New(flags, exeBase, version))
	})

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigs:
		log.Printf("caught %[1]v (%#[1]v): closing all connections", sig)
	case <-ctx.Done():
	}
	cancel()

	err := group.Wait()
	if err != nil {
		log.Print(err)
	}
}
