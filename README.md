# mcstatus
![](https://img.shields.io/github/languages/code-size/mcstatus-io/mcutil)
![](https://img.shields.io/github/issues/mcstatus-io/mcutil)
![](https://img.shields.io/github/license/mcstatus-io/mcutil)
![](https://img.shields.io/uptimerobot/ratio/m790234681-f7dbf52397f005c0699ac797)

A library for interacting with the Minecraft API in Go.

## Installation

```bash
go get github.com/mcstatus-io/mcutil
```

## Documentation

https://pkg.go.dev/github.com/mcstatus-io/mcutil

## Usage

### Status (1.7+)

```go
import "github.com/mcstatus-io/mcutil"

func main() {
    response, err := mcutil.Status("play.hypixel.net", 25565)

    if err != nil {
        panic(err)
    }

    fmt.Println(response)
}
```

### Legacy Status (< 1.7)

```go
import "github.com/mcstatus-io/mcutil"

func main() {
    response, err := mcutil.StatusLegacy("play.hypixel.net", 25565)

    if err != nil {
        panic(err)
    }

    fmt.Println(response)
}
```

### Bedrock Status

```go
import "github.com/mcstatus-io/mcutil"

func main() {
    response, err := mcutil.StatusBedrock("127.0.0.1", 19132)

    if err != nil {
        panic(err)
    }

    fmt.Println(response)
}
```

### Basic Query

```go
import "github.com/mcstatus-io/mcutil"

func main() {
    response, err := mcutil.BasicQuery("play.hypixel.net", 25565)

    if err != nil {
        panic(err)
    }

    fmt.Println(response)
}
```

### Full Query

```go
import "github.com/mcstatus-io/mcutil"

func main() {
    response, err := mcutil.FullQuery("play.hypixel.net", 25565)

    if err != nil {
        panic(err)
    }

    fmt.Println(response)
}
```

### RCON

```go
import "github.com/mcstatus-io/mcutil"

func main() {
    client := mcutil.NewRCON()

    if err := client.Dial("127.0.0.1", 25575); err != nil {
        panic(err)
    }

    if err := client.Login("mypassword"); err != nil {
        panic(err)
    }

    if err := client.Run("say Hello, world!"); err != nil {
        panic(err)
    }

    fmt.Println(<- client.Messages)

    if err := client.Close(); err != nil {
        panic(err)
    }
}
```

## Send Vote

```go
import (
    "github.com/mcstatus-io/mcutil"
    "github.com/mcstatus-io/mcutil/options"
)

func main() {
    err := mcutil.SendVote("127.0.0.1", 8192, options.Vote{
		ServiceName: "my-service",
		Username:    "PassTheMayo",
		Token:       "abc123", // server's Votifier token
		UUID:        "",       // recommended but not required, UUID with dashes
		Timestamp:   time.Now(),
		Timeout:     time.Second * 5,
	})

    if err != nil {
        panic(err)
    }
}
```