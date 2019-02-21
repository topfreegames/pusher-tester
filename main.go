package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/topfreegames/pusher-tester/constants"
	"github.com/topfreegames/pusher-tester/generators"
	"github.com/topfreegames/pusher-tester/producers"
	"github.com/topfreegames/pusher/util"

	"net/http/pprof"
	_ "net/http/pprof"
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

	games := config.GetStringSlice("games")
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
	producersA := make([]*producers.KafkaProducer, 0, prodSize*len(games))
	for _, game := range games {
		for i := 0; i < prodSize; i++ {
			producer, err := producers.NewKafkaProducer(config, logger)
			if err != nil {
				panic(fmt.Sprintf("can't start kafka producer: %s", err))
			}
			producersA = append(producersA, producer)
			wg.Add(1)
			go startToProduce(logger, &wg, producer, generator, game)
		}
	}

	// HTTP Server
	waitToClose := make(chan struct{})
	h := http.NewServeMux()
	h.HandleFunc("/healthcheck", healthcheck)
	h.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
	h.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
	h.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	h.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	h.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))

	addr := fmt.Sprintf("%s:%d",
		config.GetString("server.address.host"),
		config.GetInt("server.address.port"))

	svr := http.Server{
		Addr:    addr,
		Handler: h,
	}

	go listenSignalsAndShutDown(&svr, logger, waitToClose)

	if err := svr.ListenAndServe(); err != http.ErrServerClosed {
		logger.WithError(err).Error("Error in ListenAndServe")
	}

	run = false
	logger.Info("waiting producers to shut down")
	wg.Wait()
	for _, p := range producersA {
		p.Producer.AsyncClose()
	}
	logger.Info("closing application")
}

func healthcheck(w http.ResponseWriter, r *http.Request) {
	message := "Working"
	w.Write([]byte(message))
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
		// time.Sleep(10 * time.Second)
	}

	logger.Info("closing producer")
	wg.Done()
}

func listenSignalsAndShutDown(
	srv *http.Server, logger logrus.FieldLogger,
	c chan<- struct{},
) {
	sigint := make(chan os.Signal, 1)

	// interrupt signal sent from terminal
	signal.Notify(sigint, os.Interrupt)
	// sigterm signal sent from kubernetes
	signal.Notify(sigint, syscall.SIGTERM)

	<-sigint

	if err := srv.Shutdown(context.Background()); err != nil {
		// Error from closing listeners, or context timeout:
		logger.WithField("err", err).
			Error("HTTP server Shutdown")
	}
	close(c)
}
