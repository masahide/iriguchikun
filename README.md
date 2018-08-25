# 入り口くん
[![Go Report Card](https://goreportcard.com/badge/github.com/masahide/iriguchikun)](https://goreportcard.com/report/github.com/masahide/iriguchikun)
[![Build Status](https://travis-ci.org/masahide/iriguchikun.svg?branch=master)](https://travis-ci.org/masahide/iriguchikun)
[![codecov](https://codecov.io/gh/masahide/iriguchikun/branch/master/graph/badge.svg)](https://codecov.io/gh/masahide/iriguchikun)
[![goreleaser](https://img.shields.io/badge/powered%20by-goreleaser-green.svg?style=flat-square)](https://github.com/goreleaser)


TCP/UDP/unix domain socketの同時接続数の調整を行うdaemonです。

- サーバーの最大接続数を超えた場合に、クライアントからの接続を順番待ちさせる
- 順番待ちのクライアントはtcp-keepaliveで接続維持


# Usage

```bash
$ ./iriguchikun  --help
Usage of ./iriguchikun:
  -debug
    	debug flag
  -dialAddr string
    	Dial address (ipaddress or /path/to/xxx.sock) (default "192.168.99.100:3306")
  -dialNetwork string
    	Dial network (tcp or udp or unix) (default "tcp")
  -dialTLS
    	Dial tls connect
  -dialTimeout duration
    	Dial timeout (default 5s)
  -keepAlive
    	send keepalive messages on the connection (default true)
  -keepAlivePeriod duration
    	TCP period between keep alives (default 10s)
  -listenAddr string
    	Listen address (ipaddress or /path/to/xxx.sock) (default ":5444")
  -listenNetwork string
    	Listen network (tcp or udp or unix) (default "tcp")
  -maxClinet int
    	Max client connections (default 10)
  -maxRetry int
    	Max retry (default 5)
  -maxServer int
    	Max server connections (default 2)
  -pipeDeadLine duration
    	Pipe dead line wait time (default 2m0s)
  -retryTime duration
    	Retry wait time (default 1s)
  -tlsSkipVerify
    	Insecure skip TLS verify
  -version
    	Show version
```


## Installation

### Linux

For RHEL/CentOS:

```bash
sudo yum install https://github.com/masahide/iriguchikun/releases/download/v1.2.0/iriguchikun_amd64.rpm
```

For Ubuntu/Debian:

```bash
wget -qO /tmp/assumer_amd64.deb https://github.com/masahide/iriguchikun/releases/download/v1.2.0/iriguchikun_amd64.deb && sudo dpkg -i /tmp/iriguchikun_amd64.deb
```

### macOS


install via [brew](https://brew.sh):

```bash
brew tap masahide/assumer https://github.com/masahide/iriguchikun
brew install iriguchikun
```


