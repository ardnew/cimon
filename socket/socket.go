package socket

import (
	"context"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/ardnew/cimon/server"
)

var (
	errHaltServerRequest = errors.New("halt server requested")
)

var DefaultConfig = Config{
	port: 8080,
}

type Config struct {
	Bind    string
	port    uint
	Verbose bool
}

func (c *Config) Interface() string {
	c.port = DefaultConfig.port
	port := strconv.FormatUint(uint64(c.port), 10)
	addr := strings.TrimSpace(c.Bind)
	n := strings.LastIndex(addr, ":")
	if n >= 0 {
		sub := addr[n+1:]
		addr = addr[:n]
		u64, err := strconv.ParseUint(sub, 10, 16)
		if err == nil {
			c.port, port = uint(u64), sub
		}
	}
	c.Bind = addr + ":" + port
	return c.Bind
}

type Socket struct {
	Config
	list net.Listener
}

func New(config Config, name, version string) server.Server[net.Conn] {
	log.Println(name, "version", version)
	return &Socket{Config: config}
}

func (s *Socket) Open(ctx context.Context) (err error) {
	iface := s.Interface()
	s.list, err = new(net.ListenConfig).Listen(ctx, "tcp", iface)
	//s.list, err = net.Listen("tcp", iface)
	if err == nil {
		log.Printf("ready on %s [%s]", iface, "tcp")
	}
	return
}

func (s *Socket) Connect() (net.Conn, error) {
	conn, err := s.list.Accept()
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (s *Socket) Respond(conn net.Conn) server.Task {
	return func(ctx context.Context, cancel context.CancelFunc) error {
		lnk := newLink(conn)
		defer lnk.drop()
		log.Printf("connect[%v]", lnk.host)
		for !lnk.closed {
			recv := make(chan string)
			go lnk.readLine(ctx, recv)
			select {
			case str := <-recv:
				err := s.handle(ctx, lnk, str)
				if err != nil {
					prefix := "ERROR"
					switch {
					case errors.Is(err, errHaltServerRequest):
						prefix = "HALT"
						cancel()
					default:
					}
					log.Printf("%s[%v]: %v", prefix, lnk.host, err)
					lnk.drop()
				}
			case err := <-lnk.errs:
				log.Printf("%s", err)
				lnk.drop()
			case <-ctx.Done():
				lnk.drop()
			}
		}
		log.Printf("disconnect[%v]", lnk.host)
		return nil
	}
}

func (s *Socket) handle(ctx context.Context, lnk *link, message string) error {
	req := strings.TrimSpace(message)
	switch req {
	case "halt":
		return errHaltServerRequest
	}
	log.Printf("req[%v]: %s", lnk.host, req)
	lnk.conn.Write([]byte(message))
	return nil
}
