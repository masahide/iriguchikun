package main

import (
	"errors"
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
	pipeDeadLine         = 120 * time.Second
	retryTime            = 1 * time.Second
	maxServerConnections = 2
	maxClinetConnections = 10

	// Version is version number
	Version = "dev"
	// Date is build date
	Date = ""
)

func main() {
	var v bool
	flag.StringVar(&listenNetwork, "listenNetwork", listenNetwork, "Listen network")
	flag.StringVar(&listenAddr, "listenAddr", listenAddr, "Listen address")
	flag.StringVar(&dialNetwork, "dialNetwork", dialNetwork, "Dial network")
	flag.StringVar(&dialAddr, "dialAddr", dialAddr, "Dial address")
	flag.DurationVar(&dialTimeout, "dialTimeout", dialTimeout, "Dial timeout")
	flag.DurationVar(&retryTime, "retryTime", retryTime, "Retry wait time")
	flag.DurationVar(&pipeDeadLine, "pipeDeadLine", pipeDeadLine, "Pipe dead line wait time")
	flag.IntVar(&maxServerConnections, "maxServer", maxServerConnections, "Max server connections")
	flag.IntVar(&maxClinetConnections, "maxClinet", maxClinetConnections, "Max client connections")
	flag.BoolVar(&v, "version", v, "Show version")
	flag.Parse()
	if v {
		fmt.Printf("version: %s %s\n", Version, Date)
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
	errFatal(err)
	l, err := net.ListenTCP(listenNetwork, addr)
	errFatal(err)
	defer closeConn(l)
	for {
		conn, err := l.AcceptTCP()
		if err != nil {
			log.Println(err)
			continue
		}
		errPrint(conn.SetKeepAlive(true))
		errPrint(conn.SetKeepAlivePeriod(10 * time.Second))
		clientCh <- conn
	}
}

func dial(ctx context.Context, clientCh chan *net.TCPConn) {
	for {
		select {
		case <-ctx.Done():
			return
		case client := <-clientCh:
			dialToPipe(ctx, client)
		}
	}
}

func dialToPipe(ctx context.Context, client *net.TCPConn) {
	defer closeConn(client)
	svConn, err := openSvConn()
	if err != nil {
		log.Println(err)
		return
	}
	defer closeConn(svConn)
	deadline := time.Now().Add(pipeDeadLine)
	if err := svConn.SetDeadline(deadline); err != nil {
		log.Printf("svConn.SetDeadline err:%s", err)
	}
	if err := client.SetDeadline(deadline); err != nil {
		log.Printf("client.SetDeadline err:%s", err)
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
}

func closeConn(c io.Closer) {
	if err := c.Close(); err != nil {
		log.Printf("%T Close err:%s ", c, err)
	}
}

func openSvConn() (net.Conn, error) {
	for i := 0; i < 5; i++ {
		svConn, err := net.DialTimeout(dialNetwork, dialAddr, dialTimeout)
		if err != nil {
			log.Printf("dial err:%s, addr:%s", err, dialAddr)
			time.Sleep(retryTime * time.Duration(i*i))
			continue
		}
		return svConn, nil
	}
	return nil, errors.New("dial The retry was give up")
}

func pipe(out io.Writer, in io.Reader) chan error {
	errCh := make(chan error, 1)
	go func() {
		_, err := io.Copy(out, in)
		errCh <- err
	}()
	return errCh
}
func errPrint(err error) {
	if err != nil {
		log.Println(err)
	}
}
func errFatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
