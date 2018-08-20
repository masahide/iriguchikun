// Package netproxy is tcp to tcp proxy
package netproxy

import (
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"context"
)

// NetProxy is main struct
type NetProxy struct {
	ListenNetwork        string
	ListenAddr           string
	DialNetwork          string
	DialAddr             string
	DialTimeout          time.Duration
	PipeDeadLine         time.Duration
	RetryTime            time.Duration
	KeepAlive            bool
	KeepAlivePeriod      time.Duration
	MaxRetry             int
	MaxServerConnections int
	MaxClinetConnections int
	DebugLevel           int
}

func debugWorker(ctx context.Context, clientCh chan net.Conn) {
	ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-ticker.C:
			log.Printf("Waiting client connections: %d", len(clientCh))
		case <-ctx.Done():
			return
		}
	}
}

// MainLoop ctxでキャンセルされるまでloop
func (n *NetProxy) MainLoop(ctx context.Context) {
	clientCh := make(chan net.Conn, n.MaxClinetConnections)
	for i := 0; i < n.MaxServerConnections; i++ {
		go n.dialWorker(ctx, clientCh)
	}
	if n.DebugLevel > 0 {
		go debugWorker(ctx, clientCh)
	}
	l, err := net.Listen(n.ListenNetwork, n.ListenAddr)
	fatalIfErr(log.Fatal, err)
	defer closeConn(l)
	go n.acceptWorker(ctx, l, clientCh)
	<-ctx.Done()
}

func (n *NetProxy) dialWorker(ctx context.Context, clientCh chan net.Conn) {
	for {
		select {
		case <-ctx.Done():
			return
		case client := <-clientCh:
			n.dialToPipe(ctx, client)
		}
	}
}

func (n *NetProxy) acceptWorker(ctx context.Context, l net.Listener, clientCh chan net.Conn) {
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Printf("Failed Listener.Accept err:%s", err)
			continue
		}
		n.setKeepAliveIfTCP(conn)
		select {
		case <-ctx.Done():
			return
		case clientCh <- conn:
		}
	}
}
func (n *NetProxy) setKeepAliveIfTCP(conn net.Conn) {
	switch v := conn.(type) {
	case *net.TCPConn:
		printIfErr("conn.SetKeepAlive", v.SetKeepAlive(n.KeepAlive))
		printIfErr("conn.SetKeepAlivePeriod", v.SetKeepAlivePeriod(n.KeepAlivePeriod))
	}
}

func (n *NetProxy) dialToPipe(ctx context.Context, client net.Conn) {
	svConn, err := n.openSvConn()
	if err != nil {
		log.Println(err)
		closeConn(client)
		return
	}
	deadline := time.Now().Add(n.PipeDeadLine)
	printIfErr("svConn.SetDeadline", svConn.SetDeadline(deadline))
	printIfErr("client.SetDeadline", client.SetDeadline(deadline))
	errSv2Cl := pipe(client, svConn)
	errCl2Sv := pipe(svConn, client)
	select {
	case err = <-errCl2Sv:
	case err = <-errSv2Cl:
	case <-ctx.Done():
	}
	n.printErrIferror(err)
	closeConn(client)
	closeConn(svConn)

	for err = range errCl2Sv {
		n.printErrIferror(err)
	}
	for err = range errSv2Cl {
		n.printErrIferror(err)
	}
}

func (n *NetProxy) printErrIferror(err error) {
	if err != nil && err != io.EOF {
		log.Printf("pipe err:%s addr:%s", err, n.DialAddr)
	}
}

func (n *NetProxy) openSvConn() (net.Conn, error) {
	for i := 0; i < n.MaxRetry; i++ { // exponential backoff
		svConn, err := net.DialTimeout(n.DialNetwork, n.DialAddr, n.DialTimeout)
		if err != nil {
			log.Printf("dial err:%s, addr:%s", err, n.DialAddr)
			time.Sleep(n.RetryTime * time.Duration(i*i))
			continue
		}
		return svConn, nil
	}
	return nil, fmt.Errorf("dial The retry was give up. addr:%s", n.DialAddr)
}

func pipe(out io.Writer, in io.Reader) chan error {
	errCh := make(chan error, 1)
	go func() {
		_, err := io.Copy(out, in)
		errCh <- err
		close(errCh)
	}()
	return errCh
}

func closeConn(c io.Closer) {
	if err := c.Close(); err != nil {
		log.Printf("%T Close err:%s ", c, err)
	}
}

func printIfErr(prefix string, err error) {
	if err != nil {
		log.Printf("Failed %s err:%s", prefix, err)
	}
}

func fatalIfErr(printFunc func(...interface{}), err error) {
	if err != nil {
		printFunc(err)
	}
}
