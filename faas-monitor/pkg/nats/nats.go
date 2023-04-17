package nats

import (
	"fmt"
	"github.com/nats-io/nats.go"
	"log"
	"os"
)

var (
	natsUrl string
	subject string
)

func init() {
	var ok bool
	natsUrl, ok = os.LookupEnv("NATS_URL")
	if !ok {
		log.Fatal("$NATS_URL not set")
	}
	subject, ok = os.LookupEnv("NATS_SUBJECT")
	if !ok {
		log.Fatal("$NATS_SUBJECT not set")
	}
}

func Publish(msg []byte) {

	nc, err := nats.Connect(natsUrl)
	if err != nil {
		errMsg := fmt.Sprintf("Cannot connect to nats: %s", err)
		log.Fatal(errMsg)
	}
	defer nc.Close()

	log.Printf("Publishing %d bytes to: %q", len(msg), subject)

	err = nc.Publish(subject, msg)
	if err != nil {
		log.Fatal(err.Error())
	}
}
