package status_test

import (
	"context"
	"encoding/json"
	"log"
	"testing"

	"github.com/mcstatus-io/mcutil/v4/status"
)

func TestModern(t *testing.T) {
	resp, err := status.Modern(context.Background(), "demo.mcstatus.io")

	if err != nil {
		t.Fatal(err)
	}

	d, _ := json.MarshalIndent(resp, "", "    ")
	log.Printf("%s\n", d)

	// t.Logf("%+v\n", resp)
}
