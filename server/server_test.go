package server_test

import (
	"context"
	"testing"

	"github.com/ardnew/cimon/server"
	"github.com/ardnew/cimon/socket"
)

func TestServerRun(t *testing.T) {
	err := server.Run(context.Background(), socket.New())
	if err != nil {
		t.Fatalf("received error: %v", err)
	}
}
