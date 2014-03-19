go-cbshell
==========

Shell to interact with Couchbase Server

```
 $ ./cbshell -h
 Usage of ./cbshell:
   -b="default": bucket to connect
   -s="http://localhost:8091": couchbase server url to connect
```

How to play around with cbshell:
-------------------------------

```
$ go get github.com/abhi-bit/go-cbshell
$ cd $GOPATH/src/abhi-bit/go-cbshell
$ go build *.go
$ ./cbshell

cbshell> Set key 0 val
cbshell> Get key
val
cbshell> Delete key
cbshell> cstats
Host: 10.4.2.106:8091 Up: 444942 TotalMem: 6137208832 FreeMem: 5098254336 Status: healthy Version: 2.5.0-1059-rel-enterprise OS: x86_64-unknown-linux-gnu
CmdGet: 0 CurrItems: 55252 ReplicaCurrItems: 54937
Host: 10.4.2.104:8091 Up: 8324 TotalMem: 6137208832 FreeMem: 5488652288 Status: warmup Version: 2.2.0-837-rel-enterprise OS: x86_64-unknown-linux-gnu
CmdGet: 0 CurrItems: 54937 ReplicaCurrItems: 55252
Compaction Threshold: 30 RamQuota: 7363100672
cbshell> n1ql select * from orders limit 1
{
    "resultset": [
        {
            "custId": "abc",
            "id": "1200",
            "orderlines": [
                {
                    "productId": "coffee01",
                    "qty": 1
                },
                {
                    "productId": "sugar22",
                    "qty": 1
                }
            ],
            "shipped-on": "2012/01/02",
            "type": "order"
        }
    ],
    "info": [
        {
            "caller": "http_response:160",
            "code": 100,
            "key": "total_rows",
            "message": "1"
        },
        {
            "caller": "http_response:162",
            "code": 101,
            "key": "total_elapsed_time",
            "message": "818.5us"
        }
    ]
}
cbshell> help
Available commands:
  Set <key-name> <TTL> <value>
  Get <key-name>
  Delete <key-name>
  cstats
  n1ql <n1ql-query>
```
