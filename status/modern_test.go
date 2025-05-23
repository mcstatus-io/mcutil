package status_test

import (
	"context"
	"testing"

	"github.com/mcstatus-io/mcutil/v4/status"
	"github.com/mcstatus-io/mcutil/v4/util"
)

func TestModern(t *testing.T) {
	resp, err := status.Modern(context.Background(), "demo.mcstatus.io", util.DefaultJavaPort)

	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%+v\n", resp)
}
