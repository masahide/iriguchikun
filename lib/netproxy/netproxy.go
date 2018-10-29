// Package netproxy is tcp to tcp proxy
package netproxy

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

// NetProxy is main config struct
type NetProxy struct {
	ListenNetwork        string
	ListenAddr           string
	DialNetwork          string
	DialAddr             string
	DialTLSConfig        tls.Config
	DialTimeout          time.Duration
	PipeDeadLine         time.Duration
	RetryTime            time.Duration
	KeepAlivePeriod      time.Duration
	MaxRetry             int
	MaxServerConnections int
	MaxClientConnections int
	DialTLS              bool
	KeepAlive            bool
	Debug                bool
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
	clientCh := make(chan net.Conn, n.MaxClientConnections)
	for i := 0; i < n.MaxServerConnections; i++ {
		go n.dialWorker(ctx, clientCh)
	}
	if n.Debug {
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
			select {
			case <-ctx.Done():
				return
			default:
			}
			log.Printf("Failed Listener.Accept err:%s", err)
			continue
		}
		n.setKeepAliveIfTCP(conn)
		select {
		case <-ctx.Done():
			return
		case clientCh <- conn:
			continue
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
	svConn, err := n.openSvConn(ctx)
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
	n.printErrIferror(getFirstErr(ctx, errSv2Cl, errCl2Sv))
	closeConn(client)
	closeConn(svConn)
	n.readAllErr(errCl2Sv)
	n.readAllErr(errSv2Cl)
}

func getFirstErr(ctx context.Context, a, b chan error) error {
	var err error
	select {
	case err = <-a:
	case err = <-b:
	case <-ctx.Done():
	}
	return err
}

func (n *NetProxy) readAllErr(errCh chan error) {
	for range errCh {
	}
}

func (n *NetProxy) printErrIferror(err error) {
	if err != nil && err != io.EOF {
		log.Printf("pipe err:%s addr:%s", err, n.DialAddr)
	}
}

func (n *NetProxy) dial(ctx context.Context) (net.Conn, error) {
	dialer := &net.Dialer{Timeout: n.DialTimeout}
	if n.DialTLS {
		return tls.DialWithDialer(dialer, n.DialNetwork, n.DialAddr, &n.DialTLSConfig)
	}
	return dialer.DialContext(ctx, n.DialNetwork, n.DialAddr)
}

func (n *NetProxy) openSvConn(ctx context.Context) (net.Conn, error) {
	for i := 0; i < n.MaxRetry; i++ { // exponential backoff
		svConn, err := n.dial(ctx)
		if err != nil {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
			}
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
