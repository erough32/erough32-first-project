package main

import "time"

var readFrequencyLimit map[int]map[string]int64 = make(map[int]map[string]int64)

func addReadLimit(bid int, uid string) {
	if readFrequencyLimit[bid] == nil {
		readFrequencyLimit[bid] = make(map[string]int64)
	}
	readFrequencyLimit[bid][uid] = time.Now().Unix()
}

func readLimit(bid int, uid string) bool {
	return readFrequencyLimit[bid][uid] < time.Now().Unix()-60*60*24
}
