package Logs

import (
	"os"
	"time"
)

var file_log *os.File

func LogMessage(message string) {
	current_time := time.Now()
	file_log.WriteString(current_time.Format("2006-01-02 15:04:05") + " " + message + "\n")
}

func InitLogs() {
	file_log_string := os.Getenv("LOGS_FILE")
	if file_log_string == "" {
		panic("LOGS_FILE env not found")
	}

	var err error

	file_log, err = os.OpenFile(file_log_string, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}

	LogMessage("Logs initialized")
}

// this should be visible to the main package
func PanicLog(message string) {
	LogMessage(("\n\n\nPANIC: " + message))
	LogMessage(("PANIC: " + message))
	LogMessage(("PANIC: " + message))
	LogMessage(("PANIC: " + message + "\n\n\n"))
}
