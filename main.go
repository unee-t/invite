package main

import (
	"net/http"
	"os"

	jsonhandler "github.com/apex/log/handlers/json"
	"github.com/tj/go/http/response"

	"github.com/apex/log"
	"github.com/apex/log/handlers/text"
	"github.com/gorilla/pat"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

func init() {
	if os.Getenv("UP_STAGE") == "" {
		log.SetHandler(text.Default)
	} else {
		log.SetHandler(jsonhandler.Default)
	}
}

func main() {
	addr := ":" + os.Getenv("PORT")
	app := pat.New()
	app.Get("/", listjson)
	if err := http.ListenAndServe(addr, app); err != nil {
		log.WithError(err).Fatal("error listening")
	}

}

func set(me string) error {

	db, err := sql.Open("mysql", os.Getenv("DSN"))
	if err != nil {
		return err
	}

	defer db.Close()

	// Open doesn't open a connection. Validate DSN data:
	err = db.Ping()
	if err != nil {
		log.WithError(err).Error("failed to open database")
		return err
	}

	result, err := db.Exec(
		"INSERT INTO ut_test_foo_bar (foobar_invitation_id) VALUES (?)",
		me,
	)
	if err != nil {
		return err
	}

	log.Infof("%v", result)

	var (
		id   int
		name string
	)

	rows, err := db.Query("select foobar_id, foobar_invitation_id from ut_test_foo_bar")
	if err != nil {
		log.WithError(err).Error("failed to open database")
		return err
	}

	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&id, &name)

		if err != nil {
			log.WithError(err).Error("failed to scan")
			return err
		}
		log.Infof("%d %s", id, name)
	}

	err = rows.Err()
	return err

}

func listjson(w http.ResponseWriter, r *http.Request) {

	if os.Getenv("UP_STAGE") != "production" {
		w.Header().Set("X-Robots-Tag", "none")
	}

	value := r.URL.Query().Get("v")
	if value == "" {
		response.BadRequest(w, "Input parameter v is empty.")
		return
	}

	err := set(value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response.OK(w, "Worked")

}
