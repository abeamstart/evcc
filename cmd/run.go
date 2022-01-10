package cmd

import (
	"fmt"
	"net/http"
	_ "net/http/pprof" // pprof handler
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/evcc-io/evcc/server"
	"github.com/evcc-io/evcc/server/updater"
	"github.com/evcc-io/evcc/util"
	"github.com/evcc-io/evcc/util/pipe"
	"github.com/evcc-io/evcc/util/sponsor"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	ignoreErrors = []string{"warn", "error", "fatal"} // don't add to cache
	ignoreMqtt   = []string{"releaseNotes"}           // excessive size may crash certain brokers
)

// runCmd represents the base command when called without any subcommands
var runCmd = &cobra.Command{
	Use:              "run",
	Hidden:           true,
	Version:          fmt.Sprintf("%s (%s)", server.Version, server.Commit),
	PersistentPreRun: persistentConfig,
	PreRun:           runConfig,
	Run:              runRun,
}

func init() {
	rootCmd.AddCommand(runCmd)
}

func runConfig(cmd *cobra.Command, args []string) {
	cmd.PersistentFlags().StringP(
		"uri", "u",
		"0.0.0.0:7070",
		"Listen address",
	)
	bind(cmd, "uri")

	cmd.PersistentFlags().DurationP(
		"interval", "i",
		10*time.Second,
		"Update interval",
	)
	bind(cmd, "interval")

	cmd.PersistentFlags().Bool(
		"metrics",
		false,
		"Expose metrics",
	)
	bind(cmd, "metrics")

	cmd.PersistentFlags().Bool(
		"profile",
		false,
		"Expose pprof profiles",
	)
	bind(cmd, "profile")
}

func runRun(cmd *cobra.Command, args []string) {
	util.LogLevel(viper.GetString("log"), viper.GetStringMapString("levels"))
	log.INFO.Printf("evcc %s (%s)", server.Version, server.Commit)

	// load config and re-configure logging after reading config file
	conf, err := loadConfigFile(cfgFile)
	if err != nil {
		log.ERROR.Println("missing evcc config - switching into demo mode")
		conf = demoConfig()
	}

	util.LogLevel(viper.GetString("log"), viper.GetStringMapString("levels"))

	uri := viper.GetString("uri")
	log.INFO.Println("listening at", uri)

	// setup environment
	if err := configureEnvironment(conf); err != nil {
		log.FATAL.Fatal(err)
	}

	// setup loadpoints
	cp.TrackVisitors() // track duplicate usage

	site, err := configureSiteAndLoadpoints(conf)
	if err != nil {
		log.FATAL.Fatal(err)
	}

	// start broadcasting values
	tee := &util.Tee{}

	// value cache
	cache := util.NewCache()
	go cache.Run(pipe.NewDropper(ignoreErrors...).Pipe(tee.Attach()))

	// setup database
	if conf.Influx.URL != "" {
		configureDatabase(conf.Influx, site.LoadPoints(), tee.Attach())
	}

	// setup mqtt publisher
	if conf.Mqtt.Broker != "" {
		publisher := server.NewMQTT(conf.Mqtt.RootTopic())
		go publisher.Run(site, pipe.NewDropper(ignoreMqtt...).Pipe(tee.Attach()))
	}

	// create webserver
	socketHub := server.NewSocketHub()
	httpd := server.NewHTTPd(uri, site, socketHub, cache)

	// metrics
	if viper.GetBool("metrics") {
		httpd.Router().Handle("/metrics", promhttp.Handler())
	}

	// pprof
	if viper.GetBool("profile") {
		httpd.Router().PathPrefix("/debug/").Handler(http.DefaultServeMux)
	}

	// start HEMS server
	if conf.HEMS.Type != "" {
		hems := configureHEMS(conf.HEMS, site, httpd)
		go hems.Run()
	}

	// publish to UI
	go socketHub.Run(tee.Attach(), cache)

	// setup values channel
	valueChan := make(chan util.Param)
	go tee.Run(valueChan)

	// expose sponsor to UI
	if sponsor.Subject != "" {
		valueChan <- util.Param{Key: "sponsor", Val: sponsor.Subject}
	}

	// version check
	go updater.Run(log, httpd, tee, valueChan)

	// capture log messages for UI
	util.CaptureLogs(valueChan)

	// setup messaging
	pushChan := configureMessengers(conf.Messaging, cache)

	// set channels
	site.Prepare(valueChan, pushChan)
	site.DumpConfig()

	stopC := make(chan struct{})
	exitC := make(chan struct{})

	go func() {
		site.Run(stopC, conf.Interval)
		close(exitC)
	}()

	// uds health check listener
	go server.HealthListener(site, exitC)

	// catch signals
	go func() {
		signalC := make(chan os.Signal, 1)
		signal.Notify(signalC, os.Interrupt, syscall.SIGTERM)

		<-signalC    // wait for signal
		close(stopC) // signal loop to end

		select {
		case <-exitC: // wait for loop to end
		case <-time.NewTimer(conf.Interval).C: // wait max 1 period
		}

		os.Exit(1)
	}()

	log.FATAL.Println(httpd.ListenAndServe())
}
