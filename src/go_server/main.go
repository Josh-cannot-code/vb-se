package main

import (
	"context"
	"database/sql"
	"fmt"
	"go_server/database"
	"go_server/youtube"

	"log"
	"net/http"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	ctx := context.Background()

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Could not load environment vars: " + err.Error())
	}
	sqlDb, err := sql.Open("sqlite3", "../db/vb-se.db")
	defer sqlDb.Close()

	db := database.NewSqLiteConnection(sqlDb)
	//	vid := youtube.GetVideo("REi089fakFI")
	err = youtube.RefreshVideos(ctx, db)
	if err != nil {
		log.Fatal(err)
	}

	//	ctx := context.Background()

	/*

		cfg := mysql.Config{
			User:                 os.Getenv("USERNAME"),
			Passwd:               os.Getenv("PASSWORD"),
			Net:                  "tcp",
			Addr:                 os.Getenv("HOST"),
			DBName:               os.Getenv("DATABASE"),
			AllowNativePasswords: true,
		}

		db, err := sql.Open("mysql", cfg.FormatDSN())
		if err != nil {
			log.Fatal(err)
		}

		pingErr := db.Ping()
		if pingErr != nil {
			log.Fatal(pingErr)
		}
		fmt.Println("DB connected")

		http.Handle("/", http.HandlerFunc(testHandler(db)))
	*/

	fmt.Println("Listening on http://localhost:3001")
	err = http.ListenAndServe(":3001", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func testHandler(db *sql.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT video_id FROM videos")
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		var rowdata []byte
		rows.Next()
		rows.Scan(&rowdata)
		fmt.Println(rowdata)

		_, err = w.Write(rowdata)
		if err != nil {
			log.Fatal("Could not write response")
		}
	})
}
