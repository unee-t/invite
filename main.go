package main

import (
	"encoding/json"
	"net/http"
	"os"

	jsonhandler "github.com/apex/log/handlers/json"

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

	// db, err := sql.Open("mysql", "root:uniti@/bugzilla")
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

	// stmtIns, err := db.Prepare("INSERT INTO ut_test_foo_bar VALUES( ? )") // ? = placeholder
	// if err != nil {
	// 	return err
	// }
	// defer stmtIns.Close()
	// _, err = stmtIns.Exec(me)
	// Error 1136: Column count doesn't match value count at row 1

	// _, err = db.Exec(`SET @foo_bar_invitation_id = 'blah blah blah';
	// INSERT INTO ` + "`ut_test_foo_bar`(`foobar_invitation_id`)" + `
	// VALUES (@foo_bar_invitation_id);`)
	// Error 1064: You have an error in your SQL syntax; check the manual that corresponds to your MySQL server version for the right syntax to use near 'INSERT INTO `ut_test_foo_bar`(`foobar_invitation_id`)
	// VALUES (@foo_bar_invit' at line 2

	result, err := db.Exec(
		"INSERT INTO ut_test_foo_bar (foobar_invitation_id) VALUES (?)",
		"gopher",
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

	err := set("foobar")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode("done")

}
