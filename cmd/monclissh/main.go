package main

import (
    "log"
    "monclissh/internal/config"
    "monclissh/internal/dashboard"
)

func main() {
    // Load configuration
    cfg, err := config.LoadConfig("configs/hosts.yaml")
    if err != nil {
        log.Fatalf("Error loading configuration: %v", err)
    }

    // Start the dashboard
    dashboard.Start(cfg)
}