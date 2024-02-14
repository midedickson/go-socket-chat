package db

import (
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

var DB *sqlx.DB

func Connect() error {
	dbConnectionString := os.Getenv("DATABASE_URL")
	environment := os.Getenv("ENVIRONMENT")
	if environment != "development" {
		db, err := sqlx.Connect("postgres", dbConnectionString)
		if err != nil {
			log.Println("Error connecting to database: " + err.Error())
			return err
		}
		DB = db
	} else {
		db, err := sqlx.Connect("sqlite3", ":memory:")
		if err != nil {
			log.Println("Error connecting to database: " + err.Error())
			return err
		}
		DB = db
	}

	log.Println("DB connection established sucessfully!")

	return nil
}

func Close() {
	DB.Close()
}

func RunQuery(query string, destination any, queryParameters ...any) {

	rows, err := DB.Queryx(query, queryParameters)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {

		err = rows.StructScan(&destination)
		if err != nil {
			log.Fatal(err)
		}

	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
}

func Setup() {
	createUserTableIfNotExistsQuery := `
		CREATE TABLE IF NOT EXISTS Users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			first_name VARCHAR(255),
			last_name VARCHAR(255),
			phone_number VARCHAR(255),
			email VARCHAR(255) UNIQUE,
			gender VARCHAR(50),
			random_name VARCHAR(255),
			matched BOOLEAN
		);`

	createMatchesTableIfNotExistsQuery := `
		CREATE TABLE IF NOT EXISTS Matches (
			user_id INT,
			matched_user_id INT UNIQUE,
			FOREIGN KEY (user_id) REFERENCES Users(ID),
			FOREIGN KEY (matched_user_id) REFERENCES Users(ID)
		);`
	// Execute the query to create the Users table if it doesn't exist
	_, err := DB.Exec(createUserTableIfNotExistsQuery)
	if err != nil {
		log.Fatalf("Failed to create Users table: %v", err)
	}

	// Execute the query to create the Matches table if it doesn't exist
	_, err = DB.Exec(createMatchesTableIfNotExistsQuery)
	if err != nil {
		log.Fatalf("Failed to create Matches table: %v", err)
	}

	log.Println("Tables created successfully!")

}
