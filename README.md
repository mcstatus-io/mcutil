# mcutil
![](https://img.shields.io/github/languages/code-size/mcstatus-io/mcutil)
![](https://img.shields.io/github/issues/mcstatus-io/mcutil)
![](https://img.shields.io/github/license/mcstatus-io/mcutil)

A zero-dependency library for interacting with the Minecraft protocol in Go. Supports retrieving the status of any Minecraft server (Java or Bedrock Edition), querying a server for information, sending remote commands with RCON, and sending Votifier votes. Look at the examples in this readme or search through the documentation instead.

## Installation

```bash
go get github.com/mcstatus-io/mcutil/v4
```

## Documentation

https://pkg.go.dev/github.com/mcstatus-io/mcutil/v4

## Usage

### Status (1.7+)

Retrieves the status of the Java Edition Minecraft server. This method only works on netty servers, which is version 1.7 and above. An attempt to use on pre-netty servers will result in an error.

```go
import (
    "context"
    "fmt"
    "time"

    "github.com/mcstatus-io/mcutil/v4/status"
)

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

    defer cancel()

    response, err := status.Modern(ctx, "demo.mcstatus.io")

    if err != nil {
        panic(err)
    }

    fmt.Println(response)
}
```

### Legacy Status (‹ 1.7)

Retrieves the status of the Java Edition Minecraft server. This is a legacy method that is supported by all servers, but only retrieves basic information. If you know the server is running version 1.7 or above, please use `Status()` instead.

```go
import (
    "context"
    "fmt"
    "time"

    "github.com/mcstatus-io/mcutil/v4/status"
)

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

    defer cancel()

    response, err := status.Legacy(ctx, "demo.mcstatus.io")

    if err != nil {
        panic(err)
    }

    fmt.Println(response)
}
```

### Bedrock Status

Retrieves the status of the Bedrock Edition Minecraft server.

```go
import (
    "context"
    "fmt"
    "time"

    "github.com/mcstatus-io/mcutil/v4/status"
)

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

    defer cancel()

    response, err := status.Bedrock(ctx, "demo.mcstatus.io")

    if err != nil {
        panic(err)
    }

    fmt.Println(response)
}
```

### Basic Query

Performs a basic query lookup on the server, retrieving most information about the server. Note that the server must explicitly enable query for this functionality to work.

```go
import (
    "context"
    "fmt"
    "time"

    "github.com/mcstatus-io/mcutil/v4/query"
)

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

    defer cancel()

    response, err := query.Basic(ctx, "play.hypixel.net")

    if err != nil {
        panic(err)
    }

    fmt.Println(response)
}

```

### Full Query

Performs a full query lookup on the server, retrieving all available information. Note that the server must explicitly enable query for this functionality to work.

```go
import (
    "context"
    "fmt"
    "time"

    "github.com/mcstatus-io/mcutil/v4/query"
)

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

    defer cancel()

    response, err := query.Full(ctx, "play.hypixel.net")

    if err != nil {
        panic(err)
    }

    fmt.Println(response)
}
```

### RCON

Executes remote console commands on the server. You must know the connection details of the RCON server, as well as the password.

```go
import "github.com/mcstatus-io/mcutil/v4/rcon"

func main() {
    client, err := rcon.Dial("127.0.0.1", 25575)

    if err != nil {
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

Sends a Votifier vote to the specified server, typically used by server listing websites. The host and port must be known of the Votifier server, as well as the token or RSA public key generated by the server. This is for use on servers running Votifier 1 or Votifier 2, such as [NuVotifier](https://www.spigotmc.org/resources/nuvotifier.13449/).

```go
import (
    "context"
    "time"

    "github.com/mcstatus-io/mcutil/v4/vote"
    "github.com/mcstatus-io/mcutil/v4/options"
)

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

    defer cancel()

    err := vote.SendVote(ctx, "127.0.0.1", 8192, options.Vote{
        // General
        ServiceName: "my-service",    // Required
        Username:    "PassTheMayo",   // Required
        Timestamp:   time.Now(),      // Required
        Timeout:     time.Second * 5, // Required

        // Votifier 1
        PublicKey: "...", // Required

        // Votifier 2
        Token: "abc123", // Required
        UUID:  "",       // Optional
    })

    if err != nil {
        panic(err)
    }
}
```

## License

[MIT License](https://github.com/mcstatus-io/mcutil/blob/main/LICENSE)