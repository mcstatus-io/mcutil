package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/mcstatus-io/mcutil/v4/options"
	"github.com/mcstatus-io/mcutil/v4/status"
)

var (
	host string
	port uint16
	opts passedOptions = passedOptions{}
)

type passedOptions struct {
	Type        string `short:"t" long:"type" description:"The type of status to retrieve" default:"java"`
	Timeout     uint   `short:"T" long:"timeout" description:"The amount of seconds before the status retrieval times out" default:"5"`
	DisableSRV  bool   `short:"S" long:"disable-srv" description:"Disables SRV lookup"`
	Debug       bool   `short:"D" long:"debug" description:"Enables debug printing to the console"`
	Protocol    int    `short:"p" long:"protocol" description:"Sets the protocol version for the status ping (Java Edition only)"`
	DisablePing bool   `short:"P" long:"disable-ping" description:"Disables the extra ping-pong payloads during status retrieval"`
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

	if opts.Protocol == 0 {
		opts.Protocol = 47
	}

	host = args[0]

	if len(args) < 2 {
		switch opts.Type {
		case "java", "legacy", "raw":
			{
				port = 25565

				break
			}
		case "bedrock":
			{
				port = 19132

				break
			}
		default:
			{
				fmt.Printf("unknown --type value: %s\n", opts.Type)
			}
		}
	} else {
		value, err := strconv.ParseUint(args[1], 10, 16)

		if err != nil {
			panic(err)
		}

		port = uint16(value)
	}
}

func main() {
	var (
		result interface{}
		err    error
	)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(opts.Timeout)+time.Millisecond*500)

	defer cancel()

	switch opts.Type {
	case "java":
		{
			result, err = status.Modern(ctx, host, port, options.StatusModern{
				EnableSRV:       !opts.DisableSRV,
				Timeout:         time.Duration(opts.Timeout) * time.Second,
				ProtocolVersion: opts.Protocol,
				Ping:            !opts.DisablePing,
				Debug:           opts.Debug,
			})

			break
		}
	case "raw":
		{
			result, err = status.ModernRaw(ctx, host, port, options.StatusModern{
				EnableSRV:       !opts.DisableSRV,
				Timeout:         time.Duration(opts.Timeout) * time.Second,
				ProtocolVersion: opts.Protocol,
			})

			break
		}
	case "legacy":
		{
			result, err = status.Legacy(ctx, host, port, options.StatusLegacy{
				EnableSRV:       !opts.DisableSRV,
				Timeout:         time.Duration(opts.Timeout) * time.Second,
				ProtocolVersion: opts.Protocol,
			})

			break
		}
	case "bedrock":
		{
			result, err = status.Bedrock(ctx, host, port, options.StatusBedrock{
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
