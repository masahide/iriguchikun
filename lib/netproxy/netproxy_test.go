// Package netproxy test
package netproxy

import (
	"context"
	"errors"
	"io"
	"log"
	"net"
	"strings"
	"testing"
	"time"
)

func TestPrintErr(t *testing.T) {
	var result string
	fn := func(s ...interface{}) {
		for _, str := range s {
			result = str.(error).Error()
			return
		}
	}
	var tests = []struct {
		err  error
		fn   func(...interface{})
		want string
	}{
		{errors.New("hoge"), fn, "hoge"},
		{nil, fn, ""},
	}
	for _, test := range tests {
		result = ""
		printErr(test.fn, test.err)
		if result != test.want {
			t.Errorf("test.err(%v) = \"%v\"; want %v", test.err, result, test.want)
		}
	}
}

type testCloser struct{ err error }

func (c *testCloser) Close() error { return c.err }

type testWriter struct{ result []byte }

func (w *testWriter) Write(p []byte) (n int, err error) { w.result = p; return }

func TestCloseConn(t *testing.T) {
	tw := &testWriter{}
	log.SetOutput(tw)
	log.SetFlags(0)
	var tests = []struct {
		c    io.Closer
		want string
	}{
		{&testCloser{errors.New("hoge")}, "*netproxy.testCloser Close err:hoge"},
		{&testCloser{}, ""},
	}
	for _, test := range tests {
		tw.result = []byte{}
		closeConn(test.c)
		result := strings.TrimSpace(string(tw.result))
		if result != test.want {
			t.Errorf("closeConn(%v) = \"%v\"; want %v", test.c, result, test.want)
		}
	}
}

func TestDebugWorker(t *testing.T) {
	mes := "Waiting client connections"
	ctx, cancel := context.WithCancel(context.Background())
	clientCh := make(chan net.Conn, 1)
	pr, pw := io.Pipe()
	log.SetOutput(pw)
	log.SetFlags(0)
	go debugWorker(ctx, clientCh)
	clientCh <- &net.TCPConn{}
	b := make([]byte, len(mes))
	_, err := pr.Read(b)
	if err != nil {
		t.Fatal(err)
	}
	cancel()
	time.Sleep(1 * time.Millisecond)
	if mes != string(b) {
		t.Errorf("got: %s\nwant: %s", string(b), mes)
	}
	close(clientCh)
}

func TestMainLoop(t *testing.T) {
	tp := NetProxy{
		ListenNetwork:        "tcp",
		ListenAddr:           ":5444",
		DialNetwork:          "tcp",
		DialAddr:             "192.168.99.100:3306",
		DialTimeout:          1 * time.Second,
		PipeDeadLine:         1 * time.Second,
		RetryTime:            1 * time.Second,
		KeepAlive:            true,
		KeepAlivePeriod:      10 * time.Second,
		MaxRetry:             5,
		MaxServerConnections: 1,
		MaxClinetConnections: 1,
		DebugLevel:           1,
	}
	ctx, cancel := context.WithCancel(context.Background())
	go tp.MainLoop(ctx)
	time.Sleep(1 * time.Millisecond)
	cancel()
}
func TestDialWorker(t *testing.T) {
	tp := NetProxy{
		ListenNetwork:        "tcp",
		ListenAddr:           ":5444",
		DialNetwork:          "tcp",
		DialAddr:             "192.168.99.100:3306",
		DialTimeout:          1 * time.Second,
		PipeDeadLine:         1 * time.Second,
		RetryTime:            1 * time.Second,
		KeepAlive:            true,
		KeepAlivePeriod:      10 * time.Second,
		MaxRetry:             5,
		MaxServerConnections: 1,
		MaxClinetConnections: 1,
		DebugLevel:           1,
	}
	ctx, cancel := context.WithCancel(context.Background())
	clientCh := make(chan net.Conn, 1)
	go tp.dialWorker(ctx, clientCh)
	clientCh <- &net.TCPConn{}
	time.Sleep(1 * time.Millisecond)
	cancel()
}

func TestAcceptWorker(t *testing.T) {
	tp := NetProxy{
		ListenNetwork:        "tcp",
		ListenAddr:           ":5444",
		DialNetwork:          "tcp",
		DialAddr:             "192.168.99.100:3306",
		DialTimeout:          1 * time.Second,
		PipeDeadLine:         1 * time.Second,
		RetryTime:            1 * time.Second,
		KeepAlive:            true,
		KeepAlivePeriod:      10 * time.Second,
		MaxRetry:             5,
		MaxServerConnections: 1,
		MaxClinetConnections: 1,
		DebugLevel:           1,
	}
	ctx, cancel := context.WithCancel(context.Background())
	clientCh := make(chan net.Conn, 1)
	go tp.dialWorker(ctx, clientCh)
	clientCh <- &net.TCPConn{}
	time.Sleep(1 * time.Millisecond)
	cancel()
}
