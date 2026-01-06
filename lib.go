package tcptotcpgo

import (
	"net"
	"time"
)

func stream(
	readerClosed *bool,
	writerClosed *bool,
	streamRead *net.TCPConn,
	streamWrite *net.TCPConn,
	keepAlivePingRead *bool,
) {
	defer func() {
		if streamWrite != nil {
			_ = streamWrite.Close()
		}
		*writerClosed = true
	}()
	buf := make([]byte, 16*1024)
	for {
		if *readerClosed {
			break
		} else {
			n, err := streamRead.Read(buf)
			if err != nil || n == 0 {
				break
			}
			_, err = streamWrite.Write(buf[:n])
			if err != nil {
				break
			}
			*keepAlivePingRead = true
		}
	}
}

func Connect(streamServer *net.TCPConn, streamApp *net.TCPConn, timeoutAliveSec int64) {
	streamServerClosed := false
	streamAppClosed := false
	keepAlivePingServer := false
	keepAlivePingApp := false

	go func() {
		var timeout int64 = 5 // timeout rate 5s
		countMaxWaiting := timeoutAliveSec / timeout
		time.Sleep(time.Duration(timeout) * time.Second)
		var countWaiting int64
		for {
			if keepAlivePingApp && keepAlivePingServer {
				keepAlivePingApp = false
				keepAlivePingServer = false
				countWaiting = 0
			} else {
				countWaiting++
				// waiting 2 hours
				if countWaiting >= countMaxWaiting {
					streamServerClosed = true
					streamAppClosed = true
					if streamServer != nil {
						streamServer.Close()
					}
					if streamApp != nil {
						streamApp.Close()
					}
					break
				}
				time.Sleep(time.Duration(timeout) * time.Second)
				if streamServerClosed || streamAppClosed {
					break
				}
			}
		}
	}()

	go stream(&streamServerClosed, &streamAppClosed, streamServer, streamApp, &keepAlivePingServer)
	stream(&streamAppClosed, &streamServerClosed, streamApp, streamServer, &keepAlivePingApp)
}
