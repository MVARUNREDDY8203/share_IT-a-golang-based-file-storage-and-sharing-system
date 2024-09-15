package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

var DB *sql.DB

// ConnectDB initializes the MySQL database connection
func ConnectDB() {
    var err error
    err = godotenv.Load()
    if err != nil {
        log.Println("No .env file found, continuing...")
    }

    
    dbUser := os.Getenv("DB_USER")
    dbPassword := os.Getenv("DB_PASSWORD")
    dbName := os.Getenv("DB_NAME")
    dbHost := os.Getenv("DB_HOST")
    
    dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s", dbUser, dbPassword, dbHost, dbName)
    
    DB, err = sql.Open("mysql", dsn)
    if err != nil {
        log.Fatal("Could not connect to MySQL database: ", err)
    }

    if err = DB.Ping(); err != nil {
        log.Fatal("Could not ping the MySQL database re: ", err)
    }
    log.Println("Connected to MySQL database successfully!")
}
