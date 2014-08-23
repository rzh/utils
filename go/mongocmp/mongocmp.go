package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"

	. "github.com/visionmedia/go-debug"
)

const usageFooter = `
Each input file should be from mongo-perf:
        
MongoCmp compares old and new for each benchmark.

Output format:
benchmark                 old op/s    new op/s      delta
BenchmarkBinaryTree17  131488202467 112637283111  -14.34%
BenchmarkFannkuch11     61976254131  61972329989   -0.01%
`

var debug = Debug("single")

// parse file
// return:
//   mean of data, coefficient of variation of the data
func parseFile(path string) (map[string]float64, map[string]string) {
	re := make(map[string]float64)
	data := make(map[string][]float64)
	total := make(map[string]float64)
	cvs := make(map[string]string)

	f, err := os.Open(path)
	if err != nil {
		log.Fatalln(err)
	}

	scan := bufio.NewScanner(f)

	tmp_id := ""

	for scan.Scan() {
		debug("scan line :%s", scan.Text())
		if tmp_id == "" {
			match, err := regexp.MatchString("^[a-zA-Z.0-9]+$", scan.Text())

			if match && err == nil {
				// this potential could be an id
				tmp_id = scan.Text()
				debug("Found id: %s", tmp_id)
			} else if err != nil {
				log.Fatalln(err)
			}
		} else {
			// there is already a tmp_id, looking for potential bench results
			// if line if of format
			//      5       4141.928744882501
			// then it is bench results.
			// otherwise, reset tmp_id to ""

			match, err := regexp.MatchString(`^[0-9]+[\s]+[0-9.]+[\s]*`, scan.Text())
			if match && err == nil {
				// found some bench results!!
				words := strings.Fields(scan.Text())
				f, err := strconv.ParseFloat(words[1], 64)

				if err != nil {
					log.Fatalln(err)
				}
				re[tmp_id+"_TH-"+words[0]] += f
				data[tmp_id+"_TH-"+words[0]] = append(data[tmp_id+"_TH-"+words[0]], f)
				total[tmp_id+"_TH-"+words[0]] += 1.0

			} else {
				match, err := regexp.MatchString("^[a-zA-Z.0-9]+$", scan.Text())

				if match && err == nil {
					// this potential could be an id
					tmp_id = scan.Text()
					debug("Found id: %s", tmp_id)
				} else if err != nil {
					log.Fatalln(err)
				} else {
					// ok, reset. false alarm
					debug("False alarm, remove id:%s, line (%s)", tmp_id, scan.Text())
					tmp_id = ""
				}
			}
		}
	}

	for k, v := range total {
		re[k] = re[k] / v
		cvs[k] = cvStr(data[k])
	}
	return re, cvs
}

// Percent formats a Delta as a percent change, ranging from -100% up.
func Percent(f float64) string {
	return fmt.Sprintf("%+.2f%%", 100*f-100)
}

// Percent formats a Delta as a percent change, ranging from -100% up.
func Speedup(f float64) string {
	return fmt.Sprintf("%.2fx", f)

}

// to calculate stddev
func stdDev(numbers []float64, mean float64) float64 {
	total := 0.0
	for _, number := range numbers {
		total += math.Pow(number-mean, 2)
	}
	variance := total / float64(len(numbers)-1)
	return math.Sqrt(variance)
}

// to calculate cv (coefficient of variation) in percentage
func cvStr(data []float64) string {
	var stddev float64 = 0
	var mean float64 = 0.0
	var total float64 = 0.0

	if len(data) == 0 {
		return "Error, no data for cvStr"
	} else if len(data) == 1 {
		return "n/a"
	}

	// find mean first
	for _, n := range data {
		total += n
	}

	mean = total / float64(len(data))

	stddev = stdDev(data, mean)

	return fmt.Sprintf("%.2f%%", 100.0*stddev/mean)
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s old.txt new.txt\n\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprint(os.Stderr, usageFooter)
		os.Exit(2)
	}
	flag.Parse()
	if flag.NArg() != 2 {
		flag.Usage()
	}

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 0, 5, ' ', 0)
	defer w.Flush()

	before, before_cvs := parseFile(flag.Arg(0))
	after, after_cvs := parseFile(flag.Arg(1))

	fmt.Printf("# baseline : %s\n# new results : %s\n", flag.Arg(0), flag.Arg(1))
	// fmt.Fprint(w, "\nbenchmark\tbaseline OP/s\tnew OP/s\tspeedup\n")
	fmt.Fprint(w, "benchmark\told ns/op\tnew ns/op\tdelta\n")

	keys := make([]string, 0, len(after))
	for key := range after {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for i := range keys {
		k := keys[i]
		v := after[keys[i]]
		if _, ok := before[k]; ok {
			// has baseline
			fmt.Fprintf(w, "%s\t%.2f[%s]\t%.2f[%s]\t%s\n", k, before[k], before_cvs[k], v, after_cvs[keys[i]], Percent(v/before[k]))
		} else {
			// no baseline
			fmt.Fprintf(w, "%s\t%s\t%f[%s]\t%s\n", k, "n/a", v, "["+after_cvs[keys[i]]+"]", Percent(0.0))
		}
	}
}
