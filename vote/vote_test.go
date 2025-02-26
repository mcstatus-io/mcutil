package vote_test

import (
	"context"
	"testing"
	"time"

	"github.com/mcstatus-io/mcutil/v4/options"
	"github.com/mcstatus-io/mcutil/v4/vote"
)

func TestVote(t *testing.T) {
	err := vote.SendVote(context.Background(), "demo.mcstatus.io", 8192, options.Vote{
		ServiceName: "mcutil",
		Username:    "PassTheMayo",
		Token:       "abc123",
		UUID:        "85e5f06e-ff89-4c11-8050-329e8fdc29de",
		IPAddress:   "127.0.0.1",
		Timestamp:   time.Now(),
		Timeout:     time.Second * 5,
	})

	if err != nil {
		t.Fatal(err)
	}
}
