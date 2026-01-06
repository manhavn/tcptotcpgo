# SETUP

- Github: [https://github.com/manhavn/tcptotcpgo](https://github.com/manhavn/tcptotcpgo)
- Go Package: [https://pkg.go.dev/github.com/manhavn/tcptotcpgo](https://pkg.go.dev/github.com/manhavn/tcptotcpgo)

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
		streamApp, err := net.DialTCP("tcp", &net.TCPAddr{IP: net.IPv4zero, Port: 0}, addrApp)
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
