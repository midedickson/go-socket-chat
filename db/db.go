package db

import (
	"fmt"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

var DB *sqlx.DB

func Connect() {
	environment := os.Getenv("ENVIRONMENT")
	var (
		db  *sqlx.DB
		err error
	)

	if environment == "development" {
		db, err = sqlx.Connect("sqlite3", ":memory:")
	} else {
		dbConnectionString := "postgresql://postgres:Fidelwole%4027@localhost:5433/valentina?sslmode=disable"
		db, err = sqlx.Open("postgres", dbConnectionString)
		if err != nil {
			log.Fatalf("Failed to connect to the PostgreSQL database: %v", err)
		}
		err = db.Ping()
	}

	if err != nil {
		log.Fatalf("Database connection error: %v", err)
	}

	DB = db
	log.Println("DB connection established successfully!")
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}

func RunQuery(query string, destination any, queryParameters ...any) error {
	if DB == nil {
		return fmt.Errorf("database connection is not initialized")
	}

	rows, err := DB.Queryx(query, queryParameters...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		if err = rows.StructScan(destination); err != nil {
			return err
		}
	}

	return rows.Err()
}

func Setup() {
	if DB == nil {
		log.Fatal("Database connection is not initialized; cannot create tables.")
	}

	// SQL queries to create tables
	postgresUserTableQueryVersion := `
	CREATE TABLE IF NOT EXISTS Users (
		id SERIAL PRIMARY KEY,
		first_name VARCHAR(255),
		last_name VARCHAR(255),
		phone_number VARCHAR(255),
		email VARCHAR(255) UNIQUE,
		gender VARCHAR(50),
		random_name VARCHAR(255),
		matched BOOLEAN
	);`

	createUserTableIfNotExistsQuery := `
	CREATE TABLE IF NOT EXISTS Users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		first_name VARCHAR(255),
		last_name VARCHAR(255),
		phone_number VARCHAR(255),
		email VARCHAR(255) UNIQUE,
		gender VARCHAR(50),
		random_name VARCHAR(255),
		isPaid BOOLEAN,
		matched BOOLEAN
	);`

	createPaymentsTableQuery := `
	CREATE TABLE IF NOT EXISTS Payments (
		id SERIAL PRIMARY KEY,
		user_id INT,
		amount NUMERIC(10, 2),
		status VARCHAR(50), -- e.g., 'PENDING', 'SUCCESS', 'FAILED'
		transaction_reference VARCHAR(255),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES Users(id)
	);`

	createMatchesTableIfNotExistsQuery := `
	CREATE TABLE IF NOT EXISTS Matches (
		user_id INT,
		matched_user_id INT UNIQUE,
		FOREIGN KEY (user_id) REFERENCES Users(id),
		FOREIGN KEY (matched_user_id) REFERENCES Users(id)
	);`

	addIsPaidColumnQuery := `
	ALTER TABLE Users
	ADD COLUMN isPaid BOOLEAN;`

	// Determine environment and execute queries
	environment := os.Getenv("ENVIRONMENT")

	var err error
	if environment == "development" {
		_, err = DB.Exec(createUserTableIfNotExistsQuery)
	} else {
		_, err = DB.Exec(postgresUserTableQueryVersion)
	}
	if err != nil {
		log.Fatalf("Failed to create Users table: %v", err)
	}

	_, err = DB.Exec(createPaymentsTableQuery)
	if err != nil {
		log.Fatalf("Failed to create Payments table: %v", err)
	}

	_, err = DB.Exec(createMatchesTableIfNotExistsQuery)
	if err != nil {
		log.Fatalf("Failed to create Matches table: %v", err)
	}

	_, err = DB.Exec(addIsPaidColumnQuery)
	if err != nil {
		log.Fatalf("Failed to add isPaid column to Users table: %v", err)
	}
	log.Println("Tables created successfully!")
}
