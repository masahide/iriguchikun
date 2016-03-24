// Package tcpproxy is tcp to tcp proxy
package tcpproxy

import (
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"golang.org/x/net/context"
)

const (
	backoffMaxRetry = 5
)

// TCPProxy is main struct
type TCPProxy struct {
	ListenNetwork        string
	ListenAddr           string
	DialNetwork          string
	DialAddr             string
	DialTimeout          time.Duration
	PipeDeadLine         time.Duration
	RetryTime            time.Duration
	MaxServerConnections int
	MaxClinetConnections int
	DebugLevel           int
}

func debugWorker(ctx context.Context, clientCh chan *net.TCPConn) {
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
func (t *TCPProxy) MainLoop(ctx context.Context) {
	clientCh := make(chan *net.TCPConn, t.MaxClinetConnections)
	for i := 0; i < t.MaxServerConnections; i++ {
		go t.dialWorker(ctx, clientCh)
	}
	if t.DebugLevel > 0 {
		go debugWorker(ctx, clientCh)
	}
	addr, err := net.ResolveTCPAddr(t.ListenNetwork, t.ListenAddr)
	printErr(log.Fatal, err)
	l, err := net.ListenTCP(t.ListenNetwork, addr)
	printErr(log.Fatal, err)
	defer closeConn(l)
	go t.acceptWorker(ctx, l, clientCh)
	<-ctx.Done()
}

func (t *TCPProxy) dialWorker(ctx context.Context, clientCh chan *net.TCPConn) {
	for {
		select {
		case <-ctx.Done():
			return
		case client := <-clientCh:
			t.dialToPipe(ctx, client)
		}
	}
}

func (t *TCPProxy) acceptWorker(ctx context.Context, l *net.TCPListener, clientCh chan *net.TCPConn) {
	for {
		conn, err := l.AcceptTCP()
		if err != nil {
			log.Println(err)
			continue
		}
		printErr(log.Println, conn.SetKeepAlive(true))
		printErr(log.Println, conn.SetKeepAlivePeriod(10*time.Second))
		clientCh <- conn
	}
}

func (t *TCPProxy) dialToPipe(ctx context.Context, client *net.TCPConn) {
	svConn, err := t.openSvConn()
	if err != nil {
		log.Println(err)
		closeConn(client)
		return
	}
	deadline := time.Now().Add(t.PipeDeadLine)
	printErr(log.Println, svConn.SetDeadline(deadline))
	printErr(log.Println, client.SetDeadline(deadline))
	errSv2Cl := pipe(client, svConn)
	errCl2Sv := pipe(svConn, client)
	select {
	case err = <-errCl2Sv:
	case err = <-errSv2Cl:
	case <-ctx.Done():
	}
	if err != nil && err != io.EOF {
		log.Printf("pipe err:%s addr:%s", err, t.DialAddr)
	}
	closeConn(client)
	closeConn(svConn)

	//残ったメッセージを読み捨てる
	for range errCl2Sv {
	}
	for range errSv2Cl {
	}
}

func (t *TCPProxy) openSvConn() (net.Conn, error) {
	for i := 0; i < backoffMaxRetry; i++ { // exponential backoff
		svConn, err := net.DialTimeout(t.DialNetwork, t.DialAddr, t.DialTimeout)
		if err != nil {
			log.Printf("dial err:%s, addr:%s", err, t.DialAddr)
			time.Sleep(t.RetryTime * time.Duration(i*i))
			continue
		}
		return svConn, nil
	}
	return nil, fmt.Errorf("dial The retry was give up. addr:%s", t.DialAddr)
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

func printErr(printFunc func(...interface{}), err error) {
	if err != nil {
		printFunc(err)
	}
}
