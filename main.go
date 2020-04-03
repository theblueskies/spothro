package main

import (
	"fmt"

	"github.com/spf13/viper"
	"github.com/theblueskies/spothro/rates"
)

func main() {
	viper.BindEnv("PORT")
	viper.BindEnv("SEED_RATE_FILE")
	viper.SetDefault("PORT", "9000")
	viper.SetDefault("SEED_RATE_FILE", "rates/seed_rates.json")

	port := viper.GetString("PORT")
	port = ":" + port
	seedRateFile := viper.GetString("SEED_RATE_FILE")

	api, err := rates.NewAPI(seedRateFile)
	if err != nil {
		panic(err)
	}
	fmt.Println(port)

	router := rates.NewRouter(api)
	router.Run(port)
}
