package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
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
var tiServer string

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

func executeInternal(url, line string, w io.Writer) error {

	cmdString := strings.Fields(line)

	//Handle panic
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			_, ok = r.(error)
			if !ok {
				fmt.Errorf("pkg: %v", r)
			}
		}
	}()

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
	case "cstats":
		serverURL := url + "/pools/default"
		response, err := http.Get(serverURL)
		defer response.Body.Close()
		contents, _ := ioutil.ReadAll(response.Body)

		dec := json.NewDecoder(strings.NewReader(string(contents)))
		var cstats interface{}
		err = dec.Decode(&cstats)
		if err != nil {
			return err
		}

		if val, ok := cstats.(map[string]interface{}); ok {
			fmt.Printf("%+v\n", val)
		} else {
			return nil
		}

	case "n1ql":
		var query string
		tiServer = "http://localhost:8093/query"
		for i := range cmdString {
			if i != 0 {
				query = query + " " + cmdString[i]
			}
		}
		resp, err := http.Post(tiServer, "text/plain", strings.NewReader(query))
		defer resp.Body.Close()
		bs, _ := ioutil.ReadAll(resp.Body)
		fmt.Println(string(bs))
		return err
	case "help":
		Usage()
	default:
		fmt.Println("Error: Call not supported")
		Usage()
	}
	return nil
}

// Usage help function
func Usage() {

	fmt.Println("Available commands:\n",
		"set <key-name> <TTL> <value>\n",
		"get <key-name>\n",
		"delete <key-name>\n",
		"cstats\n",
		"n1ql <n1ql-query>")
}
