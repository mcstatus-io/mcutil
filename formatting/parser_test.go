package formatting_test

import (
	"testing"

	"github.com/mcstatus-io/mcutil/v4/formatting"
)

func TestFormatting(t *testing.T) {
	res, err := formatting.Parse("\u00A7cT\u00A7ce\u00A7cs\u00A7ct")

	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%+v\n", res.Tree)
}
