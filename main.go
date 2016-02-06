package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"golang.org/x/net/context"
)

var (
	listenNetwork        = "tcp"
	listenAddr           = ":5444"
	dialNetwork          = "tcp"
	dialAddr             = "192.168.99.100:3306"
	dialTimeout          = 5 * time.Second
	retryTime            = 1 * time.Second
	maxServerConnections = 2
	maxClinetConnections = 10
	Version              = "dev"
	BuildDate            = ""
)

func main() {
	var v bool
	flag.StringVar(&listenNetwork, "listenNetwork", listenNetwork, "Listen network")
	flag.StringVar(&listenAddr, "listenAddr", listenAddr, "Listen address")
	flag.StringVar(&dialNetwork, "dialNetwork", dialNetwork, "Dial network")
	flag.StringVar(&dialAddr, "dialAddr", dialAddr, "Dial address")
	flag.DurationVar(&dialTimeout, "dialTimeout", dialTimeout, "Dial timeout")
	flag.DurationVar(&retryTime, "retryTime", retryTime, "retry wait time")
	flag.IntVar(&maxServerConnections, "maxServer", maxServerConnections, "Max server connections")
	flag.IntVar(&maxClinetConnections, "maxClinet", maxClinetConnections, "Max client connections")
	flag.BoolVar(&v, "version", v, "Show version")
	flag.Parse()
	if v {
		fmt.Printf("version: %s %s\n", Version, BuildDate)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mainLoop(ctx)

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
		conn.SetKeepAlive(true)
		conn.SetKeepAlivePeriod(10 * time.Second)
		clientCh <- conn
	}
}

func dial(ctx context.Context, clientCh chan *net.TCPConn) {
	for {
		//log.Printf("svConn:%v", svConn)
		select {
		case <-ctx.Done():
			return
		case client := <-clientCh:
			var err error
			var svConn net.Conn
			for i := 0; i < 5; i++ {
				svConn, err = net.DialTimeout(dialNetwork, dialAddr, dialTimeout)
				if err != nil {
					log.Printf("dial err:%s, addr:%s", err, dialAddr)
					time.Sleep(retryTime * time.Duration(i*i))
					continue
				}
				break
			}
			errch1 := pipe(client, svConn)
			errch2 := pipe(svConn, client)
			select {
			case err = <-errch1:
			case err = <-errch2:
			case <-ctx.Done():
				return
			}
			if err != nil && err != io.EOF {
				log.Printf("pipe err:%s", err)
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
