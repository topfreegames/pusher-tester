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
	extensions "github.com/topfreegames/extensions/kafka"
)

// NewKafkaProducer for creating a new KafkaProducer instance
func NewKafkaProducer(config *viper.Viper, logger *log.Logger) (*extensions.SyncProducer, error) {
	kafkaConf := configure(config, logger)
	k, err := extensions.NewSyncProducer(config, logger, kafkaConf)
	if err != nil {
		return nil, err
	}
	logger.Info("kafka producer initialized")
	return k, nil
}

func loadConfigurationDefaults(config *viper.Viper) {
	config.SetDefault("extensions.kafkaproducer.brokers", "localhost:9941")
	config.SetDefault("extensions.kafkaproducer.maxMessageBytes", 3000000)
	config.SetDefault("extensions.kafkaproducer.batch.size", 1)
	config.SetDefault("extensions.kafkaproducer.linger.ms", 0)
}

func configure(config *viper.Viper, logger log.FieldLogger) *sarama.Config {
	loadConfigurationDefaults(config)

	kafkaConf := sarama.NewConfig()
	kafkaConf.Producer.Return.Errors = true
	kafkaConf.Producer.Return.Successes = true
	kafkaConf.Producer.MaxMessageBytes = config.GetInt("extensions.kafkaproducer.maxMessageBytes")
	kafkaConf.Producer.Flush.Bytes = config.GetInt("extensions.kafkaproducer.batch.size")
	kafkaConf.Producer.Flush.Frequency = time.Duration(config.GetInt("extensions.kafkaproducer.linger.ms")) * time.Millisecond
	kafkaConf.Producer.RequiredAcks = sarama.WaitForLocal

	l := logger.WithFields(log.Fields{
		"brokers": config.GetString("extensions.kafkaproducer.brokers"),
	})
	l.Debug("configuring kafka producer")
	return kafkaConf
}
