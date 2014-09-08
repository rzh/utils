package parser

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"regexp"
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

type DataPoint struct {
	ts string
	d  string
}

type ServerStats map[string][]DataPoint

// parse output from pidstat
// return value:
//     process string  [mongod/mongos]
//     stats   map[string][]DataPoint
func ParsePIDStat(file string) (string, ServerStats) {
	var cpu []DataPoint
	var mem []DataPoint
	process := ""
	stats := make(map[string][]DataPoint)

	f, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal("cannot open file " + file)
	}

	lines := strings.Split(string(f), "\n")
	re := regexp.MustCompile("     Time      TGID")

	for i := 0; i < len(lines); i++ {
		if re.MatchString(lines[i]) {
			i++ // next line is what we are looking, which is for process. Not thread

			if i < len(lines) {
				// take Datapoint
				dps := strings.Fields(lines[i])

				if len(dps) != 19 {
					// the line is either wrong format of truncated
					log.Fatalf("Error parsing pidstat, line =(%s), wrong number of data %d",
						lines[i], len(dps))
				}

				// now take data
				cpu = append(cpu, DataPoint{ts: dps[0], d: dps[6]})
				mem = append(mem, DataPoint{ts: dps[0], d: dps[12]})

				if process == "" {
					process = dps[18]
				}
			}
		}
	}

	stats["cpu"] = cpu
	stats["mem"] = mem

	return process, stats
}
