package mcutil_test

import (
	"context"
	"testing"

	"github.com/mcstatus-io/mcutil/v3"
)

func TestStatusRaw(t *testing.T) {
	resp, err := mcutil.StatusRaw(context.Background(), "demo.mcstatus.io", mcutil.DefaultJavaPort)

	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%+v\n", resp)
}
