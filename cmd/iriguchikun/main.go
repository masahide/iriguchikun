package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"time"

	"github.com/masahide/iriguchikun/lib/netproxy"
)

var (

	// Version is version number
	Version = "dev"
	// Date is build date
	Date = ""
	t    = netproxy.NetProxy{
		ListenNetwork:        "tcp",
		ListenAddr:           ":5444",
		DialNetwork:          "tcp",
		DialAddr:             "192.168.99.100:3306",
		DialTimeout:          5 * time.Second,
		PipeDeadLine:         120 * time.Second,
		RetryTime:            1 * time.Second,
		KeepAlive:            true,
		KeepAlivePeriod:      10 * time.Second,
		MaxRetry:             5,
		MaxServerConnections: 2,
		MaxClientConnections: 10,
		Debug:                false,
		DialTLS:              false,
		DialTLSConfig:        tls.Config{InsecureSkipVerify: false},
	}
	showVer bool
)

func init() {
	flag.StringVar(&t.ListenNetwork, "listenNetwork", t.ListenNetwork, "Listen network (tcp or udp or unix)")
	flag.StringVar(&t.ListenAddr, "listenAddr", t.ListenAddr, "Listen address (ipaddress or /path/to/xxx.sock)")
	flag.StringVar(&t.DialNetwork, "dialNetwork", t.DialNetwork, "Dial network (tcp or udp or unix)")
	flag.StringVar(&t.DialAddr, "dialAddr", t.DialAddr, "Dial address (ipaddress or /path/to/xxx.sock)")
	flag.DurationVar(&t.DialTimeout, "dialTimeout", t.DialTimeout, "Dial timeout")
	flag.BoolVar(&t.DialTLS, "dialTLS", t.DialTLS, "Dial tls connect")
	flag.BoolVar(&t.DialTLSConfig.InsecureSkipVerify, "tlsSkipVerify", t.DialTLSConfig.InsecureSkipVerify, "Insecure skip TLS verify")
	flag.DurationVar(&t.RetryTime, "retryTime", t.RetryTime, "Retry wait time")
	flag.IntVar(&t.MaxRetry, "maxRetry", t.MaxRetry, "Max retry")
	flag.DurationVar(&t.PipeDeadLine, "pipeDeadLine", t.PipeDeadLine, "Pipe dead line wait time")
	flag.DurationVar(&t.KeepAlivePeriod, "keepAlivePeriod", t.KeepAlivePeriod, "TCP period between keep alives")
	flag.BoolVar(&t.KeepAlive, "keepAlive", t.KeepAlive, "send keepalive messages on the connection")
	flag.IntVar(&t.MaxServerConnections, "maxServer", t.MaxServerConnections, "Max server connections")
	flag.IntVar(&t.MaxClientConnections, "maxClient", t.MaxClientConnections, "Max client connections")
	flag.BoolVar(&t.Debug, "debug", t.Debug, "debug flag")
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
