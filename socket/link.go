package socket

import (
	"bufio"
	"context"
	"net"

	"github.com/pkg/errors"
)

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

