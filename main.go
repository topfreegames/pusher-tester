package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/topfreegames/pusher-tester/constants"
	"github.com/topfreegames/pusher-tester/generators"
	producers "github.com/topfreegames/pusher-tester/producers"
	"github.com/topfreegames/pusher/util"
)

var (
	config   *viper.Viper
	logger   *logrus.Logger
	cfgFile  string
	logLevel int

	platform string

	run bool
)

func init() {
	const (
		defaultLogLevel = 4
		logUsage        = "Verbosity level => 0: Panic, " +
			"1: Fatal, " +
			"2: Error, " +
			"3: Warn, " +
			"4: Info, " +
			"5: Debug; " +
			"Default : 4`"

		defaultCfgFile = "./config/default.yaml"
		cfgFileUsage   = "config file with service configuration parameters"
	)

	flag.IntVar(&logLevel, "verbose", defaultLogLevel, logUsage)
	flag.IntVar(&logLevel, "v", defaultLogLevel, logUsage+" (shorthand)")

	flag.StringVar(&cfgFile, "cfgFile", defaultCfgFile, cfgFileUsage)
	flag.StringVar(&cfgFile, "cfg", defaultCfgFile, cfgFileUsage+" (shorthand)")

}

func main() {
	flag.Parse()

	logger := logrus.New()
	logrus.SetLevel(logrus.Level(5))
	logger.Level = logrus.Level(5)

	logger.Info("starting app")
	config, err := util.NewViperWithConfigFile(cfgFile)
	if err != nil {
		panic(err)
	}

	config.SetEnvPrefix("tester")
	config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	config.AutomaticEnv()

	game := config.GetString("game")
	platform := config.GetString("platform")
	var generator generators.MessageGenerator

	if platform == constants.APNSPlatform {
		generator = &generators.APSNMessageGenerator{}
	} else if platform == constants.GCMPlatform {
		generator = &generators.GCMMessageGenerator{}
	}

	var wg sync.WaitGroup
	run = true
	prodSize := config.GetInt("producers")
	for i := 0; i < prodSize; i++ {
		producer, err := producers.NewKafkaProducer(config, logger, nil)
		if err != nil {
			panic(fmt.Sprintf("can't start kafka producer: %s", err))
		}

		wg.Add(1)
		go startToProduce(logger, &wg, producer, generator, game)
	}

	waitToClose := make(chan struct{})
	go listenSignalsAndShutDown(waitToClose)

	<-waitToClose
	run = false
	logger.Info("waiting producers to shut down")
	wg.Wait()
	logger.Info("closing application")
}

func startToProduce(
	logger logrus.FieldLogger, wg *sync.WaitGroup,
	producer *producers.KafkaProducer,
	generator generators.MessageGenerator,
	game string,
) {
	for run {
		msg := generator.Generate()
		producer.SendMessage(game, generator.Platform(), msg)
		// time.Sleep(5 * time.Second)
	}

	logger.Info("closing producer")
	wg.Done()
}

func listenSignalsAndShutDown(
	c chan<- struct{},
) {
	sigint := make(chan os.Signal, 1)

	// interrupt signal sent from terminal
	signal.Notify(sigint, os.Interrupt)
	// sigterm signal sent from kubernetes
	signal.Notify(sigint, syscall.SIGTERM)

	<-sigint
	close(c)
}
