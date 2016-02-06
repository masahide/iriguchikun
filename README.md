# 入り口くん

TCPの同時接続数を制限するdaemonです。


# Usage 

```
$ ./iriguchikun --help
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
  -retryTime duration
    	retry wait time (default 1s)
  -version
    	Show version
```
