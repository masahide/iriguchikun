package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/masahide/iriguchikun/tcpproxy"
	"golang.org/x/net/context"
)

var (

	// Version is version number
	Version = "dev"
	// Date is build date
	Date = ""
	t    = tcpproxy.TCPProxy{
		ListenNetwork:        "tcp",
		ListenAddr:           ":5444",
		DialNetwork:          "tcp",
		DialAddr:             "192.168.99.100:3306",
		DialTimeout:          5 * time.Second,
		PipeDeadLine:         120 * time.Second,
		RetryTime:            1 * time.Second,
		MaxServerConnections: 2,
		MaxClinetConnections: 10,
	}
	showVer bool
)

func init() {
	flag.StringVar(&t.ListenNetwork, "listenNetwork", t.ListenNetwork, "Listen network")
	flag.StringVar(&t.ListenAddr, "listenAddr", t.ListenAddr, "Listen address")
	flag.StringVar(&t.DialNetwork, "dialNetwork", t.DialNetwork, "Dial network")
	flag.StringVar(&t.DialAddr, "dialAddr", t.DialAddr, "Dial address")
	flag.DurationVar(&t.DialTimeout, "dialTimeout", t.DialTimeout, "Dial timeout")
	flag.DurationVar(&t.RetryTime, "retryTime", t.RetryTime, "Retry wait time")
	flag.DurationVar(&t.PipeDeadLine, "pipeDeadLine", t.PipeDeadLine, "Pipe dead line wait time")
	flag.IntVar(&t.MaxServerConnections, "maxServer", t.MaxServerConnections, "Max server connections")
	flag.IntVar(&t.MaxClinetConnections, "maxClinet", t.MaxClinetConnections, "Max client connections")
	flag.BoolVar(&showVer, "version", showVer, "Show version")
	flag.Parse()
}

func main() {
	if showVer {
		fmt.Printf("version: %s %s\n", Version, Date)
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	t.MainLoop(ctx)
}
