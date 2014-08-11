package parser

import "testing"

func TestProcessSysbenchResult(t *testing.T) {
	cum, trend := ProcessSysbenchResult("test.txt")

	if cum != "741.86" {
		t.Error("Cumulative TPS is %v, want 741.86", cum)
	}

	if trend[0] != "772.95" {
		t.Error("History IPS[0] is %v, want 772.95", trend[0])
	}
}
