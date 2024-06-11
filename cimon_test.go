package cimon_test

import (
	"context"
	"testing"

	"github.com/ardnew/cimon"
)

func TestPadroneWork(t *testing.T) {
	err := cimon.Serve(context.Background(), cimon.NewSocket())
	if err != nil {
		t.Fatalf("received error: %v", err)
	}
}
