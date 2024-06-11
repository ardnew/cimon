package cimon

import (
	"bufio"
	"context"
	"net"
	"strings"
	"log"

	"github.com/pkg/errors"
)

var (
	errHaltServerRequest = errors.New("halt server requested")
)

type Socket struct {
	list net.Listener
}

func NewSocket() Server[net.Conn] {
	return &Socket{}
}

func (s *Socket) Open() (err error) {
	s.list, err = net.Listen("tcp", ":8080")
	return
}

func (s *Socket) Connect() (net.Conn, error) {
	conn, err := s.list.Accept()
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (s *Socket) Respond(conn net.Conn) Task {
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

type link struct {
	conn net.Conn
	host net.Addr
	read *bufio.Reader
	errs chan error
	closed bool
}

func newLink(conn net.Conn) *link {
	return &link{
		conn: conn,
		host: conn.RemoteAddr(),
		read: bufio.NewReader(conn),
		errs: make(chan error),
	}
}

func (l *link) drop() {
	if l.closed {
		return
	}
	l.conn.Close()
	close(l.errs)
	l.closed = true
}

func (l *link) readLine(ctx context.Context, recv chan<- string) {
	defer close(recv)
	if l.closed {
		l.errs <- errors.New("read on closed connection")
		return
	}
	str, err := l.read.ReadString('\n')
	if ctx.Err() == nil {
		if err != nil {
			l.errs <- err
		}
		recv <- str
	}
}

