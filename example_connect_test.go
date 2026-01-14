// example_connect_test.go
package tcptotcpgo_test

import (
	"bufio"
	"fmt"
	"net"
	"time"

	"github.com/manhavn/tcptotcpgo"
)

func ExampleConnect() {
	var rateCheckSeconds uint8 = 1
	var keepAliveDelayTimeSeconds uint64 = 60 // keep-alive delay (seconds), demo value

	// ---------------------------------------------------------------------
	// 1) Start a target TCP echo server (the "app/target" side).
	// ---------------------------------------------------------------------
	targetLn, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	if err != nil {
		fmt.Println("listen target:", err)
		return
	}
	defer targetLn.Close()

	targetAddr := targetLn.Addr().String()

	go func() {
		conn, err := targetLn.AcceptTCP()
		if err != nil {
			return
		}
		defer conn.Close()

		// Simple line-based echo.
		_ = conn.SetDeadline(time.Now().Add(5 * time.Second))
		r := bufio.NewReader(conn)
		for {
			line, rerr := r.ReadString('\n')
			if rerr != nil {
				return
			}
			_, _ = conn.Write([]byte(line))
		}
	}()

	// ---------------------------------------------------------------------
	// 2) Start a forwarder listener (the "server/listener" side).
	//    For each client connection, dial the target and bridge with Connect.
	// ---------------------------------------------------------------------
	forwardLn, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	if err != nil {
		fmt.Println("listen forward:", err)
		return
	}
	defer forwardLn.Close()

	forwardAddr := forwardLn.Addr().String()

	go func() {
		clientConn, err := forwardLn.AcceptTCP()
		if err != nil {
			return
		}

		targetTCPAddr, err := net.ResolveTCPAddr("tcp", targetAddr)
		if err != nil {
			_ = clientConn.Close()
			return
		}

		appConn, err := net.DialTCP("tcp", nil, targetTCPAddr)
		if err != nil {
			_ = clientConn.Close()
			return
		}

		// Bridge bytes both directions until one side closes.
		tcptotcpgo.Connect(clientConn, appConn, rateCheckSeconds, keepAliveDelayTimeSeconds)
	}()

	// ---------------------------------------------------------------------
	// 3) Client connects to the forwarder, sends a line, receives echoed line.
	// ---------------------------------------------------------------------
	fwdTCPAddr, err := net.ResolveTCPAddr("tcp", forwardAddr)
	if err != nil {
		fmt.Println("resolve forward:", err)
		return
	}

	c, err := net.DialTCP("tcp", nil, fwdTCPAddr)
	if err != nil {
		fmt.Println("dial forward:", err)
		return
	}
	defer c.Close()

	_ = c.SetDeadline(time.Now().Add(5 * time.Second))

	_, _ = c.Write([]byte("ping\n"))

	reply, err := bufio.NewReader(c).ReadString('\n')
	if err != nil {
		fmt.Println("read reply:", err)
		return
	}

	fmt.Print(reply)

	// Output:
	// ping
}
