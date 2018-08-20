// Package netproxy test
package netproxy

import (
	"context"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path"
	"strings"
	"testing"
	"time"
)

func TestPrintIfErr(t *testing.T) {
	tw := newTestWriter()
	log.SetOutput(tw)
	defer log.SetOutput(os.Stderr)
	log.SetFlags(0)
	defer log.SetFlags(log.LstdFlags)
	var tests = []struct {
		prefix string
		err    error
		want   string
	}{
		{"prefix", errors.New("hoge"), "Failed prefix err:hoge\n"},
		{"prefix", nil, ""},
		{"hoge", errors.New("fuga"), "Failed hoge err:fuga\n"},
	}
	for _, test := range tests {
		printIfErr(test.prefix, test.err)
		if string(tw.result) != test.want {
			t.Errorf("test.err(%v) = \"%v\"; want %v", test.err, string(tw.result), test.want)
		}
		tw.result = tw.result[:0]
	}
}

func TestFatalIfErr(t *testing.T) {
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
		fatalIfErr(test.fn, test.err)
		if result != test.want {
			t.Errorf("test.err(%v) = \"%v\"; want %v", test.err, result, test.want)
		}
	}
}

type testCloser struct{ err error }

func (c *testCloser) Close() error { return c.err }

type testWriter struct{ result []byte }

func newTestWriter() *testWriter { return &testWriter{[]byte{}} }

func (w *testWriter) Write(p []byte) (n int, err error) {
	w.result = append(w.result, p...)
	return len(w.result), nil
}

func TestCloseConn(t *testing.T) {
	tw := newTestWriter()
	log.SetOutput(tw)
	defer log.SetOutput(os.Stderr)
	log.SetFlags(0)
	defer log.SetFlags(log.LstdFlags)
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
	mes := "Waiting client connections: 1\n"
	ctx, cancel := context.WithCancel(context.Background())
	clientCh := make(chan net.Conn, 1)
	//pr, pw := io.Pipe()
	tw := newTestWriter()
	log.SetOutput(tw)
	defer log.SetOutput(os.Stderr)
	log.SetFlags(0)
	defer log.SetFlags(log.LstdFlags)
	go debugWorker(ctx, clientCh)
	clientCh <- &net.TCPConn{}
	time.Sleep(1200 * time.Millisecond)
	close(clientCh)
	cancel()
	if mes != string(tw.result) {
		t.Errorf("got: [%s]\nwant: [%s]", string(tw.result), mes)
	}
	clientCh = make(chan net.Conn, 1)
	tw.result = tw.result[:0]
	ctx, cancel = context.WithCancel(context.Background())
	go debugWorker(ctx, clientCh)
	clientCh <- &connMock{}
	time.Sleep(1200 * time.Millisecond)
	cancel()
}

func TestMainLoop(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "iriguchikun")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	inSock := path.Join(tmpDir, "in.sock")
	ctx, cancel := context.WithCancel(context.Background())
	log.SetOutput(os.Stderr)
	defer log.SetOutput(os.Stderr)
	log.SetFlags(log.LstdFlags)
	defer log.SetFlags(log.LstdFlags)
	outSock := startTmpListener(ctx)
	np := &NetProxy{
		ListenNetwork:        "unix",
		ListenAddr:           inSock,
		DialNetwork:          "unix",
		DialAddr:             outSock,
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
	go np.MainLoop(ctx)
	time.Sleep(100 * time.Millisecond)
	cancel()
}

func TestDialWorker(t *testing.T) {
	np := NetProxy{
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
	go np.dialWorker(ctx, clientCh)
	clientCh <- &net.TCPConn{}
	time.Sleep(100 * time.Millisecond)
	cancel()
}

func TestAcceptWorker(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "iriguchikun")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	inSock := path.Join(tmpDir, "in.sock")
	ctx, cancel := context.WithCancel(context.Background())
	outSock := startTmpListener(ctx)
	np := &NetProxy{
		ListenNetwork:        "unix",
		ListenAddr:           inSock,
		DialNetwork:          "unix",
		DialAddr:             outSock,
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
	clientCh := make(chan net.Conn, 1)
	tw := newTestWriter()
	log.SetOutput(tw)
	defer log.SetOutput(os.Stderr)
	log.SetFlags(0)
	defer log.SetFlags(log.LstdFlags)
	go np.acceptWorker(ctx, &listenerMock{&net.TCPConn{}}, clientCh)
	clientCh <- &net.TCPConn{}
	time.Sleep(100 * time.Millisecond)
	cancel()
	if !strings.HasPrefix(string(tw.result), "Failed") {
		t.Errorf("got prefix: [%s]\nwant prefix: [%s]", string(tw.result), "Failed")
	}
}

func startTmpListener(ctx context.Context) string {
	tmpDir, err := ioutil.TempDir("", "iriguchikun")
	if err != nil {
		log.Fatal(err)
	}
	sock := path.Join(tmpDir, "tmp.sock")
	l, err := net.Listen("unix", sock)
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		defer l.Close()
		defer os.RemoveAll(tmpDir)
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			conn, err := l.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				io.Copy(c, c)
				c.Close()
			}(conn)
		}
	}()
	return sock
}

func TestOpenSvConn(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "iriguchikun")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	inSock := path.Join(tmpDir, "in.sock")
	ctx, cancel := context.WithCancel(context.Background())
	log.SetOutput(os.Stderr)
	log.SetFlags(log.LstdFlags)
	outSock := startTmpListener(ctx)
	np := &NetProxy{
		ListenNetwork:        "unix",
		ListenAddr:           inSock,
		DialNetwork:          "unix",
		DialAddr:             outSock,
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
	_, err = np.openSvConn()
	if err != nil {
		t.Fatal(err)
	}
	cancel()
}

func TestDialToPipe(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "iriguchikun")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	inSock := path.Join(tmpDir, "in.sock")
	ctx, cancel := context.WithCancel(context.Background())
	outSock := startTmpListener(ctx)
	np := &NetProxy{
		ListenNetwork:        "unix",
		ListenAddr:           inSock,
		DialNetwork:          "unix",
		DialAddr:             outSock,
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
	tw := newTestWriter()
	log.SetOutput(tw)
	defer log.SetOutput(os.Stderr)
	log.SetFlags(0)
	defer log.SetFlags(log.LstdFlags)
	np.dialToPipe(ctx, &net.TCPConn{})
	cancel()
	if !strings.HasPrefix(string(tw.result), "Failed") {
		t.Errorf("got prefix: [%s]\nwant prefix: [%s]", string(tw.result), "Failed")
	}
}

type connMock struct{}

func (c *connMock) Read(b []byte) (n int, err error)   { return }
func (c *connMock) Write(b []byte) (n int, err error)  { return }
func (c *connMock) Close() error                       { return nil }
func (c *connMock) LocalAddr() net.Addr                { return &net.UnixAddr{} }
func (c *connMock) RemoteAddr() net.Addr               { return &net.UnixAddr{} }
func (c *connMock) SetDeadline(t time.Time) error      { return nil }
func (c *connMock) SetReadDeadline(t time.Time) error  { return nil }
func (c *connMock) SetWriteDeadline(t time.Time) error { return nil }

type listenerMock struct {
	resultConn net.Conn
}

func (l listenerMock) Accept() (net.Conn, error) { return l.resultConn, nil }

func (l listenerMock) Close() error   { return nil }
func (l listenerMock) Addr() net.Addr { return &net.UnixAddr{} }
