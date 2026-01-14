# tcptotcpgo

Go package: https://pkg.go.dev/github.com/manhavn/tcptotcpgo  
GitHub: https://github.com/manhavn/tcptotcpgo

`tcptotcpgo` is a simple **TCP ↔ TCP streaming connector** written in Go.

It is designed for cases where you already have **two established `net.TCPConn` connections**
and you want to **bridge them together**, forwarding bytes in both directions
until one side disconnects.

This library can be used to build:

- TCP relay apps
- TCP tunnel chains (multi-step forwarding)
- Socket bridges between a client-side stream and a server-side stream
- Debug/traffic bridging tools

---

## ✅ Install

```bash
go get github.com/manhavn/tcptotcpgo
```

## ✅ API

```go
func Connect(streamServer *net.TCPConn, streamApp *net.TCPConn, rateCheckSeconds uint8, keepAliveDelayTimeSeconds uint64)
```

### Parameters

- `streamServer`: TCP connection on the "server / listener" side
- `streamApp`: TCP connection on the "target / application" side
- `rateCheckSeconds`: how often the internal loop checks state / rate (seconds)
- `keepAliveDelayTimeSeconds`: keep-alive delay (seconds)

Example keep-alive for <b>2 hours</b>:

```go
var keepAliveDelayTimeSeconds uint64 = 7_200 // 60s * 60m * 2h = 7200s
```

---

## ✅ Full Example (Same logic as repo README)

This example bridges:

- `google.com:80` <-> `0.0.0.0:9000`

So you can open in browser:

> http://localhost:9000

> Note: Use Firefox or Zen Browser (not Chromium-based) for this test case.

```go
package main

import (
	"fmt"
	"net"
	"time"

	"github.com/manhavn/tcptotcpgo"
)

func main() {
	var rateCheckSeconds uint8 = 5
	var keepAliveDelayTimeSeconds uint64 = 7_200 // waiting 2 hours { 60s * 60p * 2h = 7200s }

	// TCP App Connect: Remote <-> Server ( google.com:80 <-> 0.0.0.0:9000 )
	go func() {
		time.Sleep(2 * time.Second)

		fmt.Println(
			"Test open URL from Firefox or Zen Browser (not Chromium, Chrome): http://localhost:9000",
		)

		streamServer, err := net.DialTCP(
			"tcp",
			&net.TCPAddr{IP: net.IPv4zero, Port: 0},
			&net.TCPAddr{IP: net.IPv4zero, Port: 9000},
		)
		if err != nil {
			fmt.Println(err)
			return
		}

		addrApp, err := net.ResolveTCPAddr("tcp", "google.com:80")
		if err != nil {
			fmt.Println(err)
			return
		}

		streamApp, err := net.DialTCP(
			"tcp",
			&net.TCPAddr{IP: net.IPv4zero, Port: 0},
			addrApp,
		)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println("Starting connect to stream server")

		tcptotcpgo.Connect(streamServer, streamApp, rateCheckSeconds, keepAliveDelayTimeSeconds)

		fmt.Println("Stop stream")
	}()

	// Create TCP Server Listener: 0.0.0.0:9000
	func() {
		streamServer, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.IPv4zero, Port: 9000})
		if err != nil {
			fmt.Println(err)
			return
		}

		var tmpStream *net.TCPConn

		for {
			accept, err := streamServer.AcceptTCP()
			if err != nil {
				break
			}

			if tmpStream != nil {
				go tcptotcpgo.Connect(accept, tmpStream, rateCheckSeconds, keepAliveDelayTimeSeconds)
			} else {
				tmpStream = accept
			}
		}
	}()
}
```

---

## Notes

- This package does <b>not</b> create listeners or dial targets automatically.
  You must create/manage TCP connections yourself using `net.DialTCP()` and `net.ListenTCP()`.
- Works best for building relay pipelines where each stage is responsible for its own socket lifecycle.
