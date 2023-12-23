package mcutil_test

import (
	"context"
	"testing"

	"github.com/mcstatus-io/mcutil/v3"
)

func TestStatusBedrock(t *testing.T) {
	resp, err := mcutil.StatusBedrock(context.Background(), "demo.mcstatus.io", mcutil.DefaultBedrockPort)

	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%+v\n", resp)
}
