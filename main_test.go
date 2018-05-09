package main

import (
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

var h handler

func TestMain(m *testing.M) {
	h, _ = New()
	defer h.db.Close()
	os.Exit(m.Run())
}

func Test_handler_lookupRoleID(t *testing.T) {
	type args struct {
		roleName string
	}
	tests := []struct {
		name           string
		args           args
		wantIDRoleType int
		wantErr        bool
	}{{
		"Check Owner/Landlord is 2",
		args{roleName: "Owner/Landlord"},
		2,
		false,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIDRoleType, err := h.lookupRoleID(tt.args.roleName)
			if (err != nil) != tt.wantErr {
				t.Errorf("handler.lookupRoleID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotIDRoleType != tt.wantIDRoleType {
				t.Errorf("handler.lookupRoleID() = %v, want %v", gotIDRoleType, tt.wantIDRoleType)
			}
		})
	}
}
