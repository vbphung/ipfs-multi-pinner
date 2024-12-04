# IPFS client supporting multiple Pinning services

## Usage

```go
package main

import (
    "bytes"
    "context"
    "fmt"
    "log"
    "strings"

    ipfsmultipinner "github.com/vbphung/ipfs-multi-pinner"
    "github.com/vbphung/ipfs-multi-pinner/core"
)

func main() {
    clt, err := ipfsmultipinner.NewClient([]*core.SelfHostedConf{
        {IpfsUrl: "http://localhost:5001", CIDVersion: 1},
    }, []core.PinService{
        // Add your own pinning services
        // which implement core.PinService interface
        // You can see an example of testPinningService at client_test.go
    })
    if err != nil {
        log.Fatalf("create client: %v\n", err)
    }

    content := "Just because you are a character doesn't mean that you have character."
    res, err := clt.Add(context.Background(), strings.NewReader(content))
    if err != nil {
        log.Fatalf("add content: %v\n", err)
    }

    r, err := clt.Get(context.Background(), res)
    if err != nil {
        log.Fatalf("get content: %v\n", err)
    }

    buf := new(bytes.Buffer)
    _, err = buf.ReadFrom(r)
    if err != nil {
        log.Fatalf("read content: %v\n", err)
    }

    fmt.Printf("content: %s\n", buf.String())
}
```
