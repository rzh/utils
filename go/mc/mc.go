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
	"regexp"
	"strconv"
	str "strings"
	"time"

	"github.com/ActiveState/tail"
	"github.com/kr/pretty"
	"github.com/rzh/utils/go/mc/parser"
)

var (
	ssh_server string
	ssh_client string
	ssh_hammer string
	pem_file   string
	// run_id     string
	run        string
	autorun    bool
	submitDyno bool
	config     string
	test       string
	report_url []string
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
	Run_id        string   `json:"run_id"`
	Cmd           string   `json:"cmd"`
	Hammer_folder string   `json:"hammer_folder"`
	Clients       []string `json:"clients"`
	Servers       []string `json:"servers"`
	Testbed       string   `json:"testbed",omitempty`

	// log files to be collected from client and server
	Client_logs []string `json:"client_logs"`
	Server_logs []string `json:"server_logs"`

	Type string `json:type`

	Stats parser.Stats
}

type TheRun struct {
	tasks    []Task       `json:"Tasks"`
	Runs     []HammerTask `json:"runs"`
	PemFile  string       `json:"PemFile",omitempty`
	Testbed  string       `json:"testbed",omitempty`
	MongoURL string       `json:"mongo_url",omitempty`

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

func cleanup(lines []string) []string {
	// need take out lines has ISODate here, not standard JSON.
	re := regexp.MustCompile(": ISODate")
	reNum := regexp.MustCompile(": NumberLong")
	for i := 0; i < len(lines); i++ {
		if re.MatchString(lines[i]) {
			// remove this line, and reset i to i-1
			lines = append(lines[:i], lines[i+1:]...)
			i = i - 1
		} else if reNum.MatchString(lines[i]) {
			ss := str.Replace(lines[i], "NumberLong", "", -1)
			ss = str.Replace(ss, "(", "", -1)
			ss = str.Replace(ss, ")", "", -1)
			lines[i] = ss
		}
	}
	return lines
}

func (r *TheRun) runMongoCMD(server, cmd string) []byte {
	output, err := r.runServerCmd(server,
		"~/mongo --norc --eval \"print('serverBuildInfo');printjson("+cmd+")\"")

	if err != nil {
		log.Panicln("Failed to run {", cmd, "} from server [", server, "] with error [", err, "]")
	}

	lines := str.Split(string(output), "\n")
	lines = cleanup(lines) // take out non-JSON format

	if lines[3] == "undefined" {
		return nil
	}
	return []byte(str.Join(lines[3:], "\n"))
}

func (r *TheRun) findMongoD_Info(run_id int) parser.ServerInfo {
	var buildInfo parser.MongodBuildInfo
	var hostInfo parser.MongoHostInfo
	var storageEngine parser.StorageEngineInfo

	var server string
	var err error

	if len(r.Runs[0].Servers) > 0 {
		server = r.Runs[run_id].Servers[0]
	} else {
		server = ""
	}

	// FIXME: need find where is mongo executable here. right now, just use a symbol link.

	// -- serverBuildInfo
	/*
		output, err := r.runServerCmd(server,
			"~/mongo --norc --eval \"print('serverBuildInfo');printjson(db.serverBuildInfo())\"")

		if err != nil {
			log.Panicln("Failed to find serverBuildInfo for server [", server, "] with error [", err, "]")
		}

		lines := str.Split(string(output), "\n")
	*/

	output_ := r.runMongoCMD(server, "db.serverBuildInfo()")
	err = json.Unmarshal(output_, &buildInfo)
	if err != nil {
		log.Panicln("Cannot unmarshal serverBuildInfo with error: ", err, "\n\n", string(output_))
	}

	fmt.Printf("%# v\n", pretty.Formatter(buildInfo))

	// -- hostInfo

	/*
		output, err = r.runServerCmd(server,
			"~/mongo --norc --eval \"print('hostInfo');printjson(db.hostInfo())\"")

		if err != nil {
			log.Panicln("Failed to find serverBuildInfo for server [", server, "] with error [", err, "]")
		}

		lines = str.Split(string(output), "\n")

	*/

	output_ = r.runMongoCMD(server, "db.hostInfo()")
	err = json.Unmarshal(output_, &hostInfo)

	if err != nil {
		log.Panicln("Cannot unmarshal hostInfo with error: ", err, "\nOutput:\n", string(output_))
	}

	fmt.Printf("%# v\n", pretty.Formatter(hostInfo))

	// not get storageEngine information
	/*
		output, err = r.runServerCmd(server,
			"~/mongo --norc --eval \"print('storageEngine');printjson(db.serverStatus().storageEngine)\"")

		if err != nil {
			log.Panicln("Failed to find storageEngine for server [", server, "] with error [", err, "]")
		}

		lines = str.Split(string(output), "\n")

		lines = cleanup(lines)
	*/

	output_ = r.runMongoCMD(server, "db.serverStatus().storageEngine")
	if len(output_) == 0 {
		storageEngine.Name = "mmapv0" // set to v0 for legacy version
	} else {
		err = json.Unmarshal(output_, &storageEngine)
		if err != nil {
			log.Panicln("Cannot unmarshal serverStatus with error: ", err, "\nOutput:\n", string(output_))
		}
	}

	fmt.Printf("%# v\n", pretty.Formatter(storageEngine))

	return parser.ServerInfo{BuildInfo: buildInfo, HostInfo: hostInfo, StorageEngine: storageEngine}
}

func (r *TheRun) findMongoD_PID(server string) string {
	// find out mongod pid
	var err error
	var _pid []byte

	_pid, err = r.runServerCmd(server, "/bin/pidof mongod")
	if err != nil {
		// now to try /sbin/pidof
		_pid, err = r.runServerCmd(server, "/sbin/pidof mongod")

		if err != nil {
			log.Fatalln("Failed to find MongoD PID: ", err)
		}
	}

	// TODO: make sure check for array here, could have multiple pids
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

	// first try to find out server buildInfo
	serverInfo := r.findMongoD_Info(i)

	r.Runs[i].Stats.ServerVersion = serverInfo.BuildInfo.Version
	r.Runs[i].Stats.ServerGitSHA = serverInfo.BuildInfo.GitVersion
	r.Runs[i].Stats.Attributes = make(map[string]interface{})
	r.Runs[i].Stats.Testbed.Type = "standalone"
	r.Runs[i].Stats.Testbed.Servers.Mongod = make([]parser.ServerInfo, 1, 1)
	r.Runs[i].Stats.Testbed.Servers.Mongod[0] = serverInfo
	r.Runs[i].Stats.StorageEngine = serverInfo.StorageEngine.Name

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

	var serverStatus_before parser.MongoServerStatus
	output_ := r.runMongoCMD(r.Runs[i].Servers[0], "db.serverStatus().opcounters")
	err = json.Unmarshal(output_, &serverStatus_before.Opcounters)
	if err != nil {
		log.Panicln("Cannot unmarshal serverStatus with error: ", err, "\nOutput:\n", string(output_))
	}

	pid := r.findMongoD_PID(r.Runs[i].Servers[0])
	var procstat_before, procstat_after string

	// to capture the /proc/$pid/stat here
	_procstat_before, _err := r.runServerCmd(r.Runs[i].Servers[0], "/bin/cat /proc/"+pid+"/stat")

	if _err != nil {
		procstat_before = "failed to get /proc stat before the run"
	} else {
		procstat_before = string(_procstat_before)
	}
	log.Println("Proc stat before test --> ", procstat_before)

	r.Runs[i].Stats.Start_Time.Date = time.Now().UnixNano() / int64(time.Millisecond)
	err = cmd.Run()
	if err != nil {
		// do not quit if this client return error code FIXME
		// log.Fatal("Hammer client failed with -> ", err)
	}

	r.Runs[i].Stats.End_Time.Date = time.Now().UnixNano() / int64(time.Millisecond)

	// to capture the /proc/$pid/stat here
	_procstat_after, _err := r.runServerCmd(r.Runs[i].Servers[0], "/bin/cat /proc/"+pid+"/stat")

	if _err != nil {
		procstat_after = "failed to get /proc stat after the run"
	} else {
		procstat_after = string(_procstat_after)
	}
	log.Println("Proc stat after test --> ", procstat_after)

	// take second serverStatus here
	var serverStatus_after parser.MongoServerStatus
	output_ = r.runMongoCMD(r.Runs[i].Servers[0], "db.serverStatus().opcounters")
	err = json.Unmarshal(output_, &serverStatus_after.Opcounters)
	if err != nil {
		log.Panicln("Cannot unmarshal serverStatus with error: ", err, "\nOutput:\n", string(output_))
	}

	r.Runs[i].Stats.TestRunTime = r.Runs[i].Stats.End_Time.Date - r.Runs[i].Stats.Start_Time.Date

	// FIXME check error here
	utime_before, _ := strconv.Atoi(str.Fields(procstat_before)[13])
	stime_before, _ := strconv.Atoi(str.Fields(procstat_before)[14])

	utime_after, _ := strconv.Atoi(str.Fields(procstat_after)[13])
	stime_after, _ := strconv.Atoi(str.Fields(procstat_after)[14])

	var findTotalOps = func(before, after parser.MongoServerStatus) int64 {
		var t int64 = 0
		t += after.Opcounters.Insert
		t += after.Opcounters.Query
		t += after.Opcounters.Update
		t += after.Opcounters.Delete
		t += after.Opcounters.GetMore
		t += after.Opcounters.Command
		t -= before.Opcounters.Insert
		t -= before.Opcounters.Query
		t -= before.Opcounters.Update
		t -= before.Opcounters.Delete
		t -= before.Opcounters.GetMore
		t -= before.Opcounters.Command

		return t
	}

	r.Runs[i].Stats.Utime = utime_after - utime_before
	r.Runs[i].Stats.Stime = stime_after - stime_before
	r.Runs[i].Stats.TotalOps = findTotalOps(serverStatus_before, serverStatus_after)
	r.Runs[i].Stats.TickPerOp = float64(r.Runs[i].Stats.Utime+r.Runs[i].Stats.Stime) / float64(r.Runs[i].Stats.TotalOps)

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
		Cmd:      "iostat -x 1"})

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
	var report_url_argv string
	flag.StringVar(&run, "run", "", "ID for the run")
	flag.BoolVar(&submitDyno, "submit", true, "Submit results to Dyna, defauilt to true")
	flag.StringVar(&config, "config", "", "Config JSON for the run")
	flag.StringVar(&test, "test", "", "Suffix for the report folder")
	flag.StringVar(&report_url_argv, "report", "", "URL to report test results")

	if report_url_argv != "" {
		report_url = str.Split(report_url_argv, ",")
	}
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
}
