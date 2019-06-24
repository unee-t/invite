package invite

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

type checkFunc func(*httptest.ResponseRecorder) error

func hasStatus(want int) checkFunc {
	return func(rec *httptest.ResponseRecorder) error {
		if rec.Code != want {
			return fmt.Errorf("expected status %d, found %d", want, rec.Code)
		}
		return nil
	}
}

var h handler

func TestMain(m *testing.M) {
	h, _ = New(context.Background())
	defer h.DB.Close()
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

func TestPush(t *testing.T) {

	check := func(fns ...checkFunc) []checkFunc { return fns }

	payload := func(u []Invite) io.Reader {
		requestByte, _ := json.Marshal(u)
		return bytes.NewReader(requestByte)
	}

	tests := [...]struct {
		name    string
		verb    string
		path    string
		payload io.Reader
		h       func(w http.ResponseWriter, r *http.Request)
		checks  []checkFunc
	}{
		{
			"Empty payload",
			"POST",
			"/create",
			payload([]Invite{}),
			h.handlePush,
			check(hasStatus(400)),
		},
	}
	for _, tt := range tests {
		r, err := http.NewRequest(tt.verb, tt.path, tt.payload)
		if err != nil {
			t.Fatal(err)
		}
		t.Run(tt.name, func(t *testing.T) {
			h := http.HandlerFunc(tt.h)
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, r)
			for _, check := range tt.checks {
				if err := check(rec); err != nil {
					t.Error(err)
				}
			}
		})
	}
}
