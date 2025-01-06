package towncrierconsumer

import (
	"fmt"

	"github.com/IBM/sarama"
)

type TownCrierConsumer struct {
	Consumer sarama.Consumer
	Topic string
	Brokers []string
}

func CreateConsumer(brokers []string, topic string) (*TownCrierConsumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	consumer, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		return nil, err
	}

	return &TownCrierConsumer{
		Consumer: consumer,
		Topic: topic,
		}, nil	
}

func (tcc *TownCrierConsumer) GetPartitions() ([]int32, error) {
	partitions, err := tcc.Consumer.Partitions(tcc.Topic)
	if err != nil {
		return nil, fmt.Errorf("failed to get partitions for topic %s: %v", tcc.Topic, err)
	}
	return partitions, nil
}