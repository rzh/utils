package parser

import "testing"

func TestProcessSysbenchResult(t *testing.T) {
	cum, trend, att := ProcessSysbenchResult("sysbench.txt")

	if cum != "741.86" {
		t.Error("Cumulative TPS is %v, want 741.86", cum)
	}

	if trend[0] != "772.95" {
		t.Error("History IPS[0] is %v, want 772.95", trend[0])
	}

	if att["test-type"] != "sysbench" {
		t.Error("Attribute[\"test-type\"] is %v, want sysbench", att["test-type"])
	}

	// Thread[main,5,main]  writer threads           = 64
	if att["nThreads"] != "64" {
		t.Errorf("Attribute[nThreads] is %s, want 64", att["nThreads"])
	}

	// Thread[main,5,main]  collections              = 15
	if att["nCollections"] != "15" {
		t.Errorf("Attribute[nCollections] is %s, want 15", att["nCollections"])
	}

	// Thread[main,5,main]  documents per collection = 1,000,000
	if att["nCollectionSize"] != "1,000,000" {
		t.Errorf("Attribute[nCollectionSize] is %s, want 1,000,000", att["nCollectionSize"])
	}

	// Thread[main,5,main]  feedback seconds         = 10
	if att["nFeedbackSeconds"] != "10" {
		t.Errorf("Attribute[nFeedbackSeconds] is %s, want 10", att["nFeedbackSeconds"])
	}

	// Thread[main,5,main]  run seconds              = 600
	if att["nRunSeconds"] != "600" {
		t.Errorf("Attribute[nRunSeconds] is %s, want 600", att["nRunSeconds"])
	}

	// Thread[main,5,main]  oltp range size          = 100
	if att["oltp range size"] != "100" {
		t.Errorf("Attribute[oltp range size] is %s, want 600", att["oltp range size"])
	}

	// Thread[main,5,main]  oltp point selects       = 10
	if att["oltp point selects"] != "10" {
		t.Errorf("Attribute[oltp point selects] is %s, want 600", att["oltp point selects"])
	}

	// Thread[main,5,main]  oltp simple ranges       = 11
	if att["oltp simple ranges"] != "11" {
		t.Errorf("Attribute[oltp simple ranges] is %s, want 600", att["oltp simple ranges"])
	}

	// Thread[main,5,main]  oltp sum ranges          = 12
	if att["oltp sum ranges"] != "12" {
		t.Errorf("Attribute[oltp sum ranges] is %s, want 600", att["oltp sum ranges"])
	}

	// Thread[main,5,main]  oltp order ranges        = 13
	if att["oltp order ranges"] != "13" {
		t.Errorf("Attribute[oltp order ranges] is %s, want 600", att["oltp order ranges"])
	}

	// Thread[main,5,main]  oltp distinct ranges     = 14
	if att["oltp distinct ranges"] != "14" {
		t.Errorf("Attribute[oltp distinct ranges] is %s, want 600", att["oltp distinct ranges"])
	}

	// Thread[main,5,main]  oltp index updates       = 15
	if att["oltp index updates"] != "15" {
		t.Errorf("Attribute[oltp index updates] is %s, want 600", att["oltp index updates"])
	}

	// Thread[main,5,main]  oltp non index updates   = 16
	if att["oltp non index updates"] != "16" {
		t.Errorf("Attribute[oltp non index updates] is %s, want 600", att["oltp non index updates"])
	}

	// Thread[main,5,main]  write concern            = SAFE
	if att["write concern"] != "SAFE" {
		t.Errorf("Attribute[write concern] is %s, want 600", att["write concern"])
	}
}

func TestParsePIDStat(t *testing.T) {
	dps := ParsePIDStat("pidstat.txt")

	if dps.Process != "mongod" {
		t.Error("Pidstat process-type is " + dps.Process + " expecting mongod")
	}

	if dps.Cpu[0] != 37.36 {
		t.Error("Pidstat cpu[0] is ", dps.Cpu[0], " expecting 37.36")
	}

	if dps.Mem[1] != 10.09 {
		t.Error("Pidstat mem[1] is ", dps.Mem[1], " expecting 10.09")
	}
}

func TestParseMongoSIMStat(t *testing.T) {
	r := ProcessMongoSIMResult("mongo-sim.txt")

	if r.Summary.Nodes[0]["new_user"].Op_lat_total_micros != 2755 {
		t.Error("mongo-sim op_throughput is ", r.Summary.Nodes[0]["new_user"].Op_lat_total_micros, ", expecting 2755")
	}
	if r.Summary.Nodes[2]["post_message"].Op_count != 199 {
		t.Error("mongo-sim op_count is ", r.Summary.Nodes[0]["post_message"].Op_count, ", expecting 199")
	}

	//s, _ := json.MarshalIndent(r, "  ", "    ")
	//os.Stdout.Write(s)
	//fmt.Println(len(r.Summary.Nodes))
}

func TestParseMongoPerfResult(t *testing.T) {
	r := ProcessMongoPerfResult("mongo-perf.txt")

	if r == nil {
		t.Error("[mongo-perf] return value is nil")
	}

	if len(r) != 12 {
		t.Error("[mongo-perf] receive wrong number of results, received ", len(r), " expecting 12")
	}

	//fmt.Println(r)

	if r["Geo.within.center_TH-001"].Result != 928.11 {
		t.Error("[mongo-perf] receive wrong value of results, received ", r["Geo.within.center_TH-001"].Result, " expecting 928.11")
	}

	// FIXME need more test for average and CV
}
