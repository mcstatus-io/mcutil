package mcutil_test

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/mcstatus-io/mcutil/v2"
	"github.com/mcstatus-io/mcutil/v2/options"
)

func TestLegacyVote(t *testing.T) {
	err := mcutil.SendVote(context.Background(), "localhost", 8192, options.Vote{
		Token:       "MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAu4dM/XxWjkmOz2jEe90B6wA8nMPlzSZmM0mJMFCjWG1e1IM7diXcnnpMLN0jclrceF2imm1z+c0J/LMXXs9OjLyUUdemIAs5ErAfaMpGhB+fEs8MJpLoH5PRXboHQ4c+jmKMyck8oUWeX5vQ4OG4zPBcm0IXKHgT8qu0O5lmroZ8gQGTzdL1NF+B7ws7EXikb03q/3yxvj288X0Fyl5fvrSUcSGxYCtZAyp1MotbrNG/FWlUnJEY8Nqona3tAEsAGp7gWT5SYeOM4BQuVBpr8HQ74odTgLFPNtmS2d9ZQ9Q4m19lkugD9jdjsKSykc4gNg7Y7jhzpUhWptR9GNhrUQIDAQAB",
		ServiceName: "Test",
		Username:    "PassTheMayo",
		Timestamp:   time.Now(),
		Timeout:     time.Second * 5,
	})

	log.Printf("%+v\n", err)
}
