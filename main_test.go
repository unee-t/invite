package main

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func TestMain(m *testing.M) {
	db, _ = sql.Open("mysql", os.Getenv("DSN"))
	defer db.Close()
	os.Exit(m.Run())
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
