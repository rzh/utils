package parser

import (
	"io/ioutil"
	"log"
	"regexp"
	"strings"
)

func ProcessSysbenchResult(file string) (string, []string, map[string]string) {
	var cum string
	var trend []string
	att := make(map[string]string)

	att["test-type"] = "sysbench"
	att["thread"] = "64"

	f, err := ioutil.ReadFile(file)

	if err != nil {
		log.Fatal("cannot open file " + file)
	}

	lines := strings.Split(string(f), "\n")

	re := regexp.MustCompile("seconds : cum tps=([0-9.]+) : int tps=([0-9.]+) : cum ips=[0-9.]+ : int ips=[0-9.]+")

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
