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
	"sync"
	"testing"
	"time"
)

func TestPrintIfErr(t *testing.T) {
	tw := newTestWriter()
	log.SetOutput(tw)
	log.SetFlags(0)
	defer log.SetOutput(os.Stderr)
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
		if string(tw.getResult()) != test.want {
			t.Errorf("test.err(%v) = \"%v\"; want %v", test.err, string(tw.getResult()), test.want)
		}
		tw.setResult(tw.getResult()[:0])
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

type testWriter struct {
	result []byte
	mu     sync.Mutex
}

func newTestWriter() *testWriter { return &testWriter{[]byte{}, sync.Mutex{}} }

func (w *testWriter) Write(p []byte) (n int, err error) {
	w.setResult(append(w.getResult(), p...))
	return len(w.getResult()), nil
}
func (w *testWriter) setResult(b []byte) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.result = b
}
func (w *testWriter) getResult() []byte {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.result
}

func TestCloseConn(t *testing.T) {
	tw := newTestWriter()
	log.SetOutput(tw)
	log.SetFlags(0)
	defer log.SetOutput(os.Stderr)
	defer log.SetFlags(log.LstdFlags)
	var tests = []struct {
		c    io.Closer
		want string
	}{
		{&testCloser{errors.New("hoge")}, "*netproxy.testCloser Close err:hoge"},
		{&testCloser{}, ""},
	}
	for _, test := range tests {
		tw.setResult([]byte{})
		closeConn(test.c)
		result := strings.TrimSpace(string(tw.getResult()))
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
	log.SetFlags(0)
	defer log.SetOutput(os.Stderr)
	defer log.SetFlags(log.LstdFlags)
	go debugWorker(ctx, clientCh)
	clientCh <- &net.TCPConn{}
	time.Sleep(1200 * time.Millisecond)
	close(clientCh)
	cancel()
	if mes != string(tw.getResult()) {
		t.Errorf("got: [%s]\nwant: [%s]", string(tw.getResult()), mes)
	}
	clientCh = make(chan net.Conn, 1)
	tw.setResult(tw.getResult()[:0])
	ctx, cancel = context.WithCancel(context.Background())
	go debugWorker(ctx, clientCh)
	clientCh <- &connMock{}
	time.Sleep(1200 * time.Millisecond)
	cancel()
}

func TestMainLoop(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var np *NetProxy
	var err error
	defer log.SetOutput(os.Stderr)
	defer log.SetFlags(log.LstdFlags)
	if np, err = mockNetProxy(ctx); err != nil {
		t.Fatal(err)
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

func mockNetProxy(ctx context.Context) (*NetProxy, error) {
	tmpDir, err := ioutil.TempDir("", "iriguchikun")
	if err != nil {
		return nil, err
	}
	inSock := path.Join(tmpDir, "in.sock")
	outSock, _ := startTmpListener(ctx)
	np := &NetProxy{
		ListenNetwork:        "unix",
		ListenAddr:           inSock,
		DialNetwork:          "unix",
		DialAddr:             outSock,
		DialTimeout:          30 * time.Millisecond,
		PipeDeadLine:         30 * time.Millisecond,
		RetryTime:            30 * time.Millisecond,
		KeepAlive:            true,
		KeepAlivePeriod:      10 * time.Millisecond,
		MaxRetry:             5,
		MaxServerConnections: 1,
		MaxClinetConnections: 1,
		DebugLevel:           1,
	}
	go func() {
		<-ctx.Done()
		os.RemoveAll(tmpDir)
	}()
	return np, nil
}

func TestAcceptWorker(t *testing.T) {
	clientCh := make(chan net.Conn, 1)
	defer log.SetOutput(os.Stderr)
	defer log.SetFlags(log.LstdFlags)
	var np *NetProxy
	var err error
	var ctx context.Context
	var cancel context.CancelFunc
	log.SetFlags(0)

	var tests = []struct {
		conn       net.Conn
		outCh      bool
		wantPrefix string
		wantEmpty  bool
		err        error
	}{
		{&net.TCPConn{}, false, "Failed", false, nil},
		{&connMock{}, false, "", true, nil},
		{&connMock{}, true, "", true, nil},
		{&connMock{}, false, "Failed", false, errors.New("err")},
	}
	// test1
	for i, test := range tests {
		ctx, cancel = context.WithCancel(context.Background())
		if np, err = mockNetProxy(ctx); err != nil {
			t.Fatal(err)
		}
		tw := newTestWriter()
		log.SetOutput(tw)
		go np.acceptWorker(ctx, &listenerMock{test.conn, test.err}, clientCh)
		if test.outCh {
			<-clientCh
		}
		time.Sleep(100 * time.Millisecond)
		cancel()
		if test.wantPrefix != "" {
			if !strings.HasPrefix(string(tw.getResult()), test.wantPrefix) {
				t.Errorf("test %d got prefix: [%s]\nwant prefix: [%s]", i, string(tw.getResult()), test.wantPrefix)
			}
		}
		if test.wantEmpty {
			if string(tw.result) != "" {
				t.Errorf("test %d got [%s]\nwant : \"\"", i, string(tw.getResult()))
			}
		}
	}
}

func startTmpListener(ctx context.Context) (string, error) {
	tmpDir, err := ioutil.TempDir("", "iriguchikun")
	if err != nil {
		return "", err
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
	return sock, nil
}

func TestOpenSvConn(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var np *NetProxy
	var err error
	if np, err = mockNetProxy(ctx); err != nil {
		t.Fatal(err)
	}
	_, err = np.openSvConn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	cancel()
}

func TestDialToPipe(t *testing.T) {
	var np *NetProxy
	var err error
	var ctx context.Context
	var cancel context.CancelFunc
	var tests = []struct {
		conn       net.Conn
		dialAddr   string
		wantPrefix string
		err        error
	}{
		{&net.TCPConn{}, "", "Failed", nil},
		{&connMock{}, "0.0.0.0:0", "dial err", nil},
	}
	// test1
	log.SetFlags(0)
	defer log.SetOutput(os.Stderr)
	defer log.SetFlags(log.LstdFlags)
	for i, test := range tests {
		ctx, cancel = context.WithCancel(context.Background())
		if np, err = mockNetProxy(ctx); err != nil {
			t.Fatal(err)
		}
		tw := newTestWriter()
		log.SetOutput(tw)
		if test.dialAddr != "" {
			np.DialAddr = test.dialAddr
		}
		np.dialToPipe(ctx, test.conn)
		time.Sleep(100 * time.Millisecond)
		cancel()
		time.Sleep(100 * time.Millisecond)
		if test.wantPrefix != "" {
			if !strings.HasPrefix(string(tw.getResult()), test.wantPrefix) {
				t.Errorf("test %d got prefix: [%s]\nwant prefix: [%s]", i, string(tw.getResult()), test.wantPrefix)
			}
		}
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
	resultErr  error
}

func (l listenerMock) Accept() (net.Conn, error) { return l.resultConn, l.resultErr }

func (l listenerMock) Close() error   { return nil }
func (l listenerMock) Addr() net.Addr { return &net.UnixAddr{} }
