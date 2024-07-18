package query_test

import (
	"context"
	"testing"

	"github.com/mcstatus-io/mcutil/v4/query"
)

func TestBasic(t *testing.T) {
	resp, err := query.Basic(context.Background(), "demo.mcstatus.io")

	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%+v\n", resp)
}

func TestFull(t *testing.T) {
	resp, err := query.Full(context.Background(), "demo.mcstatus.io")

	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%+v\n", resp)
}
