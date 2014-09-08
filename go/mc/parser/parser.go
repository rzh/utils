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

type NodeStats struct {
	Total_time_micros    int64   `json:"total_time_micros"`
	Op_throughput        float64 `json:"op_throughput"`
	Op_count             int64   `json:"op_count"`
	Op_errors            int64   `json:"op_errors"`
	Op_retries           int64   `json:"op_retries"`
	Op_retry_time_micros int64   `json:"op_retry_time_micros"`
	Op_median            int64   `json:"op_median"`
	Op_lat_avg_micros    int64   `json:"op_lat_avg_micros"`
	Op_lat_min_micros    int64   `json:"op_lat_min_micros"`
	Op_lat_max_micros    int64   `json:"op_lat_max_micros"`
	Op_lat_variance      int64   `json:"op_lat_variance"`
	Op_lat_avg_95th      int64   `json:"op_lat_avg_95th"`
	Op_lat_avg_99th      int64   `json:"op_lat_avg_99th"`
	Op_lat_total_micros  int64   `json:"op_lat_total_micros"`
}

type StatsSummary struct {
	AllNodes NodeStats              `json:"all_nodes"`
	Nodes    []map[string]NodeStats `json:"nodes"`
}

func ProcessMongoSIMResult(file string) StatsSummary {
	result := StatsSummary{}

	f, err := ioutil.ReadFile(file)

	if err != nil {
		log.Fatal("cannot open file " + file)
	}

	lines := strings.Split(string(f), "\n")
	re := regexp.MustCompile("==== final metrics ====")

	for i := len(lines) - 1; i >= 0; i-- {
		if re.MatchString(lines[i]) {
			json.Unmarshal([]byte(lines[i+1]), &result)

			return result
		}
	}

	return StatsSummary{AllNodes: NodeStats{Op_throughput: 100}}
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
			if cpu_loc == 0 {
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

				if len(dps) != 21 {
					// the line is either wrong format of truncated
					log.Fatalf("Error parsing pidstat, line =(%s), wrong number of data %d",
						lines[i], len(dps))
				}

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
