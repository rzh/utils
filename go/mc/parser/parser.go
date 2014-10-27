package parser

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"regexp"
	"strconv"
	"strings"
)

/*
> db.hostInfo()
{
	"system" : {
		"currentTime" : ISODate("2014-09-24T20:46:27.254Z"),
		"hostname" : "slave-5.perf.ny.cbi.10gen.cc",
		"cpuAddrSize" : 64,
		"memSizeMB" : 96734,
		"numCores" : 12,
		"cpuArch" : "x86_64",
		"numaEnabled" : true
	},
	"os" : {
		"type" : "Linux",
		"name" : "CentOS release 6.5 (Final)",
		"version" : "Kernel 2.6.32-431.el6.x86_64"
	},
	"extra" : {
		"versionString" : "Linux version 2.6.32-431.el6.x86_64 (mockbuild@c6b8.bsys.dev.centos.org) (gcc version 4.4.7 20120313 (Red Hat 4.4.7-4) (GCC) ) #1 SMP Fri Nov 22 03:15:09 UTC 2013",
		"libcVersion" : "2.12",
		"kernelVersion" : "2.6.32-431.el6.x86_64",
		"cpuFrequencyMHz" : "3466.810",
		"cpuFeatures" : "fpu vme de pse tsc msr pae mce cx8 apic sep mtrr pge mca cmov pat pse36 clflush dts acpi mmx fxsr sse sse2 ss ht tm pbe syscall nx pdpe1gb rdtscp lm constant_tsc arch_perfmon pebs bts rep_good xtopology nonstop_tsc aperfmperf pni pclmulqdq dtes64 monitor ds_cpl vmx smx est tm2 ssse3 cx16 xtpr pdcm pcid dca sse4_1 sse4_2 popcnt aes lahf_lm arat dts tpr_shadow vnmi flexpriority ept vpid",
		"pageSize" : NumberLong(4096),
		"numPages" : 24764159,
		"maxOpenFiles" : 1024
	},
	"ok" : 1
}
*/
type HostSystem struct {
	HostName    string `json:"hostname"`     // "slave-5.perf.ny.cbi.10gen.cc",
	CpuAddrSize int64  `json":"cpuAddrSize"` // 64,
	MemSizeMB   int64  `json:memSizeMB"`     // : 96734,
	NumCores    int64  `json:"numCores"`     // : 12,
	CPUArch     string `json:"cpuArch"`      // : "x86_64",
	NumaEnabled bool   `json:"numaEnabled"`  // : true
}

type HostOS struct {
	Type    string `json:"type"`    //"Linux",
	Name    string `json:"name"`    //"CentOS release 6.5 (Final)",
	Version string `json:"version"` //"Kernel 2.6.32-431.el6.x86_64"
}

type HostExra struct {
	versionString   string `json:"versionString"`   // : "Linux version 2.6.32-431.el6.x86_64 (mockbuild@c6b8.bsys.dev.centos.org) (gcc version 4.4.7 20120313 (Red Hat 4.4.7-4) (GCC) ) #1 SMP Fri Nov 22 03:15:09 UTC 2013",
	libcVersion     string `json:"libcVersion"`     // : "2.12",
	kernelVersion   string `json:"kernelVersion"`   // : "2.6.32-431.el6.x86_64",
	cpuFrequencyMHz string `json:"cpuFrequencyMHz"` // : "3466.810",
	cpuFeatures     string `json:"cpuFeatures"`     // : "fpu vme de pse tsc msr pae mce cx8 apic sep mtrr pge mca cmov pat pse36 clflush dts acpi mmx fxsr sse sse2 ss ht tm pbe syscall nx pdpe1gb rdtscp lm constant_tsc arch_perfmon pebs bts rep_good xtopology nonstop_tsc aperfmperf pni pclmulqdq dtes64 monitor ds_cpl vmx smx est tm2 ssse3 cx16 xtpr pdcm pcid dca sse4_1 sse4_2 popcnt aes lahf_lm arat dts tpr_shadow vnmi flexpriority ept vpid",
	pageSize        int64  `json:"pageSize"`        // : NumberLong(4096),
	numPages        int64  `json:"numPages"`        // : 24764159,
	maxOpenFiles    int64  `json:"maxOpenFiles"`    // : 1024
}

type MongoHostInfo struct {
	System HostSystem `json:"system"`
	Os     HostOS     `json:"os"`
	Extra  HostExra   `json:"extra"`
}

/*
	{
        "opcounters" : {
                "insert" : 30482408,
                "query" : 71843659,
                "update" : 44281266,
                "delete" : 0,
                "getmore" : 5904187,
                "command" : 1438082
        },
	...
	}
*/

type MongoServerStatus struct {
	Opcounters struct {
		Insert  int64 `json:"insert"`
		Query   int64 `json:"query"`
		Update  int64 `json:"update"`
		Delete  int64 `json:"delete"`
		GetMore int64 `json:"getmore"`
		Command int64 `json:"command"`
	} `json:"opcounters"`
	// StorageEngine StorageEngineInfo `json:"storageengine"`
}

/*
> db.serverBuildInfo()
{
	"version" : "2.6.3",
	"gitVersion" : "255f67a66f9603c59380b2a389e386910bbb52cb",
	"OpenSSLVersion" : "",
	"sysInfo" : "Linux build12.nj1.10gen.cc 2.6.32-431.3.1.el6.x86_64 #1 SMP Fri Jan 3 21:39:27 UTC 2014 x86_64 BOOST_LIB_VERSION=1_49",
	"loaderFlags" : "-fPIC -pthread -Wl,-z,now -rdynamic",
	"compilerFlags" : "-Wnon-virtual-dtor -Woverloaded-virtual -fPIC -fno-strict-aliasing -ggdb -pthread -Wall -Wsign-compare -Wno-unknown-pragmas -Winvalid-pch -pipe -Werror -O3 -Wno-unused-function -Wno-deprecated-declarations -fno-builtin-memcmp",
	"allocator" : "tcmalloc",
	"versionArray" : [
		2,
		6,
		3,
		0
	],
	"javascriptEngine" : "V8",
	"bits" : 64,
	"debug" : false,
	"maxBsonObjectSize" : 16777216,
	"ok" : 1
}
*/

type MongodBuildInfo struct {
	Version          string `json:"version"` //     "version" : "2.6.4",
	GitVersion       string `json:"gitVersion"`
	OpenSSLVersion   string `json:"OpenSSLVersion"`
	SysInfo          string `json:"sysInfo"`
	LoaderFlags      string `json:"loaderFlags"`
	CompilterFlags   string `json:"compilerFlags"`
	Allocator        string `json:"allocator"`
	JavascriptEngine string `json:"javascriptEngine"`
	Bits             int    `json:"bits"`
	Debug            bool   `json:"debug"`
}

type StorageEngineInfo struct {
	Name string `json:"name"`
}

type ServerInfo struct {
	HostInfo      MongoHostInfo     `json:"hostinfo"`
	BuildInfo     MongodBuildInfo   `json:"serverinfo"`
	StorageEngine StorageEngineInfo `json:"storageengine"`
}

type NodeStats struct {
	Total_time_micros    int64   `json:"total_time_micros,omitempty"`
	Op_throughput        float64 `json:"op_throughput,omitempty"`
	Op_count             int64   `json:"op_count,omitempty"`
	Op_errors            int64   `json:"op_errors,omitempty"`
	Op_retries           int64   `json:"op_retries,omitempty"`
	Op_retry_time_micros int64   `json:"op_retry_time_micros,omitempty"`
	Op_median            int64   `json:"op_median,omitempty"`
	Op_lat_avg_micros    int64   `json:"op_lat_avg_micros,omitempty"`
	Op_lat_min_micros    int64   `json:"op_lat_min_micros,omitempty"`
	Op_lat_max_micros    int64   `json:"op_lat_max_micros,omitempty"`
	Op_lat_variance      int64   `json:"op_lat_variance,omitempty"`
	Op_lat_avg_95th      int64   `json:"op_lat_avg_95th,omitempty"`
	Op_lat_avg_99th      int64   `json:"op_lat_avg_99th,omitempty"`
	Op_lat_total_micros  int64   `json:"op_lat_total_micros,omitempty"`
	Exhaust_avg_micros   int64   `json:"exhaust_avg_micros,omitempty"`
	Exhaust_count        int64   `json:"exhaust_count,omitempty"`
	Exhaust_errors       int64   `json:"exhaust_errors,omitempty"`
	Exhaust_max_micros   int64   `json:"exhaust_max_micros,omitempty"`
	Exhaust_min_micros   int64   `json:"exhaust_min_micros,omitempty"`
	Exhaust_retries      int64   `json:"exhaust_retries,omitempty"`
	Exhaust_retry_avg    int64   `json:"exhaust_retry_avg,omitempty"`
	Exhaust_retry_micros int64   `json:"exhaust_retry_micros,omitempty"`
	Exhaust_total_micros int64   `json:"exhaust_total_micros,omitempty"`
	Exhaust_variance     int64   `json:"exhaust_variance,omitempty"`
	Insert_avg_micros    int64   `json:"insert_avg_micros,omitempty"`
	Insert_count         int64   `json:"insert_count,omitempty"`
	Insert_errors        int64   `json:"insert_errors,omitempty"`
	Insert_max_micros    int64   `json:"insert_max_micros,omitempty"`
	Insert_min_micros    int64   `json:"insert_min_micros,omitempty"`
	Insert_retries       int64   `json:"insert_retries,omitempty"`
	Insert_retry_avg     int64   `json:"insert_retry_avg,omitempty"`
	Insert_retry_micros  int64   `json:"insert_retry_micros,omitempty"`
	Insert_total_micros  int64   `json:"insert_total_micros,omitempty"`
	Insert_variancy      int64   `json:"insert_variance,omitempty"`
	Query_avg_micros     int64   `json:"query_avg_micros,omitempty"`
	Query_count          int64   `json:"query_count,omitempty"`
	Query_errors         int64   `json:"query_errors,omitempty"`
	Query_max_micros     int64   `json:"query_max_micros,omitempty"`
	Query_min_micros     int64   `json:"query_min_micros,omitempty"`
	Query_retries        int64   `json:"query_retries,omitempty"`
	Query_retry_avg      int64   `json:"query_retry_avg,omitempty"`
	Query_retry_micros   int64   `json:"query_retry_micros,omitempty"`
	Query_total_micros   int64   `json:"query_total_micros,omitempty"`
	Query_variance       int64   `json:"query_variance,omitempty"`
	Remove_avg_micros    int64   `json:"remove_avg_micros,omitempty"`
	Remove_count         int64   `json:"remove_count,omitempty"`
	Remove_errors        int64   `json:"remove_errors,omitempty"`
	Remove_max_micros    int64   `json:"remove_max_micros,omitempty"`
	Remove_min_micros    int64   `json:"remove_min_micros,omitempty"`
	Remove_retries       int64   `json:"remove_retries,omitempty"`
	Remove_retry_avg     int64   `json:"remove_retry_avg,omitempty"`
	Remove_retry_micros  int64   `json:"remove_retry_micros,omitempty"`
	Remove_total_micros  int64   `json:"remove_total_micros,omitempty"`
	Remove_variance      int64   `json:"remove_variance,omitempty"`
	Update_avg_micros    int64   `json:"update_avg_micros,omitempty"`
	Update_count         int64   `json:"update_count,omitempty"`
	Update_errors        int64   `json:"update_errors,omitempty"`
	Update_max_micros    int64   `json:"update_max_micros,omitempty"`
	Update_min_micros    int64   `json:"update_min_micros,omitempty"`
	Update_retries       int64   `json:"update_retries,omitempty"`
	Update_retry_avg     int64   `json:"update_retry_avg,omitempty"`
	Update_retry_micros  int64   `json:"update_retry_micros,omitempty"`
	Update_total_micros  int64   `json:"update_total_micros,omitempty"`
	Update_variance      int64   `json:"update_variance,omitempty"`
}

type StatsSummary struct {
	AllNodes NodeStats              `json:"all_nodes"`
	Nodes    []map[string]NodeStats `json:"nodes"`
}

type TestbedServers struct {
	Mongod []ServerInfo `json:"mongod,omitempty"`
	Mongos []ServerInfo `json:"mongos,omitempty"`
	Config []ServerInfo `json:"config,omitempty"`
}

type Testbed struct {
	Type    string         `json:"type"`
	Servers TestbedServers `json:"servers"`
}

type TestDriver struct {
	Version   string `json:"version"`
	GitSHA    string `json:"git_hash"`
	BuildDate string `json:"build_date"`
}

type DateTime struct {
	Date int64 `json:"$date"`
}

// definition of Stat to be reported to dashboard
type Stats struct {
	Harness       string                 `json:"harness"`
	Workload      string                 `json:"workload"`
	ServerVersion string                 `json:"server_version"`
	ServerGitSHA  string                 `json:"server_git_hash"`
	StorageEngine string                 `json:"storageengine"`
	Attributes    map[string]interface{} `json:"attributes"`
	Testbed       Testbed                `json:"testbed"`
	TestDriver    TestDriver             `json:"test_driver"`
	Summary       StatsSummary           `json:"summary"`

	// TPS string
	// Run_Date   string
	Start_Time  DateTime `json:"start_time"` //epoch time, time.Now().Unix()
	End_Time    DateTime `json:"end_time"`
	Utime       int      `json:"utime"`
	Stime       int      `json:"stime"`
	TotalOps    int64    `json:"total_ops"`
	TickPerOp   float64  `json:"tick_per_op"`
	Throughput  float64  `json:"throughput"`
	TestRunTime int64    `json:"test_run_time"`
	// ID         string
	// Type         string // hammertime, sysbench, mongo-sim
	History      []string
	Server_Stats map[string]ServerStats `json:"server_stats"`
}

func ProcessMongoSIMResult(file string) *Stats {
	result := Stats{}

	f, err := ioutil.ReadFile(file)

	if err != nil {
		log.Fatal("cannot open file " + file)
	}

	lines := strings.Split(string(f), "\n")
	re := regexp.MustCompile("==== perf metrics ====")

	for i := len(lines) - 1; i >= 0; i-- {
		if re.MatchString(lines[i]) {
			json.Unmarshal([]byte(lines[i+1]), &result)

			return &result
		}
	}
	return nil
}

func ProcessSysbenchResult(file string) (string, []string, map[string]string) {
	var cum string
	var trend []string
	att := make(map[string]string)

	att["test-type"] = "sysbench"
	// att["nThread"] = "64"

	f, err := ioutil.ReadFile(file)

	if err != nil {
		log.Fatal("cannot open file " + file)
	}

	lines := strings.Split(string(f), "\n")

	// find thread in this format : writer threads           = 64

	find_parameter := func(lines []string, pattern string) string {
		var re string
		re_thread := regexp.MustCompile(pattern)

		for i := 0; i < len(lines); i++ {
			t := re_thread.FindStringSubmatch(lines[i])

			if len(t) > 0 {
				if re == "" || re == lines[i] {
					re = t[1]
				} else {
					log.Panicln("[sysbecn-parser] Found different writer threads number: ", lines[i], " vs ", re)
				}
			}
		}

		if re == "" {
			log.Panicf("Failed to find value for regexp: %s", pattern)
		}
		return re
	}

	att["nThreads"] = find_parameter(lines, "writer threads[ ]+= ([0-9]+)")
	att["nCollections"] = find_parameter(lines, "collections[ ]+= ([0-9]+)")
	att["nCollectionSize"] = find_parameter(lines, "documents per collection[ ]+= ([0-9,]+)")
	att["nFeedbackSeconds"] = find_parameter(lines, "feedback seconds[ ]+= ([0-9]+)")
	att["nRunSeconds"] = find_parameter(lines, "run seconds[ ]+= ([0-9]+)")
	att["oltp range size"] = find_parameter(lines, "oltp range size[ ]+= ([0-9]+)")
	att["oltp point selects"] = find_parameter(lines, "oltp point selects[ ]+= ([0-9]+)")
	att["oltp simple ranges"] = find_parameter(lines, "oltp simple ranges[ ]+= ([0-9]+)")
	att["oltp sum ranges"] = find_parameter(lines, "oltp sum ranges[ ]+= ([0-9]+)")
	att["oltp order ranges"] = find_parameter(lines, "oltp order ranges[ ]+= ([0-9]+)")
	att["oltp distinct ranges"] = find_parameter(lines, "oltp distinct ranges[ ]+= ([0-9]+)")
	att["oltp index updates"] = find_parameter(lines, "oltp index updates[ ]+= ([0-9]+)")
	att["oltp non index updates"] = find_parameter(lines, "oltp non index updates[ ]+= ([0-9]+)")
	att["write concern"] = find_parameter(lines, "write concern[ ]+= ([A-Z]+)")

	re := regexp.MustCompile("seconds : cum tps=([0-9.,]+) : int tps=([0-9.,]+) : cum ips=[0-9.,]+ : int ips=[0-9.,]+")

	// find the cumulative number
	for i := len(lines) - 1; i >= 0; i-- {
		t := re.FindStringSubmatch(lines[i])
		if len(t) > 0 {
			cum = t[1]
			break
		} else {
			// no match, just skip
		}
	}

	// get the historical interval number
	for i := 0; i < len(lines); i++ {
		t := re.FindStringSubmatch(lines[i])
		if len(t) > 0 {
			trend = append(trend, t[2])
		} else {
			// no match, just skip
		}
	}
	return cum, trend, att
}

func ProcessSysbenchInsertResult(file string) (string, []string, map[string]string) {
	var cum string
	var trend []string
	att := make(map[string]string)

	att["test-type"] = "sysbench-insert"

	f, err := ioutil.ReadFile(file)

	if err != nil {
		log.Fatal("cannot open file " + file)
	}

	lines := strings.Split(string(f), "\n")

	// find thread in this format : writer threads           = 64

	find_parameter := func(lines []string, pattern string) string {
		var re string
		re_thread := regexp.MustCompile(pattern)

		for i := 0; i < len(lines); i++ {
			t := re_thread.FindStringSubmatch(lines[i])

			if len(t) > 0 {
				if re == "" || re == lines[i] {
					re = t[1]
				} else {
					log.Panicln("[sysbecn-parser] Found different writer threads number: ", lines[i], " vs ", re)
				}
			}
		}

		if re == "" {
			log.Panicf("Failed to find value for regexp: %s", pattern)
		}
		return re
	}

	att["nThreads"] = find_parameter(lines, "([0-9]+) writer thread")
	att["nCollections"] = find_parameter(lines, "([0-9]+) collections")
	att["nCollectionSize"] = find_parameter(lines, "([0-9,]+) documents per collection")
	att["nFeedbackSeconds"] = find_parameter(lines, "Feedback every ([0-9]+) seconds")
	//att["nRunSeconds"] = find_parameter(lines, "run seconds[ ]+= ([0-9]+)")
	//att["oltp range size"] = find_parameter(lines, "oltp range size[ ]+= ([0-9]+)")
	//att["oltp point selects"] = find_parameter(lines, "oltp point selects[ ]+= ([0-9]+)")
	//att["oltp simple ranges"] = find_parameter(lines, "oltp simple ranges[ ]+= ([0-9]+)")
	//att["oltp sum ranges"] = find_parameter(lines, "oltp sum ranges[ ]+= ([0-9]+)")
	//att["oltp order ranges"] = find_parameter(lines, "oltp order ranges[ ]+= ([0-9]+)")
	//att["oltp distinct ranges"] = find_parameter(lines, "oltp distinct ranges[ ]+= ([0-9]+)")
	//att["oltp index updates"] = find_parameter(lines, "oltp index updates[ ]+= ([0-9]+)")
	//att["oltp non index updates"] = find_parameter(lines, "oltp non index updates[ ]+= ([0-9]+)")
	att["write concern"] = find_parameter(lines, "write concern[ ]+= ([A-Z]+)")

	// Thread[Thread-2,5,main]1,494,000 inserts : 120 seconds : cum ips=12,445.96 : int ips=5,794.21
	re := regexp.MustCompile("seconds : cum ips=([0-9.,]+) : int ips=([0-9.,]+)")

	// find the cumulative number
	for i := len(lines) - 1; i >= 0; i-- {
		t := re.FindStringSubmatch(lines[i])
		if len(t) > 0 {
			cum = t[1]
			break
		} else {
			// no match, just skip
		}
	}

	// get the historical interval number
	for i := 0; i < len(lines); i++ {
		t := re.FindStringSubmatch(lines[i])
		if len(t) > 0 {
			trend = append(trend, t[2])
		} else {
			// no match, just skip
		}
	}
	return cum, trend, att
}

type ServerStats struct {
	Process string    `json:"process "`
	Ts      []int64   `json:"ts"`
	Cpu     []float64 `json:"cpu"`
	Mem     []float64 `json:"mem"`
}

// parse output from pidstat
// return value:
//     process string  [mongod/mongos]
//     stats   map[string][]DataPoint
func ParsePIDStat(file string) ServerStats {
	var cpu []float64
	var mem []float64
	var ts []int64

	var cpu_loc, mem_loc int

	cpu_loc = -1
	mem_loc = -1

	process := ""

	f, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal("cannot open file " + file)
	}

	lines := strings.Split(string(f), "\n")
	re := regexp.MustCompile("%usr %system  %guest    %CPU")

	mline := 0
	total_mline := 0

	for i := 0; i < len(lines); i++ {
		if re.MatchString(lines[i]) {
			total_mline++
		}
	}

	for i := 0; i < len(lines); i++ {
		if re.MatchString(lines[i]) {
			if cpu_loc == -1 {
				// let's figure out location of %CPU and %MEM
				t := strings.Fields(lines[i])

				for j := 0; j < len(t); j++ {
					if t[j] == "%CPU" {
						cpu_loc = j - 1
					} else if t[j] == "%MEM" {
						mem_loc = j - 1
					}
				}
			}
			i++     // next line is what we are looking, which is for process. Not thread
			mline++ // found one data line

			// skip the first and last ten line
			if i < len(lines) && mline >= 10 && mline < total_mline-10 {
				// take Datapoint
				dps := strings.Fields(lines[i])

				// fmt.Println("cpu loc: ", cpu_loc, "  | mem loc: ", mem_loc)

				/*
					if len(dps) != 21 {
						// the line is either wrong format of truncated
						log.Fatalf("Error parsing pidstat, line =(%s), wrong number of data %d",
							lines[i], len(dps))
					}
				*/

				// now take data
				f, err := strconv.ParseFloat(dps[cpu_loc], 64)

				if err != nil {
					log.Panicln("Failed to parse CPU for pidstat with error ", err)
				}
				cpu = append(cpu, f)

				f, err = strconv.ParseFloat(dps[mem_loc], 64)
				if err != nil {
					log.Panicln("Failed to parse Mem for pidstat with error ", err)
				}
				mem = append(mem, f)

				if process == "" {
					process = dps[len(dps)-1] // the last one is process name
				}

				// ts
				t, e := strconv.ParseInt(dps[0], 10, 64)
				if e != nil {
					log.Panicln("Failed to parse Timestamp for pidstat with error ", e)
				}

				ts = append(ts, t)
			}
		}
	}

	fmt.Println("**** process  ", process)

	return ServerStats{
		Process: process,
		Cpu:     cpu,
		Mem:     mem,
		Ts:      ts,
	}
}
