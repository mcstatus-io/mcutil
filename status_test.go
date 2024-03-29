package mcutil_test

import (
	"context"
	"testing"

	"github.com/mcstatus-io/mcutil/v3"
)

func TestStatus(t *testing.T) {
	resp, err := mcutil.Status(context.Background(), "demo.mcstatus.io", mcutil.DefaultJavaPort)

	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%+v\n", resp)
}
