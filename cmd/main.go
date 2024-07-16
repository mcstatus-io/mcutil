package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/mcstatus-io/mcutil/v4/options"
	"github.com/mcstatus-io/mcutil/v4/status"
)

var (
	host string
	opts Options = Options{}
)

type Options struct {
	Type       string `short:"t" long:"type" description:"The type of status to retrieve" default:"java"`
	Timeout    uint   `short:"T" long:"timeout" description:"The amount of seconds before the status retrieval times out" default:"5"`
	DisableSRV bool   `short:"S" long:"disable-srv" description:"Disables SRV lookup"`
}

func init() {
	args, err := flags.Parse(&opts)

	if err != nil {
		if flags.WroteHelp(err) {
			os.Exit(0)

			return
		}

		panic(err)
	}

	if len(args) < 1 {
		fmt.Println("missing host argument")

		os.Exit(1)
	}

	host = args[0]
}

func main() {
	var (
		result interface{}
		err    error
	)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(opts.Timeout))

	defer cancel()

	switch opts.Type {
	case "java":
		{
			result, err = status.Modern(ctx, host, options.JavaStatus{
				EnableSRV:       !opts.DisableSRV,
				Timeout:         time.Duration(opts.Timeout) * time.Second,
				ProtocolVersion: 47,
			})

			break
		}
	case "raw":
		{
			result, err = status.ModernRaw(ctx, host, options.JavaStatus{
				EnableSRV:       !opts.DisableSRV,
				Timeout:         time.Duration(opts.Timeout) * time.Second,
				ProtocolVersion: 47,
			})

			break
		}
	case "legacy":
		{
			result, err = status.Legacy(ctx, host, options.JavaStatusLegacy{
				EnableSRV:       !opts.DisableSRV,
				Timeout:         time.Duration(opts.Timeout) * time.Second,
				ProtocolVersion: 47,
			})

			break
		}
	case "bedrock":
		{
			result, err = status.Bedrock(ctx, host, options.BedrockStatus{
				Timeout: time.Duration(opts.Timeout) * time.Second,
			})

			break
		}
	default:
		{
			fmt.Printf("unknown --type value: %s\n", opts.Type)

			return
		}
	}

	if err != nil {
		panic(err)
	}

	out, err := json.MarshalIndent(result, "", "    ")

	if err != nil {
		panic(err)
	}

	fmt.Printf("%s\n", out)
}
