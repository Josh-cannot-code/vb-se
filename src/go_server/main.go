package main

import (
	//"context"
	"database/sql"
	"fmt"

	//"io"
	"log"
	"net/http"

	//"os"

	//"github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	err := godotenv.Load(".env")
	//ctx := context.Background()
	fmt.Println(getVideo("this"))
	//getVideoIds(ctx, "UCGaVdbSav8xWuFWTadK6loA")

	/*
		if err != nil {
			log.Fatal("Could not load environment vars: " + err.Error())
		}

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

		// Test youtube API
		ytApiKey := os.Getenv("YOUTUBE_API_KEY")
		resp, err := http.Get("https://www.googleapis.com/youtube/v3/videos?id=7lCDEYXw3mM&key=" + ytApiKey + "&part=snippet,contentDetails,statistics,status")
		if err != nil {
			log.Fatal(err)
		}

		body, err := io.ReadAll(resp.Body)
		fmt.Println(string(body))

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
