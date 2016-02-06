package main

import (
	"io"
	"log"
	"net"
	"time"

	"golang.org/x/net/context"
)

const (
	listenNetwork        = "tcp"
	listenAddr           = ":5444"
	dialNetwork          = "tcp"
	dialAddr             = "192.168.99.100:3306"
	dialTimeout          = 10 * time.Second
	maxServerConnections = 2
	maxClinetConnections = 10
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	mainLoop(ctx)
	defer cancel()

}

func mainLoop(ctx context.Context) {
	clientCh := make(chan *net.TCPConn, maxClinetConnections)
	for i := 0; i < maxServerConnections; i++ {
		go dial(ctx, clientCh)
	}
	addr, err := net.ResolveTCPAddr(listenNetwork, listenAddr)
	if err != nil {
		log.Fatal(err)
	}
	l, err := net.ListenTCP(listenNetwork, addr)
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()
	for {
		// Wait for a connection.
		conn, err := l.AcceptTCP()
		if err != nil {
			log.Fatal(err)
		}
		clientCh <- conn
	}
}

func dial(ctx context.Context, clientCh chan *net.TCPConn) {
	for {
		svConn, err := net.DialTimeout(dialNetwork, dialAddr, dialTimeout)
		if err != nil {
			log.Printf("dial err:%s, addr:%s", err, dialAddr)
			time.Sleep(dialTimeout)
			continue
		}
		select {
		case <-ctx.Done():
			return
		case client := <-clientCh:
			errch1 := pipe(client, svConn)
			errch2 := pipe(svConn, client)
			var err1, err2 error
			for i := 0; i < 2; i++ {
				select {
				case err1 = <-errch1:
				case err2 = <-errch2:
				case <-ctx.Done():
					return
				}
			}
			if err1 != io.EOF || err2 != io.EOF {
				log.Printf("pipe err1:%s, err2:%s", err1, err2)
			}
			client.Close()
			svConn.Close()
		}
	}
}

func pipe(a, b net.Conn) chan error {
	errCh := make(chan error, 1)
	go func() {
		_, err := io.Copy(a, b)
		errCh <- err
	}()
	return errCh
}
