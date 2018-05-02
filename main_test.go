package main

import (
	"database/sql"
	"os"
	"testing"

	"github.com/apex/log"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func TestMain(m *testing.M) {

	db, err := sql.Open("mysql", os.Getenv("DSN"))
	if err != nil {
		log.WithError(err).Fatal("error opening database")
	}

	defer db.Close()
	log.Info("starting test")
	log.Infof("here testing with %v", db)
	code := m.Run()
	log.Info("finished test")
	os.Exit(code)

}

func Test_getRole(t *testing.T) {
	type args struct {
		db       *sql.DB
		roleName string
	}
	tests := []struct {
		name             string
		args             args
		wantId_role_type int
		wantErr          bool
	}{{
		"Check Owner/Landlord is 2",
		// db here is nil, why?
		args{db: db, roleName: "Owner/Landlord"},
		2,
		false,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotId_role_type, err := getRole(tt.args.db, tt.args.roleName)
			if (err != nil) != tt.wantErr {
				t.Errorf("getRole() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotId_role_type != tt.wantId_role_type {
				t.Errorf("getRole() = %v, want %v", gotId_role_type, tt.wantId_role_type)
			}
		})
	}
}
