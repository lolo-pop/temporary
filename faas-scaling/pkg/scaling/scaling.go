package scaling

import (
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	"github.com/nats-io/nats.go"
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

const (
	batchSize    = 4
	batchTimeout = 500 * time.Millisecond
)

func Handle() {
	nc, err := nats.Connect(natsUrl)
	if err != nil {
		panic(err)
	}
	defer nc.Close()
	log.Printf("request subject: %s", reqSubject)
	sub, err := nc.SubscribeSync(reqSubject)
	if err != nil {
		panic(err)
	}

	msgs := make([]*nats.Msg, 0, batchSize)
	timer := time.NewTimer(batchTimeout)
	for {
		select {
		case <-timer.C:
			if len(msgs) > 0 {
				sendBatch(nc, msgs)
				fmt.Printf("the length of batch is %d", len(msgs))
				msgs = msgs[:0]
				// } else if len(msgs) == 0 {
				// 	fmt.Println("batch is empty!")
			}
			timer.Reset(batchTimeout)
		default:
			msg, err := sub.NextMsg(1 * time.Second)

			// fmt.Printf("NATS message type: %T", msg)
			if err != nil {
				if err == nats.ErrTimeout {
					// no new message, continue
					continue
				}
				panic(err)
			}
			msgs = append(msgs, msg)
			if len(msgs) == batchSize {
				sendBatch(nc, msgs)
				fmt.Printf("the length of batch is %d", batchSize)
				msgs = msgs[:0]
				timer.Stop()
				timer.Reset(batchTimeout)
			} else if !timer.Stop() && len(timer.C) > 0 {
				<-timer.C
			}
			timer.Reset(batchTimeout)
		}
	}
}

func sendBatch(nc *nats.Conn, msgs []*nats.Msg) {
	fmt.Println("Sending batch:", msgs)
	// send batch of messages here
}
func CalculateTimeout() {

}
