package parser

import (
	"fmt"
	"testing"
)

func TestProcessSysbenchResult(t *testing.T) {
	cum, trend, att := ProcessSysbenchResult("test.txt")

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
	// Thread[main,5,main]  oltp point selects       = 10
	// Thread[main,5,main]  oltp simple ranges       = 1
	// Thread[main,5,main]  oltp sum ranges          = 1
	// Thread[main,5,main]  oltp order ranges        = 1
	// Thread[main,5,main]  oltp distinct ranges     = 1
	// Thread[main,5,main]  oltp index updates       = 1
	// Thread[main,5,main]  oltp non index updates   = 1
	// Thread[main,5,main]  write concern            = SAFE
}

func TestParsePIDStat(t *testing.T) {
	s, dps := ParsePIDStat("pidstat.txt")

	if s != "mongod" {
		t.Error("Pidstat process-type is " + s + " expecting mongod")
	}

	if dps["cpu"][0].d != "91.01" {
		t.Error("Pidstat cpu[0] is " + dps["cpu"][0].d + " expecting 91.01")
	}

	if dps["mem"][1].d != "23.29" {
		t.Error("Pidstat mem[1] is " + dps["mem"][1].d + " expecting 23.29")
	}
}

func TestParseMongoSIMStat(t *testing.T) {
	r := ProcessMongoSIMResult("mongo-sim.txt")

	if r.AllNodes.Op_per_second != 0 {
		t.Error("mongo-sim op_per_second is ", r.AllNodes.Op_per_second, ", expecting 0")
	}
	if r.Nodes[0]["st_staging_minutes"].Op_count != 43500 {
		t.Error("mongo-sim op_per_second is ", r.Nodes[0]["st_staging_minutes"].Op_count, ", expecting 100")
	}
}

func TestParseMongoPerfResult(t *testing.T) {
	r := ProcessMongoPerfResult("mongo-perf.txt")

	if r == nil {
		t.Error("[mongo-perf] return value is nil")
	}

	if len(r) != 12 {
		t.Error("[mongo-perf] receive wrong number of results, received ", len(r), " expecting 12")
	}

	fmt.Println(r)

	if r["Geo.within.center_TH-001"].Result != 928.11 {
		t.Error("[mongo-perf] receive wrong value of results, received ", r["Geo.within.center_TH-001"].Result, " expecting 928.11")
	}

	// FIXME need more test for average and CV
}
