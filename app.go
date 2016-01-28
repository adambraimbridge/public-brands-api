package main

import (
	"github.com/Financial-Times/base-ft-rw-app-go"
	"github.com/Financial-Times/go-fthealth/v1a"
	"github.com/Financial-Times/http-handlers-go"
	"github.com/Financial-Times/public-brands-api/brands"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/jawher/mow.cli"
	"github.com/jmcvetta/neoism"
	"github.com/rcrowley/go-metrics"
	"net/http"
	"os"
	"strings"
)

func main() {
	app := cli.App("public-brands-api", "A public RESTful API for accessing Brands in neo4j")
	neoURL := app.StringOpt("neo-url", "http://localhost:7474/db/data", "neo4j endpoint URL")
	port := app.StringOpt("port", "8080", "Port to listen on")
	logLevel := app.StringOpt("log-level", "INFO", "Logging level (DEBUG, INFO, WARN, ERROR)")
	graphiteTCPAddress := app.StringOpt("graphiteTCPAddress", "",
		"Graphite TCP address, e.g. graphite.ft.com:2003. Leave as default if you do NOT want to output to graphite (e.g. if running locally)")
	graphitePrefix := app.StringOpt("graphitePrefix", "",
		"Prefix to use. Should start with content, include the environment, and the host name. e.g. content.test.public.brands.api.ftaps59382-law1a-eu-t")
	logMetrics := app.BoolOpt("logMetrics", false, "Whether to log metrics. Set to true if running locally and you want metrics output")

	app.Action = func() {
		setLogLevel(strings.ToUpper(*logLevel))
		baseftrwapp.OutputMetricsIfRequired(*graphiteTCPAddress, *graphitePrefix, *logMetrics)
		log.Infof("public-brands-api will listen on port: %s, connecting to: %s", *port, *neoURL)
		runServer(*neoURL, *port)
	}
	app.Run(os.Args)
}

func runServer(neoURL string, port string) {
	db, err := neoism.Connect(neoURL)
	if err != nil {
		log.Fatalf("Error connecting to neo4j %s", err)
	}
	brands.BrandsDriver = brands.NewCypherDriver(db)
	r := mux.NewRouter()

	// Healthchecks and standards first
	r.HandleFunc("/__health", v1a.Handler("BrandsReadWriteNeo4j Healthchecks",
		"Checks for accessing neo4j", brands.HealthCheck()))
	r.HandleFunc("/ping", brands.Ping)
	r.HandleFunc("/__ping", brands.Ping)

	// Then API specific ones:
	r.HandleFunc("/brands/{uuid}", brands.GetBrand).Methods("GET")

	if err := http.ListenAndServe(":"+port,
		httphandlers.HTTPMetricsHandler(metrics.DefaultRegistry,
			httphandlers.TransactionAwareRequestLoggingHandler(log.StandardLogger(), r))); err != nil {
		log.Fatalf("Unable to start server: %v", err)
	}
}

func setLogLevel(level string) {
	switch level {
	case "DEBUG":
		log.SetLevel(log.DebugLevel)
	case "INFO":
		log.SetLevel(log.InfoLevel)
	case "WARN":
		log.SetLevel(log.WarnLevel)
	case "ERROR":
		log.SetLevel(log.ErrorLevel)
	default:
		log.Errorf("Requested log level %s is not supported, will default to INFO level", level)
		log.SetLevel(log.InfoLevel)
	}
	log.Debugf("Logging level set to %s", level)
}
