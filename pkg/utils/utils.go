package utils

import (
    "log"
    "os"
)

// CheckError logs the error message and exits the program if the error is not nil.
func CheckError(err error) {
    if err != nil {
        log.Fatalf("Error: %v", err)
    }
}

// LogInfo logs informational messages to the console.
func LogInfo(message string) {
    log.Println("INFO:", message)
}

// LogWarning logs warning messages to the console.
func LogWarning(message string) {
    log.Println("WARNING:", message)
}

// FormatBytes converts bytes to a human-readable format.
func FormatBytes(bytes uint64) string {
    const (
        KB = 1024
        MB = KB * 1024
        GB = MB * 1024
    )
    switch {
    case bytes >= GB:
        return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
    case bytes >= MB:
        return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
    case bytes >= KB:
        return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
    default:
        return fmt.Sprintf("%d Bytes", bytes)
    }
}