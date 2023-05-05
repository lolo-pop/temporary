package scaling

import (
	"fmt"
	"log"
	"os"
	"sort"
)

type Kv struct {
	Key   string
	Value float32
}

var (
	natsUrl        string
	metricsSubject string
	reqSubject     string
)

func init() {
	var ok bool
	natsUrl, ok = os.LookupEnv("NATS_URL")
	if !ok {
		log.Fatal("$NATS_URL not set")
	}
	metricsSubject, ok = os.LookupEnv("METRICS_SUBJECT")
	if !ok {
		log.Fatal("$METRICS_SUBJECT not set")
	}
	reqSubject, ok = os.LookupEnv("REQ_SUBJECT")
	if !ok {
		log.Fatal("$REQ_SUBJECT not set")
	}

}

func Hello(name string) {
	fmt.Println("hello scaling:", name)
}

func FunctionAccuracyMapSort(acc map[string]float32) []Kv {

	var result []Kv
	for k, v := range acc {
		result = append(result, Kv{k, v})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Value > result[j].Value
	})
	// for _, kvpair := range result {
	//	fmt.Println(kvpair)
	// }
	return result
}

func getLevel(acc float64, SCMap map[int][]int) (int, bool) {
	for level, limits := range SCMap {
		if acc < float64(limits[1]) && acc >= float64(limits[0]) {
			return level, true
		}
	}
	return -1, false
}
func ServiceContainerSLO(SCMap map[int][]int, functionAccuracy map[string]float64, functionLatency map[string]float64) map[int][]float64 {
	levelnum := len(SCMap)
	minlatency := make([]float64, levelnum)
	for i := 0; i < levelnum; i++ {
		minlatency[i] = 0
	}
	for fname, acc := range functionAccuracy {
		lat := functionLatency[fname]
		level, err := getLevel(acc, SCMap)
		if !err {
			fmt.Printf("get function %s accuracy level failed", fname)
		}
		if minlatency[level] > lat {
			minlatency[level] = lat
		}
	}
	SCSLO := make(map[int][]float64)
	for i := 0; i < levelnum; i++ {
		SCSLO[i] = append(SCSLO[i], float64(SCMap[i][0]), float64(SCMap[i][1]), float64(minlatency[i]))
	}
	return SCSLO
}
