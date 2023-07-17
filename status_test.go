package mcutil_test

import (
	"fmt"
	"testing"

	"github.com/mcstatus-io/mcutil"
)

func TestBedrock(t *testing.T) {
	status, err := mcutil.StatusBedrock("mc.surocraft.eu", 19132)

	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("%+v\n", status)
}
