package socket

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/ardnew/cimon/server"
	"github.com/ardnew/cimon/socket/config"
	"github.com/pkg/errors"
)

var errHaltServerRequest = errors.New("halt server requested")

const hostContextKey ContextKey = "host"

type (
	Peer       struct{ conn net.Conn }
	Host       server.Server[Peer]
	ContextKey string
)

type Socket struct {
	*config.Flags
	list net.Listener
}

func New(flags *config.Flags, name, version string) *Socket {
	log.Println(name, "version", version)
	return &Socket{Flags: flags}
}

func (s *Socket) Open(ctx context.Context) (err error) {
	s.list, err = new(net.ListenConfig).Listen(ctx, "tcp", s.Bind.String())
	// s.list, err = net.Listen("tcp", bind)
	if err == nil {
		log.Printf("ready on %v [%s]", s.Bind.String(), "tcp")
	}
	return errors.WithStack(err)
}

func (s *Socket) Connect(ctx context.Context, clients chan<- Peer) error {
	switch s.Proto {
	case config.TCP:
		return s.connectTCP(ctx, clients)
	case config.HTTP:
		return s.connectHTTP(ctx, clients)
	}
	return errors.WithStack(config.ErrProtocol)
}

func (s *Socket) Respond(_ context.Context, client Peer) server.Task {
	switch s.Proto {
	case config.TCP:
		return func(ctx context.Context, cancel context.CancelFunc) error {
			return s.respondTCP(ctx, cancel, client)
		}
	case config.HTTP:
		return func(ctx context.Context, cancel context.CancelFunc) error {
			return s.respondHTTP(ctx, cancel, client)
		}
	}
	return nil
}

func (s *Socket) connectTCP(_ context.Context, clients chan<- Peer) error {
	defer close(clients)
	conn, err := s.list.Accept()
	if err != nil {
		return errors.WithStack(err)
	}
	clients <- Peer{conn: conn}
	return nil
}

func (s *Socket) connectHTTP(ctx context.Context, clients chan<- Peer) (err error) {
	defer close(clients)
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(rsp http.ResponseWriter, req *http.Request) {
		// The "/" pattern matches everything.
		if req.URL.Path != "/" {
			http.NotFound(rsp, req)
			return
		}
		fmt.Fprintf(rsp, "ready")
	})
	log.Printf("http[%v]: service start", s.Bind.String())
	serv := (&http.Server{
		Addr:    s.Bind.String(),
		Handler: mux,
		BaseContext: func(l net.Listener) context.Context {
			ctx = context.WithValue(ctx, hostContextKey, l.Addr().String())
			return ctx
		},
		ReadTimeout:       s.Timeout,
		ReadHeaderTimeout: s.HeaderTimeout,
	})
	err = serv.Serve(s.list)
	if err != nil {
		log.Printf("ERROR[%v]: %v", s.Bind.String(), err)
	}
	log.Printf("http[%v]: service stop", s.Bind.String())
	return errors.WithStack(err)
}

func (s *Socket) respondTCP(ctx context.Context, cancel context.CancelFunc, client Peer) error {
	lnk := newLink(client.conn)
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

func (s *Socket) respondHTTP(context.Context, context.CancelFunc, Peer) error {
	return nil
}

func (s *Socket) handle(_ context.Context, lnk *link, message string) error {
	req := strings.TrimSpace(message)
	switch req {
	case "halt":
		return errHaltServerRequest
	}
	log.Printf("req[%v]: %s", lnk.host, req)
	_, err := lnk.conn.Write([]byte(message))
	return errors.WithStack(err)
}
