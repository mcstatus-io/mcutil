package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/mcstatus-io/mcutil/v3"
)

var (
	host string
	port uint16
	opts Options = Options{}
)

type Options struct {
	Type    string `short:"t" long:"type" description:"The type of status to retrieve" default:"java"`
	Timeout uint   `short:"T" long:"timeout" description:"The amount of seconds before the status retrieval times out" default:"5"`
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

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(opts.Timeout))

	defer cancel()

	switch opts.Type {
	case "java":
		{
			result, err = mcutil.Status(ctx, host, port)

			break
		}
	case "raw":
		{
			result, err = mcutil.StatusRaw(ctx, host, port)

			break
		}
	case "legacy":
		{
			result, err = mcutil.StatusLegacy(ctx, host, port)

			break
		}
	case "bedrock":
		{
			result, err = mcutil.StatusBedrock(ctx, host, port)

			break
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
