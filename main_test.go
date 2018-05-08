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

func Test_handler_getRole(t *testing.T) {
	type fields struct {
		db *sql.DB
	}
	type args struct {
		roleName string
	}
	tests := []struct {
		name             string
		fields           fields
		args             args
		wantId_role_type int
		wantErr          bool
	}{{
		"Check Owner/Landlord is 2",
		fields{db: db},
		args{roleName: "Owner/Landlord"},
		2,
		false,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := handler{
				db: tt.fields.db,
			}
			gotId_role_type, err := h.lookupRoleID(tt.args.roleName)
			if (err != nil) != tt.wantErr {
				t.Errorf("handler.getRole() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotId_role_type != tt.wantId_role_type {
				t.Errorf("handler.getRole() = %v, want %v", gotId_role_type, tt.wantId_role_type)
			}
		})
	}
}

func Test_markInvitesProcessed(t *testing.T) {
	type args struct {
		ids []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{{
		name:    "Bad id",
		args:    args{ids: []string{"noway this should exist"}},
		wantErr: true,
	},
		{
			name:    "No id",
			args:    args{ids: []string{}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := markInvitesProcessed(tt.args.ids); (err != nil) != tt.wantErr {
				t.Errorf("markInvitesProcessed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
