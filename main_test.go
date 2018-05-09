package main

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/apex/log"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func TestMain(m *testing.M) {
	cfg, err := external.LoadDefaultAWSConfig(external.WithSharedConfigProfile("uneet-dev"))
	if err != nil {
		log.WithError(err).Error("failed to load config")
		return
	}
	ssm := ssm.New(cfg)

	db, _ = sql.Open("mysql", fmt.Sprintf("bugzilla:%s@tcp(%s:3306)/bugzilla?multiStatements=true&sql_mode=TRADITIONAL",
		getSecret(ssm, "MYSQL_PASSWORD"),
		udomain("auroradb", getSecret(ssm, "STAGE"))))
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
