package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	var db *sqlx.DB
	if os.Getenv("PRIMARY_REGION") == os.Getenv("FLY_REGION") {
		fmt.Println("init'ing primary")
		var err error
		if db, err = sqlx.Open("sqlite3", "/litefs/foo.db?mode=rwc&_busy_timeout=750&_journal_mode=WAL&_sync=0"); err != nil {
			log.Fatal(err)
		}

		if _, err = db.Exec(`CREATE TABLE IF NOT EXISTS foos (id INTEGER PRIMARY KEY, created_at DATETIME DEFAULT CURRENT_TIMESTAMP)`); err != nil {
			log.Fatal(err)
		}

		go func() {
			for {
				time.Sleep(time.Second)
				if _, err := db.Exec(`INSERT INTO foos DEFAULT VALUES`); err != nil {
					log.Println(err)
				}
			}
		}()

	} else {
		fmt.Println("init'ing replica")
		var err error
		if db, err = sqlx.Open("sqlite3", "/litefs/foo.db?mode=r&_busy_timeout=750&_journal_mode=WAL&_sync=0"); err != nil {
			log.Fatal(err)
		}
	}

	http.DefaultServeMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if region := r.URL.Query().Get("region"); region != "" && region != os.Getenv("FLY_REGION") {
			log.Printf("replaying in %s\n", region)
			w.Header().Set("fly-replay", fmt.Sprintf("region=%s", region))
			w.WriteHeader(http.StatusOK)
			return
		}

		var t time.Time
		if err := db.Get(&t, `SELECT created_at FROM foos ORDER BY id DESC LIMIT 1;`); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if _, err := fmt.Fprintf(w, "%#v", t); err != nil {
			log.Println(err)
		}
	})

	log.Println(http.ListenAndServe(":8080", nil))
}
