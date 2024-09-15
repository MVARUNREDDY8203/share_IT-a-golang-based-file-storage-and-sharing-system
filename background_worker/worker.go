package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"sendit/db" // Update this import path based on your actual module path

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Retrieve configuration from environment variables
	cleanupIntervalStr := os.Getenv("CLEANUP_INTERVAL")
	expiryDurationStr := os.Getenv("EXPIRY_DURATION")

	cleanupInterval, err := time.ParseDuration(cleanupIntervalStr)
	if err != nil {
		log.Fatalf("Error parsing CLEANUP_INTERVAL: %v", err)
	}

	expiryDuration, err := time.ParseDuration(expiryDurationStr)
	if err != nil {
		log.Fatalf("Error parsing EXPIRY_DURATION: %v", err)
	}

	// Initialize the database connection
	db.ConnectDB()
	defer db.DB.Close()

	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		if err := cleanupOldFiles(expiryDuration); err != nil {
			log.Printf("Error during cleanup: %v", err)
		}
	}
}

func cleanupOldFiles(expiryDuration time.Duration) error {
	// Define the cutoff time for file deletion
	cutoffTime := time.Now().Add(-expiryDuration)

	// Delete old file metadata from the database
	_, err := db.DB.Exec("DELETE FROM files WHERE created_at < ?", cutoffTime)
	if err != nil {
		return fmt.Errorf("error deleting old file metadata from database: %w", err)
	}

	// Optionally, delete files from storage
	err = deleteOldFilesFromStorage(cutoffTime)
	if err != nil {
		return fmt.Errorf("error deleting old files from storage: %w", err)
	}

	log.Println("Old files cleaned up successfully")
	return nil
}

func deleteOldFilesFromStorage(cutoffTime time.Time) error {
	rows, err := db.DB.Query("SELECT file_path FROM files WHERE created_at < ?", cutoffTime)
	if err != nil {
		return fmt.Errorf("error querying old file paths: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var filePath string
		if err := rows.Scan(&filePath); err != nil {
			return fmt.Errorf("error scanning file path: %w", err)
		}

		// Delete the file from the file system
		if err := os.Remove(filePath); err != nil {
			log.Printf("error deleting file %s: %v", filePath, err)
		}
	}

	return nil
}
