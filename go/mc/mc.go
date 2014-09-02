/*
 a file to do automation to control test

 This is not ready for general purpose use, you can hack to use it but please do not check in your local change.
 The file here is for demo how control center works.

 This is also need install some tools in the server to be monitorred
  - dstat
  - dstat mongodb plugin
  - pidstat
  - iostat/systat
*/

package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	str "strings"
	"time"

	"github.com/kr/pretty"
	"github.com/rzh/utils/go/mc/parser"

	"github.com/ActiveState/tail"
)

var (
	ssh_server string
	ssh_client string
	ssh_hammer string
	pem_file   string
	// run_id     string
	run    string
	config string
	test   string
)

const report_folder string = "reports"

type ITask interface {
	// Prepare(server string, pemfile string)
	Run()
	Cleanup()
}

// definition of processes
type Process struct {
	Pid  string // pid
	Name string // executable name, mongod or mongos
	Cmd  string // how this is runned
}

// definition of Stat

type Stats struct {
	TPS string
	// Run_Date   string
	Start_Time   int64 //epoch time, time.Now().Unix()
	End_Time     int64
	ID           string
	Type         string // hammertime, sysbench, mongo-sim
	History      []string
	Server_Stats map[string]parser.ServerStats
	Attributes   map[string]interface{}
}

// start of definition of Task
type Task struct {
	Cmd_exec *exec.Cmd
	Ssh_url  string
	Pem_file string
	Logfile  string
	Cmd      string
	Out      bytes.Buffer // buffer to hold output of exec_cmd
	Outfile  *os.File
}

func (p *Task) Run() {
	if p.Pem_file != "" {
		p.Cmd_exec = exec.Command(
			"/usr/bin/ssh",
			"-i", p.Pem_file,
			p.Ssh_url,
			p.Cmd)
	} else {
		p.Cmd_exec = exec.Command(
			"/usr/bin/ssh",
			p.Ssh_url,
			p.Cmd)
	}

	go func() {
		outfile, err := os.Create(p.Logfile)
		if err != nil {
			panic(err)
		}
		// defer outfile.Close()
		p.Cmd_exec.Stdout = outfile
		p.Outfile = outfile

		err = p.Cmd_exec.Start()
		if err != nil {
			log.Fatal("failed with -> ", err)
		}
		p.Cmd_exec.Wait()
	}()
}

func (p *Task) Cleanup() {
	// stop cmd and close log file
	p.Outfile.Close()
	p.Cmd_exec.Process.Kill()
}

// end of Task

type HammerTask struct {
	Run_id        string   `json:run_id`
	Cmd           string   `json:cmd`
	Hammer_folder string   `json:hammer_folder`
	Clients       []string `json:clients`
	Servers       []string `json:servers`

	// log files to be collected from client and server
	Client_logs []string `json:client_logs`
	Server_logs []string `json:server_logs`

	Type string `json:type`

	Stats Stats
}

type TheRun struct {
	tasks   []Task
	Runs    []HammerTask `json:runs`
	PemFile string       `json:PemFile`

	MongoDBPath        string
	MongoStagingDBPath string
	MongoDCmd          string

	run_dir string
}

func (r *TheRun) runServerCmd(server string, cmd string) ([]byte, error) {
	if r.PemFile != "" {
		return exec.Command(
			"/usr/bin/ssh",
			"-i", r.PemFile,
			server,
			cmd).Output()
	} else {
		return exec.Command(
			"/usr/bin/ssh",
			server,
			cmd).Output()
	}
}

func (r *TheRun) findMongoD_CMD(server string) string {
	_cmd, err := r.runServerCmd(
		server,
		"/bin/ps -ef | grep mongod | grep -v grep | awk 'BEGIN {ORS=\" \"} {for(i=8;i<=NF;++i) print $i} END {print \"\n\"}'")

	if err == nil {
		return string(_cmd)
	}

	_cmd, err = r.runServerCmd(
		server,
		"/bin/ps -ef | grep mongos | grep -v grep | awk 'BEGIN {ORS=\" \"} {for(i=8;i<=NF;++i) print $i} END {print \"\n\"}'")

	if err == nil {
		return string(_cmd)
	}
	return ""
}

func (r *TheRun) findMongoD_PID(server string) string {
	// find out mongod pid
	var err error
	var _pid []byte

	if r.PemFile != "" {
		_pid, err = exec.Command(
			"/usr/bin/ssh",
			"-i", r.PemFile,
			server,
			"/bin/ps -e | grep mongod | grep -v grep | awk '{print $1}'").Output()
	} else {
		_pid, err = exec.Command(
			"/usr/bin/ssh",
			server,
			"/bin/ps -e | grep mongod | grep -v grep | awk '{print $1}'").Output()
	}

	// exec.Command("/bin/sh", "-c", "/bin/ps -e | grep mongod | grep -v grep | awk '{print $1}'").Output()

	if err != nil {
		log.Fatalln("error getting MongoD PID: ", err)
	}

	fmt.Println("PD is |", string(_pid[:len(_pid)-1]), "|")

	pid, err := strconv.Atoi(string(_pid[:len(_pid)-1]))

	if err != nil {
		log.Fatalln("cannot convert mongod pid (", string(_pid), ") to int")
	}
	fmt.Println("PID of Mongod is -> ", pid)

	return string(_pid[:len(_pid)-1])
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func joinstr(s1 string, s2 string, sep string) string {
	r := str.Join([]string{s1, s2}, sep)
	fmt.Println(r)

	return r
}

func (r *TheRun) RunClientTasks(i int, run_dir string) {
	var cmd *exec.Cmd

	log.Println("\n\n>>>>>>>>>>>>\n|    start test <" + r.Runs[i].Run_id + ">\n<<<<<<<<<<<<")
	// var cp_cmd *exec.Cmd

	if len(r.Runs[i].Clients) == 0 {
		// local
		cmd = exec.Command(r.Runs[i].Cmd)
		if r.Runs[i].Hammer_folder == "" {
			r.Runs[i].Hammer_folder = "."
		}
	} else {
		// via ssh
		if r.PemFile != "" {
			cmd = exec.Command(
				"/usr/bin/ssh",
				"-i", r.PemFile,
				r.Runs[i].Clients[0],
				r.Runs[i].Cmd)
		} else {
			cmd = exec.Command(
				"/usr/bin/ssh",
				r.Runs[i].Clients[0],
				r.Runs[i].Cmd)
		}

		if r.Runs[i].Hammer_folder == "" {
			r.Runs[i].Hammer_folder = "" // fixme:
		}
	}

	log_file := run_dir + "/test_screen_capture.log--" + r.Runs[i].Clients[0]

	outfile, err := os.Create(log_file)
	if err != nil {
		panic(err)
	}
	defer outfile.Close()
	cmd.Stdout = outfile
	cmd.Stderr = outfile

	// start tail log file to stdout
	go func() {
		t, err := tail.TailFile(log_file, tail.Config{Follow: true})

		if err != nil {
			panic("cannot not tail hammer logfile ")
		}
		for line := range t.Lines {
			fmt.Println(line.Text)
		}
	}()

	r.Runs[i].Stats.Start_Time = time.Now().Unix()
	err = cmd.Run()
	if err != nil {
		// do not quit if this client return error code FIXME
		// log.Fatal("Hammer client failed with -> ", err)
	}
	r.Runs[i].Stats.End_Time = time.Now().Unix()

	time.Sleep(5 * time.Second) //chill for 5 second to collect some system stats after test done

	// scp file here
	// FIXME: only take the first client for now. Don't believe we need this at all
	for j := 0; j < len(r.Runs[i].Client_logs); j++ {

		// copy client logs
		if len(r.Runs[i].Clients) == 0 {
			// local
			ee, er := exec.Command("/bin/cp",
				r.Runs[i].Client_logs[j]+" "+run_dir+"/"+r.Runs[i].Client_logs[j]+"--"+
					r.Runs[i].Clients[0]).Output()
			if er != nil {
				log.Fatal("Failed to copy result of client due to -> ", er, string(ee))
			}
		} else {
			// via ssh
			log.Println(r.Runs[i])
			if r.PemFile != "" {
				ee, er := exec.Command(
					"/bin/sh", "-c",
					fmt.Sprintf("/usr/bin/scp -i %s %s%s%s %s",
						r.PemFile, r.Runs[i].Clients[0], ":",
						r.Runs[i].Client_logs[j],
						run_dir+"/"+r.Runs[i].Client_logs[j]+"--"+r.Runs[i].Clients[0])).Output()
				log.Println(run_dir + "/" + r.Runs[i].Client_logs[j] + "--" + r.Runs[i].Clients[0])
				if er != nil {
					log.Fatal("Failed to copy result of client due to -> ", er, string(ee))
				}
				log.Println(ee)
			} else {
				ee, er := exec.Command(
					"/bin/sh", "-c",
					fmt.Sprintf("/usr/bin/scp %s%s%s %s", r.Runs[i].Clients[0],
						":", r.Runs[i].Client_logs[j],
						run_dir+"/"+r.Runs[i].Client_logs[j]+"--"+r.Runs[i].Clients[0])).Output()
				if er != nil {
					log.Fatal("Failed to copy result of client due to -> ", er, string(ee))
				}
				log.Println(ee)
			}
		}
	} //for_j for Client_logs

	// copy server logs
	for j := 0; j < len(r.Runs[i].Server_logs); j++ {
		if len(r.Runs[i].Servers) == 0 {
			// local
			ee, er := exec.Command("/bin/cp",
				r.Runs[i].Server_logs[j]+" "+run_dir+"/"+r.Runs[i].Server_logs[j]+"--"+
					r.Runs[i].Servers[0]).Output()
			if er != nil {
				log.Fatal("Failed to copy result of client due to -> ", er, string(ee))
			}
		} else {
			// via ssh
			log.Println(r.Runs[i])
			for k := 0; k < len(r.Runs[i].Servers); k++ {
				if r.PemFile != "" {
					ee, er := exec.Command(
						"/bin/sh", "-c",
						fmt.Sprintf("/usr/bin/scp -i %s %s%s%s %s",
							r.PemFile, r.Runs[i].Servers[k], ":",
							r.Runs[i].Server_logs[j],
							run_dir+"/"+r.Runs[i].Server_logs[j]+"--"+r.Runs[i].Servers[k])).Output()
					log.Println(run_dir + "/" + r.Runs[i].Server_logs[j] + "--" + r.Runs[i].Servers[k])
					if er != nil {
						log.Fatal("Failed to copy result of client due to -> ", er, string(ee))
					}
					log.Println(ee)
				} else {
					ee, er := exec.Command(
						"/bin/sh", "-c",
						fmt.Sprintf("/usr/bin/scp %s%s%s %s", r.Runs[i].Servers[k],
							":", r.Runs[i].Server_logs[j],
							run_dir+"/"+r.Runs[i].Server_logs[j]+"--"+r.Runs[i].Servers[k])).Output()
					if er != nil {
						log.Fatal("Failed to copy result of client due to -> ", er, string(ee))
					}
					log.Println(ee)
				}
			}
		} // for_k for Servers
	} //for_j for Server_logs

	r.reportResults(i, log_file, run_dir)
}

func (r *TheRun) reportResults(run_id int, log_file string, run_dir string) {
	// this is the place to analyze results.
	t := str.ToLower(r.Runs[run_id].Type)
	r.Runs[run_id].Stats.Type = t
	r.Runs[run_id].Stats.ID = r.Runs[run_id].Run_id

	// cache run first
	//rr, _ := json.Marshal(r.Runs[run_id])
	rr := r.Runs[run_id]
	r.Runs[run_id].Stats.Attributes = make(map[string]interface{})
	r.Runs[run_id].Stats.Attributes["run-by"] = "hammer-mc"
	r.Runs[run_id].Stats.Attributes["hammer-mc-cmd"] = HammerTask{Run_id: rr.Run_id,
		Cmd: rr.Cmd, Clients: rr.Clients, Servers: rr.Servers,
		Client_logs: rr.Client_logs, Server_logs: rr.Server_logs,
		Type: rr.Type}

	switch t {
	case "sysbench":
		log.Println("analyzing sysbench results")
		cum, history, att := parser.ProcessSysbenchResult(log_file)

		r.Runs[run_id].Stats.TPS = cum
		r.Runs[run_id].Stats.History = history

		// merge attribute into Stats
		for k, v := range att {
			r.Runs[run_id].Stats.Attributes[k] = v
		}

	default:
		log.Println("no type infor, ignore results analyzing")
	}

	// report pidstat here
	r.Runs[run_id].Stats.Server_Stats = make(map[string]parser.ServerStats)
	for k := 0; k < len(r.Runs[run_id].Servers); k++ {
		pidfile := run_dir + "/pidstat.log--" + r.Runs[run_id].Servers[k]
		_, stats := parser.ParsePIDStat(pidfile)

		r.Runs[run_id].Stats.Server_Stats[r.Runs[run_id].Servers[k]] = make(parser.ServerStats)
		for kk, vv := range stats {
			// r.Runs[run_id].Stats.Server_Stats[r.Runs[run_id].Servers[k]][kk] = make([]parser.DataPoint, len(vv))
			// log.Println("++++> ", copy(r.Runs[run_id].Stats.Server_Stats[r.Runs[run_id].Servers[k]][kk], vv))
			//append(r.Runs[run_id].Stats.Server_Stats[r.Runs[run_id].Servers[k]][kk], vv)
			r.Runs[run_id].Stats.Server_Stats[r.Runs[run_id].Servers[k]][kk] = vv
		}
	}

	// print
	/*
		s, _ := json.MarshalIndent(r.Runs[run_id].Stats, "  ", "  ")
		os.Stdout.Write(s)
		fmt.Println("********")
	*/
	fmt.Printf("%# v", pretty.Formatter(r.Runs[run_id].Stats))
}

func (r *TheRun) monitorServer(server string, run_dir string) {
	pid := r.findMongoD_PID(server)

	r.tasks = append(r.tasks, Task{
		Ssh_url:  server,
		Pem_file: r.PemFile,
		Logfile:  joinstr(run_dir, "/pidstat.log--"+server, ""),
		Cmd:      joinstr("pidstat 1 -Ihtruwd -p", pid, " ")})

	r.tasks = append(r.tasks, Task{
		Ssh_url:  server,
		Pem_file: r.PemFile,
		Logfile:  joinstr(run_dir, "/iostat.log--"+server, ""),
		Cmd:      "iostat 1"})

	// r.tasks = append(r.tasks, Task{
	// 	Ssh_url:  ssh_server,
	// 	Pem_file: pem_file,
	// 	Logfile:  joinstr(run_dir, "/dstat.log", ""),
	// 	Cmd:      "dstat -tcm -dn --disk-util --disk-t"})
}

func (r *TheRun) stagingMongoD(server string) {
	_, err := r.runServerCmd(server, fmt.Sprint("rm -rf ", r.MongoDBPath, "/*"))
	panicOnError(err)

	_, err = r.runServerCmd(server, fmt.Sprint("cp ", r.MongoStagingDBPath, "/* ", r.MongoDBPath))
	panicOnError(err)
}

/*
func (r *TheRun) stagingDBs() {
	// staging mongo for all the server

		for _, server := range r.Server {
			r.stagingMongoD(server)
		}
}
*/

func (r *TheRun) StartAllServerMonitorTasks(run_id int, run_dir string) {
	// first filling the tasks to monitor server,

	// make sure tasks array is empty
	if len(r.tasks) != 0 {
		log.Fatal("Try to start a new run when old task still going")
	}

	// create all the tasks
	for i := 0; i < len(r.Runs[run_id].Servers); i++ {
		r.monitorServer(r.Runs[run_id].Servers[i], run_dir)
	}

	// start all tasks
	for i := 0; i < len(r.tasks); i++ {
		r.tasks[i].Run()
	}
}

func (r *TheRun) StopAllTasks() {
	// stop all tasks

	for i := 0; i < len(r.tasks); i += 1 {
		r.tasks[i].Cleanup()
	}

	// clean the task array to empty
	r.tasks = nil

	// give 25 second pause
	// disable for now, FIXME
	// time.Sleep(25 * time.Second)
}

func (r *TheRun) Run(run_dir string) {
	r.run_dir = run_dir

	for i := 0; i < len(r.Runs); i++ {
		if str.ToLower(r.Runs[i].Run_id) == "staging" {
			// r.stagingDBs()
		} else {
			// make the folder here
			log_dir := joinstr(run_dir, r.Runs[i].Run_id, "/")
			err := os.Mkdir(log_dir, os.ModePerm)
			if err != nil {
				log.Panicln("failed to create ", log_dir)
			}

			r.StartAllServerMonitorTasks(i, log_dir)
			r.RunClientTasks(i, log_dir)
			r.StopAllTasks()
		}
	}
}

func initRunID(run string) string {
	if run == "" {
		fmt.Println("Please specify a run ID with -run!")
		os.Exit(1)
	}

	// check whether ./run exists, if not create it
	has_reports_folder, _ := exists("./" + report_folder + test)
	if !has_reports_folder {
		// create ./runs
		os.Mkdir("./"+report_folder+test, os.ModePerm)
	}

	// first make sure "./runs/run_id" folder does not exist
	run_dir := joinstr("./"+report_folder+test+"/", run, "")
	has_run_dir, _ := exists(run_dir)

	if has_run_dir {
		log.Fatal("the run_id ", run, " already exists!")
	}

	// then create the folder
	err := os.Mkdir(run_dir, os.ModePerm)
	if err != nil {
		log.Println("Failed to create ", run_dir)
	}

	return run_dir
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

// parse and return file's header & last line
func getCsvSummary(f string) (string, string) {
	csvF, err := os.Open(f)
	panicOnError(err)

	defer csvF.Close()

	bio := csv.NewReader(csvF)
	line, err := bio.Read()
	panicOnError(err)

	var li []string
	var lastline []string
	var lastlastline []string

	for err == nil {
		lastlastline = lastline
		lastline = li
		li, err = bio.Read()
	}

	return str.Join(line, ","), str.Join(lastlastline, ",")
}

func summarizeCSV(file string, h string) string {
	var header string
	var li string

	header, li = getCsvSummary(file)
	if h != "" {
		if header != h {
			fmt.Println("filename,", header)
		}
	} else {
		fmt.Println("filename,", header)
	}
	fmt.Println(file, ",", li)

	return header
}

func summarizeFolder(folder string) {
	folder_list, err := ioutil.ReadDir(folder)

	if err != nil {
		log.Panicln("reading folder ", folder, " failed with error : ", err)
	}

	var h string

	for _, value := range folder_list {
		_file := fmt.Sprint(folder, "/", value.Name(), "/perf_test_data.csv")
		e, _ := exists(_file)
		if e {
			h = summarizeCSV(_file, h)
		} else {
			fmt.Println(_file, "doesn't exist!")
		}
	}
}

func init() {
	flag.StringVar(&run, "run", "", "ID for the run")
	flag.StringVar(&config, "config", "", "Config JSON for the run")
	flag.StringVar(&test, "test", "", "Suffix for the report folder")
}

func main() {
	flag.Parse()

	run_dir := initRunID(run)

	if config == "" {
		config = "./runner.json"
	}

	// demarshal run file
	var r TheRun

	// read json file
	jsonBlob, e := ioutil.ReadFile(config)
	if e != nil {
		fmt.Printf("File error: %v\n", e)
		os.Exit(1)
	}

	err := json.Unmarshal(jsonBlob, &r)
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Printf("%+v", r)

	r.Run(run_dir)

	// fmt.Println(">>> SUMMARY <<<<\n\n\n")
	// disable for now, FIXME
	// summarizeFolder(run_dir)
}
