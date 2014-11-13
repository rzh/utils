package parser

import "testing"

func TestProcessSysbenchInsertResult(t *testing.T) {
	cum, _, _ := ProcessSysbenchInsertResult("sysbench-insert.txt")

	if cum != "12,445.96" {
		t.Errorf("Cumulative IPS is %v, want 12,445.96", cum)
	}
}

func TestProcessSysbenchResult(t *testing.T) {
	cum, trend, att := ProcessSysbenchResult("sysbench.txt")

	if cum != "741.86" {
		t.Errorf("Cumulative TPS is %v, want 741.86", cum)
	}

	if trend[0] != "772.95" {
		t.Errorf("History IPS[0] is %v, want 772.95", trend[0])
	}

	if att["test-type"] != "sysbench" {
		t.Errorf("Attribute[\"test-type\"] is %v, want sysbench", att["test-type"])
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
		t.Errorf("Pidstat process-type is %.2s%s ", dps.Process, " expecting mongod")
	}

	if dps.Cpu[0] != 37.36 {
		t.Errorf("Pidstat cpu[0] is %.2f%s ", dps.Cpu[0], " expecting 37.36")
	}

	if dps.Mem[1] != 10.09 {
		t.Errorf("Pidstat mem[1] is %.2f%s ", dps.Mem[1], " expecting 10.09")
	}

	if dps.KB_rd[0] != 1017.00 {
		t.Errorf("Pidstat KB_rd[0] is %.2f%s ", dps.KB_rd[0], " expecting 1017.00")
	}

	if dps.KB_wr[0] != 24144.0 {
		t.Errorf("Pidstat KB_wr[0] is %.2f%s ", dps.KB_wr[0], " expecting 24144.00")
	}

	if dps.Cswch[0] != 29718.0 {
		t.Errorf("Pidstat Cswch[0] is %.2f%s ", dps.Cswch[0], " expecting 29718.00")
	}

	if dps.Nvcswch[1] != 1196.0 {
		t.Errorf("Pidstat Nvcswch[0] is %.2f%s ", dps.Nvcswch[1], " expecting 1196.00")
	}

	// test a different format
	dps = ParsePIDStat("pidstat2.txt")
	if dps.Process != "mongod" {
		t.Errorf("Pidstat process-type is " + dps.Process + " expecting mongod")
	}

	if dps.Cpu[0] != 29.09 {
		t.Errorf("Pidstat cpu[0] is ", dps.Cpu[0], " expecting 29.09")
	}

	if dps.Mem[1] != 25.59 {
		t.Errorf("Pidstat mem[1] is ", dps.Mem[1], " expecting 25.59")
	}

	if dps.Cswch[0] != 2271.0 {
		t.Errorf("Pidstat Cswch[0] is %.2f%s ", dps.Cswch[0], " expecting 2271.00")
	}

	if dps.Nvcswch[1] != 270492.0 {
		t.Errorf("Pidstat Nvcswch[0] is %.2f%s ", dps.Nvcswch[1], " expecting 270492.00")
	}
}

func TestParseMongoSIMStat(t *testing.T) {
	r := ProcessMongoSIMResult("mongo-sim.txt")

	if r.Summary.Nodes[0]["new_user"].Op_lat_total_micros != 2755 {
		t.Errorf("mongo-sim op_throughput is ", r.Summary.Nodes[0]["new_user"].Op_lat_total_micros, ", expecting 2755")
	}
	if r.Summary.Nodes[2]["post_message"].Op_count != 199 {
		t.Errorf("mongo-sim op_count is ", r.Summary.Nodes[0]["post_message"].Op_count, ", expecting 199")
	}
}

func TestParseMongoPerfResult(t *testing.T) {
	r := ProcessMongoPerfResult("mongo-perf.txt")

	if r == nil {
		t.Errorf("[mongo-perf] return value is nil")
	}

	if len(r) != 12 {
		t.Errorf("[mongo-perf] receive wrong number of results, received ", len(r), " expecting 12")
	}

	//fmt.Println(r)
	if r["Geo.within.center_TH-001"].Version != "db version: 2.6.5-rc2-pre-" {
		t.Errorf("[mongo-perf] receive wrong value of results, received ", r["Geo.within.center_TH-001"].Version, " expecting db version: 2.6.5-rc2-pre-")
	}

	if r["Geo.within.center_TH-001"].ClientVersion != "MongoDB shell version: 2.7.5-pre-" {
		t.Errorf("[mongo-perf] receive wrong value of results, received ", r["Geo.within.center_TH-001"].ClientVersion, " expecting db version: 2.6.5-rc2-pre-")
	}

	if r["Geo.within.center_TH-001"].Result != 928.11 {
		t.Errorf("[mongo-perf] receive wrong value of results, received ", r["Geo.within.center_TH-001"].Result, " expecting 928.11")
	}

	// FIXME need more test for average and CV
}

func TestProcessHammerResult(t *testing.T) {
	cum, trendRps, trendAvg, att := ProcessHammerResult("hammer.txt")

	if cum != "9567" {
		t.Errorf("Cumulative TPS is %v, want 741.86", cum)
	}

	if att["test-type"] != "hammer" {
		t.Errorf("Attribute[\"test-type\"] is %v, want sysbench", att["test-type"])
	}

	if trendRps[0] != "106059" {
		t.Errorf("History RPS[0] is %v, want 106059", trendRps[0])
	}

	if trendAvg[0] != "0.094142" {
		t.Errorf("History avgResponse[0] is %v, want 0.094142", trendAvg[0])
	}

	if att["nThreads"] != "4" {
		t.Errorf("Attribute[nThreads] is %s, want 64", att["nThreads"])
	}

	if att["avgResponseTime"] != "0.104305" {
		t.Errorf("Attribute[avgResponseTime] is %s, want 64", att["avgResponseTime"])
	}

	if att["errorRatio"] != "10.00" {
		t.Errorf("Attribute[errorRatio] is %s, want 10.00", att["errorRatio"])
	}

	if att["slowRatio"] != "18.00" {
		t.Errorf("Attribute[slowRatio] is %s, want 18.00", att["slowRatio"])
	}
}
