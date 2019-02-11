/*
 * Copyright (c) 2017 TFG Co
 * Author: TFG Co <backend@tfgco.com>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of
 * this software and associated documentation files (the "Software"), to deal in
 * the Software without restriction, including without limitation the rights to
 * use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
 * the Software, and to permit persons to whom the Software is furnished to do so,
 * subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
 * FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
 * COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
 * IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
 * CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package producers

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
	raven "github.com/getsentry/raven-go"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/topfreegames/pusher/interfaces"
	"github.com/topfreegames/pusher/util"
)

// KafkaProducerClient interface
type KafkaProducerClient interface {
	Events() chan kafka.Event
	ProduceChannel() chan *kafka.Message
}

// KafkaProducer for producing push feedbacks to a kafka queue
type KafkaProducer struct {
	Brokers   string
	Config    *viper.Viper
	Producer  interfaces.KafkaProducerClient
	BatchSize int
	LingerMs  int
	Logger    *log.Logger
}

// NewKafkaProducer for creating a new KafkaProducer instance
func NewKafkaProducer(config *viper.Viper, logger *log.Logger, clientOrNil ...interfaces.KafkaProducerClient) (*KafkaProducer, error) {
	q := &KafkaProducer{
		Config: config,
		Logger: logger,
	}
	var producer interfaces.KafkaProducerClient
	if len(clientOrNil) == 1 {
		producer = clientOrNil[0]
	}
	err := q.configure(producer)
	return q, err
}

func (q *KafkaProducer) loadConfigurationDefaults() {
	q.Config.SetDefault("kafka.brokers", "localhost:9941")
	q.Config.SetDefault("kafka.linger.ms", 0)
	q.Config.SetDefault("kafka.batch.size", 1048576)
}

func (q *KafkaProducer) configure(producer interfaces.KafkaProducerClient) error {
	q.loadConfigurationDefaults()
	q.Brokers = q.Config.GetString("kafka.brokers")
	q.BatchSize = q.Config.GetInt("kafka.batch.size")
	q.LingerMs = q.Config.GetInt("kafka.linger.ms")
	c := &kafka.ConfigMap{
		"queue.buffering.max.kbytes": q.BatchSize,
		"linger.ms":                  q.LingerMs,
		"bootstrap.servers":          q.Brokers,
	}
	l := q.Logger.WithFields(log.Fields{
		"brokers": q.Brokers,
	})
	l.Debug("configuring kafka producer")

	if producer == nil {
		p, err := kafka.NewProducer(c)
		q.Producer = p
		if err != nil {
			l.WithError(err).Error("error configuring kafka producer client")
			return err
		}
	} else {
		q.Producer = producer
	}
	go q.listenForKafkaResponses()
	l.Info("kafka producer initialized")
	return nil
}

func (q *KafkaProducer) listenForKafkaResponses() {
	l := q.Logger.WithFields(log.Fields{
		"method": "listenForKafkaResponses",
	})
	for e := range q.Producer.Events() {
		switch ev := e.(type) {
		case *kafka.Message:
			m := ev
			if m.TopicPartition.Error != nil {
				raven.CaptureError(m.TopicPartition.Error, map[string]string{
					"version":   util.Version,
					"extension": "kafka-producer",
				})
				l.WithError(m.TopicPartition.Error).Error("error sending feedback to kafka")
			} else {
				l.WithFields(log.Fields{
					"topic":     *m.TopicPartition.Topic,
					"partition": m.TopicPartition.Partition,
					"offset":    m.TopicPartition.Offset,
				}).Debug("delivered feedback to topic")
			}
			break
		default:
			l.WithField("event", ev).Warn("ignored kafka response event")
		}
	}
}

// SendMessage sends a message to the kafka Queue
func (q *KafkaProducer) SendMessage(game string, platform string, message []byte) {
	topic := "push-" + game + "_" + platform + "-massive"
	m := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: kafka.PartitionAny,
		},
		Value: message,
	}

	q.Producer.ProduceChannel() <- m
}
