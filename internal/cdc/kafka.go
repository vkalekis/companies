package cdc

import (
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/vkalekis/companies/api"
	"github.com/vkalekis/companies/internal/config"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type KafkaCDCOperator struct {
	producer *kafka.Producer
	topic    string
}

func NewKafkaCDCOperator(config *config.Config) (*KafkaCDCOperator, error) {
	operator := KafkaCDCOperator{}

	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": config.Kafka.BootstrapServers,
	})
	if err != nil {
		return nil, err
	}

	operator.producer = producer
	operator.topic = config.Kafka.Topic
	return &operator, nil
}

func (operator *KafkaCDCOperator) LogCDCOperation(op Operation) {

	var id uuid.UUID
	var msg *api.CDC
	var before, after *api.Company

	if op.Before != nil {
		id = op.Before.Id
		before = &api.Company{
			Id:          op.Before.Id.String(),
			Name:        op.Before.Name,
			Description: op.Before.Description,
			Employees:   int32(op.Before.Employees),
			Registered:  op.Before.Registered,
			CompanyType: string(op.Before.CompanyType),
			CreatedAt:   timestamppb.New(op.Before.CreatedAt),
			UpdatedAt:   timestamppb.New(op.Before.UpdatedAt),
		}
	}
	if op.After != nil {
		id = op.After.Id
		after = &api.Company{
			Id:          op.After.Id.String(),
			Name:        op.After.Name,
			Description: op.After.Description,
			Employees:   int32(op.After.Employees),
			Registered:  op.After.Registered,
			CompanyType: string(op.After.CompanyType),
			CreatedAt:   timestamppb.New(op.After.CreatedAt),
			UpdatedAt:   timestamppb.New(op.After.UpdatedAt),
		}
	}

	key := fmt.Sprintf("%s/%s", op.Op, id)

	protoop := func(op Op) api.Operation {
		switch op {
		case Op_Create:
			return api.Operation_OP_CREATE
		case Op_Update:
			return api.Operation_OP_UPDATE
		case Op_Delete:
			return api.Operation_OP_DELETE
		default:

		}
		return api.Operation_OP_UNKNOWN
	}(op.Op)

	msg = &api.CDC{
		Before:    before,
		After:     after,
		Op:        protoop,
		Timestamp: timestamppb.Now(),
	}
	value, err := proto.Marshal(msg)
	if err != nil {
		log.Printf("Error on marshal: %v", err)
		return
	}

	deliveryChan := make(chan kafka.Event)

	err = operator.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &operator.topic,
			Partition: 0}, // Topic with a single partition
		Key:   []byte(key),
		Value: []byte(value),
	}, deliveryChan)
	if err != nil {
		log.Printf("Error on produce: %v", err)
		return
	}

	e := <-deliveryChan
	m := e.(*kafka.Message)

	if m.TopicPartition.Error != nil {
		log.Printf("Delivery failed for op %s: %v", op, m.TopicPartition.Error)
	} else {
		log.Printf("Delivered message for op %s to %v", op, m.TopicPartition)
	}
}
