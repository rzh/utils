package parser

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"

	. "github.com/visionmedia/go-debug"
)

var debug = Debug("info")

type MongoPerfResult struct {
	Name          string
	Thread        int64
	Result        float64
	CV            string
	Version       string
	ClientVersion string
	GitSHA        string
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

func ProcessMongoPerfResult(file string) map[string]MongoPerfResult {

	re := make(map[string]float64)
	data := make(map[string][]float64)
	total := make(map[string]float64)
	thread := make(map[string]int64)
	cvs := make(map[string]string)
	name := make(map[string]string)
	sha := "" // db_version, git_sha
	db_version := ""
	client_version := ""

	f, err := os.Open(file)
	if err != nil {
		log.Fatalln(err)
	}

	scan := bufio.NewScanner(f)

	tmp_id := ""

	for scan.Scan() {
		debug("scan line :%s", scan.Text())

		// check whether it is db version
		// format:
		//    db version: 2.7.5
		if match, err := regexp.MatchString("^MongoDB shell version: ", scan.Text()); match && err == nil {
			// found db version

			if client_version == "" {
				client_version = scan.Text()
			} else {
				// make db_version is the same
				if client_version != scan.Text() {
					log.Fatalln("DB version is different from log file. Got " + client_version + " and " + scan.Text())
				}
			}

			continue
		}

		// if match, err := regexp.MatchString("^db version: [0-9.prec]+$", scan.Text()); match && err == nil {
		if match, err := regexp.MatchString("^db version: ", scan.Text()); match && err == nil {
			// found db version

			if db_version == "" {
				db_version = scan.Text()
			} else {
				// make db_version is the same
				if db_version != scan.Text() {
					log.Fatalln("DB version is different from log file. Got " + db_version + " and " + scan.Text())
				}
			}

			// get next line, it should be SHA
			if scan.Scan() {
				sha = scan.Text()
			}

			continue
		}

		if tmp_id == "" {
			match, err := regexp.MatchString("^[a-zA-Z.0-9_-]+$", scan.Text())

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

				t, e := strconv.ParseInt(words[0], 10, 64)

				if e != nil {
					log.Fatalln("Error parsing line ", scan.Text(), " with error ", e)
				}
				debug("parsing thread count, which is %3d\n", t)
				th := fmt.Sprintf("%03d", t)
				re[tmp_id+"_TH-"+th] += f
				thread[tmp_id+"_TH-"+th] = t
				name[tmp_id+"_TH-"+th] = tmp_id
				data[tmp_id+"_TH-"+th] = append(data[tmp_id+"_TH-"+th], f)
				total[tmp_id+"_TH-"+th] += 1.0

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

	result := make(map[string]MongoPerfResult)
	for k, v := range total {
		re[k] = re[k] / v
		cvs[k] = cvStr(data[k])

		result[k] = MongoPerfResult{
			GitSHA:        sha,
			Result:        re[k],
			CV:            cvs[k],
			Name:          name[k],
			Version:       db_version,
			ClientVersion: client_version,
			Thread:        thread[k]}
	}

	//return re, cvs, db_version + " | SHA: " + sha
	return result
}
