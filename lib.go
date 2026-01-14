// Package tcptotcpgo provides a minimal TCP-to-TCP stream bridge.
//
// The library is intended for scenarios where you already have two established
// TCP connections and want to connect them together, forwarding bytes in both
// directions until one side closes or an error occurs.
//
// Typical use-cases:
//
//   - TCP relay / port-forwarding building blocks
//   - Multi-hop forwarding pipelines (TCP -> TCP -> ...)
//   - Bridging a client-side stream to an upstream application stream
//
// This package does not create listeners or dial remote targets for you.
// You are responsible for creating and managing *net.TCPConn objects
// (e.g. using net.ListenTCP / net.DialTCP) and passing them to Connect.
package tcptotcpgo

import (
	"net"
	"time"
)

func stream(closed *bool, reader *net.TCPConn, writer *net.TCPConn, ping *bool) {
	defer func() {
		*closed = true
		if reader != nil {
			_ = reader.Close()
		}
		if writer != nil {
			_ = writer.Close()
		}
	}()
	buf := make([]byte, 16*1024)
	for {
		if *closed {
			break
		} else {
			n, err := reader.Read(buf)
			if err != nil || n == 0 {
				break
			}
			*ping = true
			_, err = writer.Write(buf[:n])
			if err != nil {
				break
			}
		}
	}
}

// Connect bridges two established TCP connections and streams data both ways.
//
// Connect continuously copies data between streamServer and streamApp in both
// directions (server -> app, and app -> server). It returns only after the
// bridge ends due to connection close or an I/O error.
//
// Parameters:
//
//   - streamServer is the TCP connection on the "server/listener" side
//     (usually the accepted connection from net.ListenTCP).
//
//   - streamApp is the TCP connection on the "app/target" side
//     (usually the outbound connection created by net.DialTCP).
//
//   - rateCheckSeconds defines how often the internal loop checks status/rate.
//     This value is in seconds. A small value (e.g. 1–5) makes the bridge react
//     faster; a larger value reduces CPU wakeups.
//
//   - keepAliveDelayTimeSeconds configures TCP keep-alive behavior (seconds).
//     When enabled by the implementation, keep-alive helps detect dead peers
//     on long-lived idle connections.
//
// Behavior notes:
//
//   - Connect assumes both connections are valid (non-nil) and already connected.
//     If either connection is nil or not connected, the bridge will end with error.
//
//   - Ownership/Lifecycle: Connect may read from and write to both connections.
//     The caller should not concurrently use the same *net.TCPConn while Connect
//     is running. Closing either connection from another goroutine will typically
//     cause Connect to exit.
//
//   - Concurrency: The implementation commonly uses two copy loops/goroutines
//     (one for each direction). Connect blocks the caller until the bridge ends.
//
// Example (conceptual):
//
//	ln, _ := net.ListenTCP("tcp", &net.TCPAddr{IP: net.IPv4zero, Port: 9000})
//	serverConn, _ := ln.AcceptTCP()
//
//	appAddr, _ := net.ResolveTCPAddr("tcp", "example.com:80")
//	appConn, _ := net.DialTCP("tcp", nil, appAddr)
//
//	tcptotcpgo.Connect(serverConn, appConn, 5, 7200)
func Connect(
	stream1 *net.TCPConn,
	stream2 *net.TCPConn,
	rateCheckSeconds uint8,
	keepAliveDelayTimeSeconds uint64,
) {
	var closed, ping1, ping2 bool

	go stream(&closed, stream1, stream2, &ping1)
	go stream(&closed, stream2, stream1, &ping2)

	if rateCheckSeconds < 1 {
		rateCheckSeconds = 1
	}
	if keepAliveDelayTimeSeconds < 2 {
		keepAliveDelayTimeSeconds = 2
	}

	var delay uint64
	maxDelay := keepAliveDelayTimeSeconds / uint64(rateCheckSeconds)
	rateCheck := time.Duration(rateCheckSeconds) * time.Second
	for {
		if ping1 && ping2 {
			ping1 = false
			ping2 = false
			delay = 0 // reset delay count
		} else {
			if delay > maxDelay {
				closed = true
				if stream1 != nil {
					_ = stream1.Close()
				}
				if stream2 != nil {
					_ = stream2.Close()
				}
				break
			}
			delay++
			time.Sleep(rateCheck)
			if closed {
				break
			}
		}
	}
}
