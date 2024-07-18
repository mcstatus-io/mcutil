package status_test

import (
	"context"
	"testing"

	"github.com/mcstatus-io/mcutil/v4/status"
)

func TestModernRaw(t *testing.T) {
	resp, err := status.ModernRaw(context.Background(), "demo.mcstatus.io")

	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%+v\n", resp)
}
