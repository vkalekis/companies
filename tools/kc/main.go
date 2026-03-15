package main

import (
	"flag"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/vkalekis/companies/api"
	"google.golang.org/protobuf/proto"
)

func genRandomGroupId() string {

	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	groupId := []rune("kc-tool-")
	for range 10 {
		groupId = append(groupId, letters[rand.Intn(len(letters))])
	}
	return string(groupId)
}

func main() {

	log := log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)

	bootstrapServers := flag.String("kafka", "", "kafka bootstrap servers")
	topic := flag.String("topic", "", "kafka topic")
	flag.Parse()

	groupId := genRandomGroupId()

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": *bootstrapServers,
		"group.id":          groupId,
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	log.Printf("Starting consumer in topic %s with consumer group id %s", *topic, groupId)

	osSigCh := make(chan os.Signal, 1)
	signal.Notify(osSigCh, syscall.SIGINT, syscall.SIGTERM)

	rebalanceCB := func(c *kafka.Consumer, ev kafka.Event) error {
		switch e := ev.(type) {
		case kafka.AssignedPartitions:
			log.Printf("Assigned %v partitions", e.Partitions)
		case kafka.RevokedPartitions:
			log.Printf("Revoked %v partitions", e.Partitions)
		}
		return nil
	}
	c.SubscribeTopics([]string{*topic}, rebalanceCB)

	for {
		select {
		case <-osSigCh:
			log.Print("Closing")
			return
		default:
			msg, err := c.ReadMessage(100 * time.Millisecond)
			if err != nil {
				continue
			}

			var val api.CDC
			if err := proto.Unmarshal(msg.Value, &val); err != nil {
				log.Printf("Failed to unmarshal: %v", err)
				continue
			}

			log.Printf("Message at %d@%d : key=%s",
				msg.TopicPartition.Offset, msg.TopicPartition.Partition, string(msg.Key))
			log.Printf("\t op=%s, before=%+v, after=%+v, ts=%s",
				val.Op.String(), val.Before, val.After, val.Timestamp.AsTime().Format(time.RFC3339))
		}
	}
}
