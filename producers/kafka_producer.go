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
	"time"

	"github.com/Shopify/sarama"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// KafkaProducer for producing push feedbacks to a kafka queue
type KafkaProducer struct {
	Brokers   string
	Config    *viper.Viper
	Producer  sarama.AsyncProducer
	BatchSize int
	LingerMs  int
	Logger    *log.Logger
}

// NewKafkaProducer for creating a new KafkaProducer instance
func NewKafkaProducer(config *viper.Viper, logger *log.Logger) (*KafkaProducer, error) {
	q := &KafkaProducer{
		Config: config,
		Logger: logger,
	}
	err := q.configure()
	return q, err
}

func (q *KafkaProducer) loadConfigurationDefaults() {
	q.Config.SetDefault("kafka.brokers", "localhost:9941")
	q.Config.SetDefault("kafka.linger.ms", 0)
	q.Config.SetDefault("kafka.batch.size", 1048576)
	q.Config.SetDefault("kafka.maxMessageBytes", 100000)
}

func (q *KafkaProducer) configure() error {
	q.loadConfigurationDefaults()
	q.Brokers = q.Config.GetString("kafka.brokers")
	q.BatchSize = q.Config.GetInt("kafka.batch.size")
	q.LingerMs = q.Config.GetInt("kafka.linger.ms")

	kafkaConf := sarama.NewConfig()
	kafkaConf.Producer.Return.Errors = true
	kafkaConf.Producer.Return.Successes = true
	kafkaConf.Producer.MaxMessageBytes = q.Config.GetInt("kafka.maxMessageBytes")
	kafkaConf.Producer.Flush.Bytes = q.Config.GetInt("kafka.batch.size")
	kafkaConf.Producer.Flush.Frequency = time.Duration(q.Config.GetInt("kafka.linger.ms")) * time.Millisecond
	kafkaConf.Producer.RequiredAcks = sarama.WaitForLocal

	l := q.Logger.WithFields(log.Fields{
		"brokers": q.Brokers,
	})
	l.Debug("configuring kafka producer")

	producer, err := sarama.NewAsyncProducer([]string{q.Brokers}, kafkaConf)
	if err != nil {
		return err
	}
	q.Producer = producer

	go q.listenForKafkaSuccesses()
	go q.listenForKafkaFailures()
	l.Info("kafka producer initialized")
	return nil
}

func (q *KafkaProducer) listenForKafkaSuccesses() {
	l := q.Logger.WithFields(log.Fields{
		"method": "listenForKafkaSuccesses",
	})
	for range q.Producer.Successes() {
		l.Info("sent to kafka success")
	}
}

func (q *KafkaProducer) listenForKafkaFailures() {
	l := q.Logger.WithFields(log.Fields{
		"method": "listenForKafkaFailures",
	})
	for e := range q.Producer.Errors() {
		l.WithField("error", e.Err).Error("sent to kafka failure")
	}
}

// SendMessage sends a message to the kafka Queue
func (q *KafkaProducer) SendMessage(game string, platform string, message []byte) {
	topic := "push-" + game + "_" + platform + "-massive"
	m := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(message),
	}

	q.Producer.Input() <- m
}
