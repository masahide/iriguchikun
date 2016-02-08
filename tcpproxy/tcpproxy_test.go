// Package tcpproxy test
package tcpproxy

import (
	"errors"
	"io"
	"log"
	"strings"
	"testing"
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
		{&testCloser{errors.New("hoge")}, "*tcpproxy.testCloser Close err:hoge"},
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
