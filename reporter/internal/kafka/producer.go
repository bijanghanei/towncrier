package kafka

import (
	"encoding/json"
	"log"

	"github.com/IBM/sarama"
)

type Producer struct {
	producer sarama.SyncProducer
}

func NewProducer(broker string) *Producer {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer([]string{broker}, config)
	if err != nil {
		log.Fatalf("failed to start kafka producer : %v", err)
	}

	return &Producer{producer: producer}
}

func (p *Producer) SendMessage(topic string, username string, message any) error {
	msg , err := json.Marshal(message)
	if err != nil {
		return err
	}
	_, _, err = p.producer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Key: sarama.StringEncoder(username),
		Value: sarama.ByteEncoder(msg),
	})
	return err
}