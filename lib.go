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
