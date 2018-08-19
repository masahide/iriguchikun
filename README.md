# 入り口くん
[![Circle CI](https://circleci.com/gh/masahide/iriguchikun.svg?style=svg)](https://circleci.com/gh/masahide/iriguchikun) [![Coverage Status](https://coveralls.io/repos/github/masahide/iriguchikun/badge.svg?branch=master)](https://coveralls.io/github/masahide/iriguchikun?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/masahide/iriguchikun)](https://goreportcard.com/report/github.com/masahide/iriguchikun)
[![Build Status](https://travis-ci.org/masahide/iriguchikun.svg?branch=master)](https://travis-ci.org/masahide/iriguchikun)
[![codecov](https://codecov.io/gh/masahide/iriguchikun/branch/master/graph/badge.svg)](https://codecov.io/gh/masahide/iriguchikun)
[![goreleaser](https://img.shields.io/badge/powered%20by-goreleaser-green.svg?style=flat-square)](https://github.com/goreleaser)


TCPの同時接続数の調整を行うdaemonです。

* サーバーの最大接続数を超えた場合に、クライアントからの接続を順番待ちさせる
* 順番待ちのクライアントはtcp-keepaliveで接続維持


# Usage 

```
$ ./iriguchikun --help
flag provided but not defined: -he
Usage of ./iriguchikun:
  -dialAddr string
    	Dial address (default "192.168.99.100:3306")
  -dialNetwork string
    	Dial network (default "tcp")
  -dialTimeout duration
    	Dial timeout (default 5s)
  -listenAddr string
    	Listen address (default ":5444")
  -listenNetwork string
    	Listen network (default "tcp")
  -maxClinet int
    	Max client connections (default 10)
  -maxServer int
    	Max server connections (default 2)
  -pipeDeadLine duration
    	Pipe dead line wait time (default 2m0s)
  -retryTime duration
    	Retry wait time (default 1s)
  -version
    	Show version
```
