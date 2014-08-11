package main

import (
	"io/ioutil"
	"log"
	"regexp"
	"strings"
)

func ProcessSysbenchResult(file string) (string, []string) {
	var cum string
	var trend []string

	f, err := ioutil.ReadFile(file)

	if err != nil {
		log.Fatal("cannot opern file " + "test.txt")
	}

	lines := strings.Split(string(f), "\n")

	re := regexp.MustCompile("seconds : cum tps=([0-9.]+) : int tps=([0-9.]+) : cum ips=[0-9.]+ : int ips=[0-9.]+")

	// find the cumulative number
	for i := len(lines) - 1; i >= 0; i-- {
		t := re.FindStringSubmatch(lines[i])
		if len(t) > 0 {
			cum = t[1]
			break
		} else {
			// no match, just skip
		}
	}

	// get the historical interval number
	for i := 0; i < len(lines); i++ {
		t := re.FindStringSubmatch(lines[i])
		if len(t) > 0 {
			trend = append(trend, t[2])
		} else {
			// no match, just skip
		}
	}
	return cum, trend
}
