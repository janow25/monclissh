package main

import (
	"flag"
	"log"
	"monclissh/internal/config"
	"monclissh/internal/dashboard"
	"time"
)

func main() {
	// Parse flags
	tickDelay := flag.Duration("t", 2*time.Second, "update interval for server metrics (e.g. 1s, 500ms)")
	debug := flag.Bool("debug", false, "show hosts with errors even if never loaded successfully")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig("configs/hosts.yaml")
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	// Start the dashboard with the tick delay and debug flag
	dashboard.Start(cfg, *tickDelay, *debug)
}
