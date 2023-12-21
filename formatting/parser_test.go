package formatting_test

import (
	"encoding/json"
	"log"
	"testing"

	"github.com/mcstatus-io/mcutil/v3/formatting"
)

func TestParser(t *testing.T) {
	result, err := formatting.Parse("\u00A7c\u00A7lHello world!\nThis is a test.")

	if err != nil {
		t.Fatal(err)
	}

	log.Printf("%s\n", result.HTML)

	data, _ := json.MarshalIndent(result.Tree, "", "    ")
	log.Printf("%s\n", data)
}
