package status_test

import (
	"context"
	"testing"

	"github.com/mcstatus-io/mcutil/v4/status"
)

func TestModern(t *testing.T) {
	resp, err := status.Modern(context.Background(), "demo.mcstatus.io")

	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%+v\n", resp)
}
