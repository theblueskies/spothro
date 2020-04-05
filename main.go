package main

import (
	"log"
	"net/http"

	"github.com/spf13/viper"
	"github.com/theblueskies/spothro/rates"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// PORT is used to decide which port the service will run on
	// The default is set to 9000
	viper.BindEnv("PORT")
	viper.SetDefault("PORT", "9000")
	port := viper.GetString("PORT")
	port = ":" + port
	// SEED_RATE_FILE is used to decide which seed rate file to load initial rates from
	// The default is set to "rates/seed_rates.json"
	viper.BindEnv("SEED_RATE_FILE")
	viper.SetDefault("SEED_RATE_FILE", "rates/seed_rates.json")
	seedRateFile := viper.GetString("SEED_RATE_FILE")

	// Get an instance of the API
	api, err := rates.NewAPI(seedRateFile)
	if err != nil {
		panic(err)
	}
	log.Println(port)

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":2112", nil)
	}()
	// Get an instance of the router and pass in the API as parameter
	// rates.API implements the rates.Service interface
	router := rates.NewRouter(api)
	// the service is started
	router.Run(port)
}
