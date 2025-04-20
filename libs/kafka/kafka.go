package kafka

import (
	"log"
	"time"

	"github.com/IBM/sarama"
)

// Config содержит конфигурацию для клиента Kafka
type Config struct {
	Brokers  []string
	ClientID string
	Topic    string
}

// Producer представляет Kafka Producer
type Producer struct {
	producer sarama.AsyncProducer
	topic    string
}

// Consumer представляет Kafka Consumer
type Consumer struct {
	consumer sarama.Consumer
	topic    string
}

// NewProducer создает новый асинхронный Producer
func NewProducer(cfg Config) (*Producer, error) {
	config := sarama.NewConfig()
	config.ClientID = cfg.ClientID
	config.Producer.RequiredAcks = sarama.WaitForLocal
	config.Producer.Compression = sarama.CompressionSnappy
	config.Producer.Flush.Frequency = 500 * time.Millisecond
	config.Producer.Return.Successes = true

	producer, err := sarama.NewAsyncProducer(cfg.Brokers, config)
	if err != nil {
		return nil, err
	}

	// Запуск обработки ошибок
	go func() {
		for err := range producer.Errors() {
			log.Printf("Failed to send message to Kafka: %s", err.Error())
		}
	}()

	return &Producer{
		producer: producer,
		topic:    cfg.Topic,
	}, nil
}

// Send отправляет сообщение в Kafka асинхронно
func (p *Producer) Send(key string, value []byte) {
	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(value),
	}
	p.producer.Input() <- msg
}

// Close закрывает Producer
func (p *Producer) Close() error {
	return p.producer.Close()
}

// MessageHandler обрабатывает полученные сообщения
type MessageHandler func(key string, value []byte) error

// NewConsumer создает новый Consumer
func NewConsumer(cfg Config) (*Consumer, error) {
	config := sarama.NewConfig()
	config.ClientID = cfg.ClientID
	config.Consumer.Return.Errors = true

	consumer, err := sarama.NewConsumer(cfg.Brokers, config)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		consumer: consumer,
		topic:    cfg.Topic,
	}, nil
}

// Subscribe подписывается на сообщения из топика
func (c *Consumer) Subscribe(handler MessageHandler) error {
	partConsumer, err := c.consumer.ConsumePartition(c.topic, 0, sarama.OffsetNewest)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case msg := <-partConsumer.Messages():
				err := handler(string(msg.Key), msg.Value)
				if err != nil {
					log.Printf("Error handling message: %s", err.Error())
				}
			case err := <-partConsumer.Errors():
				log.Printf("Error consuming from Kafka: %s", err.Error())
			}
		}
	}()

	return nil
}

// Close закрывает Consumer
func (c *Consumer) Close() error {
	return c.consumer.Close()
}
