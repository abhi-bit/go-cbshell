package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/couchbaselabs/clog"
	"github.com/couchbaselabs/go-couchbase"
)

func mf(err error) {
	if err != nil {
		clog.Error(err)
	}
}

var bucket *couchbase.Bucket
var pool couchbase.Pool

var url = flag.String("s", "http://localhost:8091", "couchbase server url to connect")
var bname = flag.String("b", "default", "bucket to connect")

func main() {
	flag.Parse()

	cb, err := couchbase.Connect(*url)
	mf(err)

	pool, err = cb.GetPool("default")
	mf(err)

	bucket, err = pool.GetBucket(*bname)
	mf(err)

	HandleInteractiveMode(*url, filepath.Base(os.Args[0]))
}

func execute_internal(url, line string, w io.Writer) error {

	cmdString := strings.Fields(line)

	err := performOp(url, cmdString)

	if err != nil {
		clog.Error(err)
	}

	return err
}

func performOp(url string, cmdString []string) error {
	cmd := cmdString[0]

	switch cmd {
	case "get":
		key := cmdString[1]
		var op string
		err := bucket.Get(key, &op)
		fmt.Println(op)
		return err
	case "set":
		key := cmdString[1]
		TTL, _ := strconv.Atoi(cmdString[2])
		value := cmdString[3]
		return bucket.Set(key, TTL, value)
	case "delete":
		key := cmdString[1]
		return bucket.Delete(key)
	case "nodes":
        for i := 0; i < len(bucket.NodesJSON); i++ {
            hostname := bucket.NodesJSON[i].Hostname
            //stats := bucket.NodesJSON[i].InterestingStats
            OS := bucket.NodesJSON[i].OS
            uptime := bucket.NodesJSON[i].Uptime
            version := bucket.NodesJSON[i].Version
            fmt.Println("H:", hostname, "OS:", OS, "UP:", uptime, "V:", version)
        }
	case "help":
		Usage()
	default:
		fmt.Println("Error: Call not supported")
		Usage()
	}
	return nil
}

func Usage() {

	fmt.Println("Available commands:\n",
		"Set <key-name> <TTL> <value>\n",
		"Get <key-name>\n",
		"Delete <key-name>\n")
}
