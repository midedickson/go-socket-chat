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
	if environment == "development" {
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
