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

type Testbed struct {
	Type    string                       `json:type`
	Servers map[string]map[string]string `json:"servers"`
}

// definition of Stat to be reported to dashboard
type Stats struct {
	Harness       string                 `json:"harness"`
	Workload      string                 `json:"workload"`
	ServerVersion string                 `json:"server_version"`
	ServerGitSHA  string                 `json:"server_git_hash"`
	Attributes    map[string]interface{} `json:"attributes"`
	Testbed       Testbed                `json:"test_bed"`
	Summary       parser.StatsSummary    `json:"summary"`

	TPS string
	// Run_Date   string
	Start_Time int64 `json:"start_time"` //epoch time, time.Now().Unix()
	End_Time   int64 `json:"end_time"`
	ID         string
	// Type         string // hammertime, sysbench, mongo-sim
	History      []string
	Server_Stats map[string]parser.ServerStats
}

func replaceDot(s string) string {
	return strings.Replace(s, ".", "_", -1)
}

func (r *TheRun) reportResults(run_id int, log_file string, run_dir string) {
	// this is the place to analyze results.
	t := strings.ToLower(r.Runs[run_id].Type)
	// r.Runs[run_id].Stats.Type = t
	r.Runs[run_id].Stats.Harness = t
	r.Runs[run_id].Stats.ID = r.Runs[run_id].Run_id
	r.Runs[run_id].Stats.Workload = r.Runs[run_id].Run_id

	// cache run first
	//rr, _ := json.Marshal(r.Runs[run_id])
	rr := r.Runs[run_id]
	r.Runs[run_id].Stats.Attributes["run-by"] = "hammer-mc"
	r.Runs[run_id].Stats.Attributes["hammer-mc-cmd"] = HammerTask{Run_id: rr.Run_id,
		Cmd: rr.Cmd, Clients: rr.Clients, Servers: rr.Servers,
		Client_logs: rr.Client_logs, Server_logs: rr.Server_logs,
		Type: rr.Type}

	var err error
	switch t {
	case "sysbench":
		log.Println("analyzing sysbench results")
		cum, history, att := parser.ProcessSysbenchResult(log_file)

		r.Runs[run_id].Stats.TPS = cum
		r.Runs[run_id].Stats.Summary.AllNodes.Op_per_second, err = strconv.ParseFloat(strings.Replace(cum, ",", "", -1), 64)
		if err != nil {
			log.Panicln("Error parsing op_per_second ", cum, ", error: ", err)
		}

		r.Runs[run_id].Stats.History = history

		// merge attribute into Stats
		for k, v := range att {
			r.Runs[run_id].Stats.Attributes[k] = v
		}

	case "mongo-sim":
		log.Println("processing mongo-sim results")
		result_ := parser.ProcessMongoSIMResult(log_file)

		r.Runs[run_id].Stats.Summary = result_

	default:
		log.Println("no type infor, ignore results analyzing")
	}

	// report pidstat here
	r.Runs[run_id].Stats.Server_Stats = make(map[string]parser.ServerStats)
	for k := 0; k < len(r.Runs[run_id].Servers); k++ {
		pidfile := run_dir + "/pidstat.log--" + r.Runs[run_id].Servers[k]
		_, stats := parser.ParsePIDStat(pidfile)

		r.Runs[run_id].Stats.Server_Stats[replaceDot(r.Runs[run_id].Servers[k])] = make(parser.ServerStats)
		for kk, vv := range stats {
			// r.Runs[run_id].Stats.Server_Stats[r.Runs[run_id].Servers[k]][kk] = make([]parser.DataPoint, len(vv))
			// log.Println("++++> ", copy(r.Runs[run_id].Stats.Server_Stats[r.Runs[run_id].Servers[k]][kk], vv))
			//append(r.Runs[run_id].Stats.Server_Stats[r.Runs[run_id].Servers[k]][kk], vv)
			r.Runs[run_id].Stats.Server_Stats[replaceDot(r.Runs[run_id].Servers[k])][kk] = vv
		}
	}

	// print
	s, _ := json.MarshalIndent(r.Runs[run_id].Stats, "  ", "    ")
	os.Stdout.Write(s)
	fmt.Println("\n********")
	//fmt.Printf("%# v", pretty.Formatter(r.Runs[run_id].Stats))

	report_url = "http://54.68.84.192:8080/api/v1/results"
	if report_url != "" {
		// report to report_url if it is not empty
		r, err := http.Post(report_url, "application/json", bytes.NewBuffer(s))

		if err == nil {
			log.Println("Submit result to server, reponse: ", r)
		} else {
			log.Panicln("Submit results failed with error: ", err)
		}
	}
}
