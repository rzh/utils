package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/rzh/utils/go/mc/parser"
)

/*
type Testbed struct {
	Type    string                       `json:type`
	Servers map[string]map[string]string `json:"servers"`
}

type TestDriver struct {
	Version   string `json:version`
	GitSHA    string `json:git_hash`
	BuildDate string `json:build_date`
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
	Attributes    map[string]interface{} `json:"attributes"`
	Testbed       Testbed                `json:"test_bed"`
	TestDriver    TestDriver             `json:"test_driver"`
	Summary       parser.StatsSummary    `json:"summary"`

	// TPS string
	// Run_Date   string
	Start_Time DateTime `json:"start_time"` //epoch time, time.Now().Unix()
	End_Time   DateTime `json:"end_time"`
	// ID         string
	// Type         string // hammertime, sysbench, mongo-sim
	History      []string
	Server_Stats map[string]parser.ServerStats `json:"server_stats"`
}
*/

func replaceDot(s string) string {
	return strings.Replace(s, ".", "_", -1)
}

// only used for clean report
type HammerTask_ struct {
	Run           string   `json:"run"`
	Run_id        string   `json:"run_id"`
	Cmd           string   `json:"cmd"`
	Hammer_folder string   `json:"hammer_folder"`
	Clients       []string `json:"clients"`
	Servers       []string `json:"servers"`

	// log files to be collected from client and server
	Client_logs []string `json:"client_logs"`
	Server_logs []string `json:"server_logs"`

	Type string `json:"type"`
}

func (r *TheRun) reportResults(run_id int, log_file string, run_dir string) {
	// this is the place to analyze results.
	t := strings.ToLower(r.Runs[run_id].Type)
	// r.Runs[run_id].Stats.Type = t
	r.Runs[run_id].Stats.Harness = t
	// r.Runs[run_id].Stats.ID = r.Runs[run_id].Run_id
	r.Runs[run_id].Stats.Workload = r.Runs[run_id].Run_id

	// cache run first
	//rr, _ := json.Marshal(r.Runs[run_id])
	rr := r.Runs[run_id]
	r.Runs[run_id].Stats.Attributes["run-by"] = "hammer-mc"
	r.Runs[run_id].Stats.Attributes["hammer-mc-cmd"] = HammerTask_{Run_id: rr.Run_id,
		Cmd: rr.Cmd, Clients: rr.Clients, Servers: rr.Servers,
		Client_logs: rr.Client_logs, Server_logs: rr.Server_logs,
		Type: rr.Type, Run: run}

	if len(report_url) == 0 {
		report_url = append(report_url, "http://54.68.84.192:8080/api/v1/results")
		report_url = append(report_url, "http://dyno.mongodb.parts/api/v1/results")
	}
	var err error
	switch t {
	case "sysbench":
		log.Println("Process sysbench results")
		cum, history, att := parser.ProcessSysbenchResult(log_file)

		// r.Runs[run_id].Stats.TPS = cum
		r.Runs[run_id].Stats.Summary.AllNodes.Op_throughput, err = strconv.ParseFloat(strings.Replace(cum, ",", "", -1), 64)
		if err != nil {
			log.Panicln("Error parsing op_throughput ", cum, ", error: ", err)
		}

		r.Runs[run_id].Stats.History = history

		// merge attribute into Stats
		for k, v := range att {
			r.Runs[run_id].Stats.Attributes[k] = v
		}

	case "sysbench-insert":
		log.Println("Process sysbench results")
		cum, history, att := parser.ProcessSysbenchInsertResult(log_file)

		// r.Runs[run_id].Stats.TPS = cum
		r.Runs[run_id].Stats.Summary.AllNodes.Op_throughput, err = strconv.ParseFloat(strings.Replace(cum, ",", "", -1), 64)
		if err != nil {
			log.Panicln("Error parsing op_throughput ", cum, ", error: ", err)
		}

		r.Runs[run_id].Stats.History = history
		r.Runs[run_id].Stats.Harness = "sysbench" // still keep the same harness name

		// merge attribute into Stats
		for k, v := range att {
			r.Runs[run_id].Stats.Attributes[k] = v
		}
	case "mongo-sim":
		log.Println("Processing mongo-sim results")
		result_ := parser.ProcessMongoSIMResult(log_file)

		// need merge the two Stats together. Will copy
		//	r.Runs[run_id].Stats.Summary = result_.Summary
		// log.Printf("%# v\n", result_.Summary)
		r.Runs[run_id].Stats.Summary.Nodes = make([]map[string]parser.NodeStats, 10, 10)
		copy(r.Runs[run_id].Stats.Summary.Nodes, result_.Summary.Nodes)
		// r.Runs[run_id].Stats.Testbed = result_.Testbed
		r.Runs[run_id].Stats.TestDriver = result_.TestDriver

		// merge attributes together
		for k, v := range result_.Attributes {
			if val, ok := r.Runs[run_id].Stats.Attributes[k]; ok {
				log.Println("Discard hammer-mc attribute[", k, "] = ", val, " with new value from mongo-sim ", v)
			}
			r.Runs[run_id].Stats.Attributes[k] = v
		}

	case "mongo-perf":
		log.Println("Processing mongo-perf results")
		result_ := parser.ProcessMongoPerfResult(log_file)

		for k, v := range result_ {
			_ = k
			_ = v
			/*
				Name    string
				Thread  int64
				Result  float64
				CV      string
				Version string
				GitSHA  string
			*/
			r.Runs[run_id].Stats.Workload = v.Name
			r.Runs[run_id].Stats.Attributes["nThread"] = v.Thread
			r.Runs[run_id].Stats.Attributes["CV"] = v.CV
			fmt.Println("server version is :", v.Version)
			r.Runs[run_id].Stats.ServerVersion = strings.Fields(v.Version)[2]
			r.Runs[run_id].Stats.ServerGitSHA = v.GitSHA
			r.Runs[run_id].Stats.Summary.AllNodes.Op_throughput = v.Result

			// print
			s, _ := json.MarshalIndent(r.Runs[run_id].Stats, "  ", "    ")
			os.Stdout.Write(s)
			fmt.Println("\n********")

			// report to server
			if len(report_url) != 0 {
				// report to report_url if it is not empty
				for _, rurl_ := range report_url {
					r, err := http.Post(rurl_, "application/json", bytes.NewBuffer(s))

					if err == nil {
						log.Println("Submit results to server succeeded with reponse:\n", r)
					} else {
						log.Panicln("Submit results failed with error: ", err)
					}
				}
			}
		}

		// mongo-perf will not report server stats since it is not meaningful
		return

	case "task":
		// will not report task
		return

	default:
		log.Println("no type infor, ignore results analyzing")

		// not return results
		return
	}

	// report pidstat here
	r.Runs[run_id].Stats.Server_Stats = make(map[string]parser.ServerStats)

	for k := 0; k < len(r.Runs[run_id].Servers); k++ {
		pidfile := run_dir + "/pidstat.log--" + r.Runs[run_id].Servers[k]
		stats := parser.ParsePIDStat(pidfile)

		// r.Runs[run_id].Stats.Server_Stats[replaceDot(r.Runs[run_id].Servers[k])] = make(ServerStats)
		// r.Runs[run_id].Stats.Server_Stats[r.Runs[run_id].Servers[k]][kk] = make([]parser.DataPoint, len(vv))
		// log.Println("++++> ", copy(r.Runs[run_id].Stats.Server_Stats[r.Runs[run_id].Servers[k]][kk], vv))
		//append(r.Runs[run_id].Stats.Server_Stats[r.Runs[run_id].Servers[k]][kk], vv)
		log.Println("\n\n++++++++-----+++++++++\n\n")
		//r.Runs[run_id].Stats.Server_Stats[replaceDot(r.Runs[run_id].Servers[k])].Cpu = make([]parser.DataPoint, len(stats["cpu"]), len(stats["cpu"]))
		// copy(r.Runs[run_id].Stats.Server_Stats[replaceDot(r.Runs[run_id].Servers[k])].Cpu, stats["cpu"])
		r.Runs[run_id].Stats.Server_Stats[replaceDot(r.Runs[run_id].Servers[k])] = stats
		// log.Println(r.Runs[run_id].Stats.Server_Stats[replaceDot(r.Runs[run_id].Servers[k])])
	}

	r.Runs[run_id].Stats.Throughput = r.Runs[run_id].Stats.Summary.AllNodes.Op_throughput
	s, _ := json.MarshalIndent(r.Runs[run_id].Stats, "  ", "    ")
	if len(report_url) != 0 {
		// report to report_url if it is not empty

		for _, rurl := range report_url {
			r, err := http.Post(rurl, "application/json", bytes.NewBuffer(s))

			if err == nil {
				log.Println("Submit results to server [", rurl, "] succeeded with reponse:\n", r)
			} else {
				log.Panicln("Submit results to server [", rurl, "] failed with error: ", err)
			}
		}
	}

	// print
	os.Stdout.Write(s)
	fmt.Println("\n********")
	//fmt.Printf("%# v", pretty.Formatter(r.Runs[run_id].Stats))
}
