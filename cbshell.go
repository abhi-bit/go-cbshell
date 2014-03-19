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

func execute_internal(url, line string, w io.Writer) error {

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

		var cstats ClusterStats
		json.Unmarshal(contents, &cstats)
		for _, node := range cstats.Nodes {
			fmt.Println("Host:", node.Hostname,
				"Up:", node.Uptime,
				"TotalMem:", node.MemoryTotal,
				"FreeMem:", node.MemoryFree,
				"Status:", node.Status,
				"Version:", node.Version,
				"OS:", node.Os)
			iStats := node.InterestingStats
			fmt.Println("CmdGet:", iStats.CmdGet,
				"CurrItems:", iStats.CurrItems,
				"ReplicaCurrItems:", iStats.VbReplicaCurrItems)
		}
		fmt.Println("Compaction Threshold:", cstats.AutoCompactionSettings.ViewFragmentationThreshold.Percentage,
			"RamQuota:", cstats.StorageTotals.Ram.QuotaTotal)
		return err
	case "n1ql":
		var query string
		tiServer = "http://localhost:8093/query"
		for i, _ := range cmdString {
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

type ClusterStats struct {
	StorageTotals struct {
		Ram struct {
			Total      int64 `json:"total"`
			QuotaTotal int   `json:"quotaTotal"`
			QuotaUsed  int   `json:"quotaUsed"`
			Used       int64 `json:"used"`
			UsedByData int   `json:"usedByData"`
		} `json:"ram"`
		Hdd struct {
			Total      int64 `json:"total"`
			QuotaTotal int64 `json:"quotaTotal"`
			Used       int64 `json:"used"`
			UsedByData int   `json:"usedByData"`
			Free       int64 `json:"free"`
		} `json:"hdd"`
	} `json:"storageTotals"`
	ServerGroupsUri  string        `json:"serverGroupsUri"`
	Name             string        `json:"name"`
	Alerts           []interface{} `json:"alerts"`
	AlertsSilenceURL string        `json:"alertsSilenceURL"`
	Nodes            []struct {
		SystemStats struct {
			CpuUtilizationRate float32 `json:"cpu_utilization_rate"`
			SwapTotal          int     `json:"swap_total"`
			SwapUsed           int     `json:"swap_used"`
			MemTotal           int64   `json:"mem_total"`
			MemFree            int64   `json:"mem_free"`
		} `json:"systemStats"`
		InterestingStats struct {
			CmdGet                   int `json:"cmd_get"`
			CouchDocsActualDiskSize  int `json:"couch_docs_actual_disk_size"`
			CouchDocsDataSize        int `json:"couch_docs_data_size"`
			CouchViewsActualDiskSize int `json:"couch_views_actual_disk_size"`
			CouchViewsDataSize       int `json:"couch_views_data_size"`
			CurrItems                int `json:"curr_items"`
			CurrItemsTot             int `json:"curr_items_tot"`
			EpBgFetched              int `json:"ep_bg_fetched"`
			GetHits                  int `json:"get_hits"`
			MemUsed                  int `json:"mem_used"`
			Ops                      int `json:"ops"`
			VbReplicaCurrItems       int `json:"vb_replica_curr_items"`
		} `json:"interestingStats"`
		Uptime               string `json:"uptime"`
		MemoryTotal          int64  `json:"memoryTotal"`
		MemoryFree           int64  `json:"memoryFree"`
		McdMemoryReserved    int    `json:"mcdMemoryReserved"`
		McdMemoryAllocated   int    `json:"mcdMemoryAllocated"`
		CouchApiBase         string `json:"couchApiBase"`
		ClusterMembership    string `json:"clusterMembership"`
		Status               string `json:"status"`
		OtpNode              string `json:"otpNode"`
		ThisNode             bool   `json:"thisNode"`
		Hostname             string `json:"hostname"`
		ClusterCompatibility int    `json:"clusterCompatibility"`
		Version              string `json:"version"`
		Os                   string `json:"os"`
		Ports                struct {
			HttpsMgmt int `json:"httpsMgmt"`
			HttpsCAPI int `json:"httpsCAPI"`
			SslProxy  int `json:"sslProxy"`
			Proxy     int `json:"proxy"`
			Direct    int `json:"direct"`
		} `json:"ports"`
	} `json:"nodes"`
	Buckets struct {
		Uri                       string `json:"uri"`
		TerseBucketsBase          string `json:"terseBucketsBase"`
		TerseStreamingBucketsBase string `json:"terseStreamingBucketsBase"`
	} `json:"buckets"`
	RemoteClusters struct {
		Uri         string `json:"uri"`
		ValidateURI string `json:"validateURI"`
	} `json:"remoteClusters"`
	Controllers struct {
		AddNode struct {
			Uri string `json:"uri"`
		} `json:"addNode"`
		Rebalance struct {
			Uri string `json:"uri"`
		} `json:"rebalance"`
		FailOver struct {
			Uri string `json:"uri"`
		} `json:"failOver"`
		ReAddNode struct {
			Uri string `json:"uri"`
		} `json:"reAddNode"`
		EjectNode struct {
			Uri string `json:"uri"`
		} `json:"ejectNode"`
		SetAutoCompaction struct {
			Uri         string `json:"uri"`
			ValidateURI string `json:"validateURI"`
		} `json:"setAutoCompaction"`
		Replication struct {
			CreateURI   string `json:"createURI"`
			ValidateURI string `json:"validateURI"`
		} `json:"replication"`
		SetFastWarmup struct {
			Uri         string `json:"uri"`
			ValidateURI string `json:"validateURI"`
		} `json:"setFastWarmup"`
	} `json:"controllers"`
	RebalanceStatus        string `json:"rebalanceStatus"`
	RebalanceProgressUri   string `json:"rebalanceProgressUri"`
	StopRebalanceUri       string `json:"stopRebalanceUri"`
	NodeStatusesUri        string `json:"nodeStatusesUri"`
	MaxBucketCount         int    `json:"maxBucketCount"`
	AutoCompactionSettings struct {
		ParallelDBAndViewCompaction    bool `json:"parallelDBAndViewCompaction"`
		DatabaseFragmentationThreshold struct {
			Percentage int    `json:"percentage"`
			Size       string `json:"size"`
		} `json:"databaseFragmentationThreshold"`
		ViewFragmentationThreshold struct {
			Percentage int    `json:"percentage"`
			Size       string `json:"size"`
		} `json:"viewFragmentationThreshold"`
	} `json:"autoCompactionSettings"`
	FastWarmupSettings struct {
		FastWarmupEnabled  bool `json:"fastWarmupEnabled"`
		MinMemoryThreshold int  `json:"minMemoryThreshold"`
		MinItemsThreshold  int  `json:"minItemsThreshold"`
	} `json:"fastWarmupSettings"`
	Tasks struct {
		Uri string `json:"uri"`
	} `json:"tasks"`
	Counters struct {
	} `json:"counters"`
}

func Usage() {

	fmt.Println("Available commands:\n",
		"set <key-name> <TTL> <value>\n",
		"get <key-name>\n",
		"delete <key-name>\n",
		"cstats",
		"n1ql <n1ql-query>")
}
