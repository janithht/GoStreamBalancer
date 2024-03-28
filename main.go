package main

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

func main() {
	var config Config

	data, err := os.ReadFile("config.yaml") // Read the file
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	err = yaml.Unmarshal(data, &config) // Unmarshal the data to strcuts defined above
	if err != nil {
		log.Fatalf("error parsing config file: %v", err)
	}
	fmt.Println()
	fmt.Println("Config parsed successfully:", config)

	go performHealthChecks(&config)
	startServer(&config)
}
