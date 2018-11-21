package invite

import (
	"net/http"
	"os"
	"testing"

	"github.com/appleboy/gofight"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
)

var h handler

func TestMain(m *testing.M) {
	h, _ = New()
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

// Test route for the sake of https://github.com/unee-t/frontend/issues/297#issuecomment-398604129
func TestPushRoute(t *testing.T) {
	r := gofight.New()
	r.POST("/").
		SetJSON(gofight.D{
			"foo": "bar",
		}).
		// turn on the debug mode.
		SetDebug(true).
		Run(h.BasicEngine(), func(r gofight.HTTPResponse, rq gofight.HTTPRequest) {
			assert.Equal(t, http.StatusBadRequest, r.Code)
		})
}

func TestEmptyPayloadPushRoute(t *testing.T) {
	r := gofight.New()
	r.POST("/").
		SetBody("[]").
		// turn on the debug mode.
		SetDebug(true).
		Run(h.BasicEngine(), func(r gofight.HTTPResponse, rq gofight.HTTPRequest) {
			assert.Equal(t, "Empty payload\n", r.Body.String())
			assert.Equal(t, http.StatusBadRequest, r.Code)
		})
}
