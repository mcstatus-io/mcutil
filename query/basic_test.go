package query_test

import (
	"context"
	"testing"

	"github.com/mcstatus-io/mcutil/v4/query"
	"github.com/mcstatus-io/mcutil/v4/util"
)

func TestBasic(t *testing.T) {
	resp, err := query.Basic(context.Background(), "demo.mcstatus.io", util.DefaultJavaPort)

	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%+v\n", resp)
}
