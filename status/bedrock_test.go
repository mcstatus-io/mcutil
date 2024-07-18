package status_test

import (
	"context"
	"testing"

	"github.com/mcstatus-io/mcutil/v4/status"
)

func TestBedrock(t *testing.T) {
	resp, err := status.Bedrock(context.Background(), "demo.mcstatus.io")

	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%+v\n", resp)
}
