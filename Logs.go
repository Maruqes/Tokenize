package Tokenize

import (
	"os"
	"time"
)

var file_log *os.File

func logMessage(message string) {
	current_time := time.Now()
	file_log.WriteString(current_time.Format("2006-01-02 15:04:05") + " " + message + "\n")
}

func initLogs() {
	file_log_string := os.Getenv("LOGS_FILE")
	if file_log_string == "" {
		panic("LOGS_FILE env not found")
	}

}
