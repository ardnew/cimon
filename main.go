package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"golang.org/x/sync/errgroup"

	"github.com/ardnew/cimon/server"
	"github.com/ardnew/cimon/socket"
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
	fs := flag.NewFlagSet(exeBase, flag.ExitOnError)
	fs.Usage = func() {
		fmt.Println(exeBase, "version", version, "usage:")
		fmt.Println()
		fmt.Println("  -h               Hi.")
		fmt.Println("  -l [ADDR]:PORT   Bind to interface [ADDR]:PORT.")
		fmt.Println("                    (Omit ADDR to listen on all interfaces.)")
		fmt.Println("  -v               Enable additional, verbose logging.")
		fmt.Println()
	}

	config := socket.DefaultConfig
	fs.StringVar(&config.Bind, "l", config.Bind, "")
	fs.BoolVar(&config.Verbose, "v", config.Verbose, "")
	if err := fs.Parse(os.Args[1:]); err != nil {
		log.Fatal(err)
	}

	group, ctx := errgroup.WithContext(context.Background())
	ctx, cancel := context.WithCancel(ctx)

	group.Go(func() error {
		return server.Run(ctx, cancel, socket.New(config, exeBase, version))
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
