package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/apex/log"
	uuid "github.com/satori/go.uuid"
	// This is a hardcoded variable <-- should be moved
	"github.com/unee-t-ins/invite"
	// END This is a hardcoded variable
)

func TestPOST500(t *testing.T) {
	ctx := context.Background()
	h, err := invite.New(ctx)
	if err != nil {
		log.WithError(err).Fatal("error setting configuration")
		return
	}
	// Database should become unavailable and result in 500s
	h.DB.Close()
	app := h.BasicEngine()
	server := httptest.NewServer(app)
	defer server.Close()
	// Delete
	// map[_id:zbHFMYRpSiHmMNzgh caseId:0 invitedBy:638 invitee:542 isOccupant:false mefeInvitationIdIntValue:36624 role:Management Company type:remove_user unitId:105]
	// Put back
	// map[_id:6Z7e5ExhE4fP9L6R2 caseId:0 invitedBy:638 invitee:542 isOccupant:false mefeInvitationIdIntValue:36625 role:Management Company type:keep_default unitId:105]
	invites := []invite.Invite{{
		ID:         fmt.Sprintf("%s", uuid.Must(uuid.NewV4())), // Error 1062: Duplicate entry 'zbHFMYRpSiHmMNzgh' for key 'mefe_invitation_id_must_be_unique'
		InvitedBy:  638,
		Invitee:    542,
		Role:       "Management Company",
		IsOccupant: false,
		CaseID:     0,
		UnitID:     105,
		Type:       "remove_user",
	}}

	b, err := json.Marshal(&invites)
	if err != nil {
		t.Fatalf("Marshal to JSON error: %v", err)
	}
	resp, err := http.Post(server.URL, "application/json", bytes.NewBuffer(b))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("got %d, want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestPOSTdelete(t *testing.T) {
	ctx := context.Background()
	h, err := invite.New(ctx)
	if err != nil {
		log.WithError(err).Fatal("error setting configuration")
		return
	}
	defer h.DB.Close()
	app := h.BasicEngine()
	server := httptest.NewServer(app)
	defer server.Close()
	// Delete
	// map[_id:zbHFMYRpSiHmMNzgh caseId:0 invitedBy:638 invitee:542 isOccupant:false mefeInvitationIdIntValue:36624 role:Management Company type:remove_user unitId:105]
	// Put back
	// map[_id:6Z7e5ExhE4fP9L6R2 caseId:0 invitedBy:638 invitee:542 isOccupant:false mefeInvitationIdIntValue:36625 role:Management Company type:keep_default unitId:105]
	invites := []invite.Invite{{
		ID:         fmt.Sprintf("%s", uuid.Must(uuid.NewV4())), // Error 1062: Duplicate entry 'zbHFMYRpSiHmMNzgh' for key 'mefe_invitation_id_must_be_unique'
		InvitedBy:  638,
		Invitee:    542,
		Role:       "Management Company",
		IsOccupant: false,
		CaseID:     0,
		UnitID:     105,
		Type:       "remove_user",
	}}

	b, err := json.Marshal(&invites)
	if err != nil {
		t.Fatalf("Marshal to JSON error: %v", err)
	}
	resp, err := http.Post(server.URL, "application/json", bytes.NewBuffer(b))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("got %d, want %d", resp.StatusCode, http.StatusOK)
	}
	expected := fmt.Sprintf("Pushed 1")
	actual, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(actual), expected) {
		t.Errorf("got %s, want %s", actual, expected)
	}
}

func TestPOSTputback(t *testing.T) {
	ctx := context.Background()
	h, err := invite.New(ctx)
	if err != nil {
		log.WithError(err).Fatal("error setting configuration")
		return
	}
	defer h.DB.Close()
	app := h.BasicEngine()
	server := httptest.NewServer(app)
	defer server.Close()

	invites := []invite.Invite{{
		ID:         fmt.Sprintf("%s", uuid.Must(uuid.NewV4())), // Error 1062: Duplicate entry 'zbHFMYRpSiHmMNzgh' for key 'mefe_invitation_id_must_be_unique'
		InvitedBy:  638,
		Invitee:    542,
		Role:       "Management Company",
		IsOccupant: false,
		CaseID:     0,
		UnitID:     105,
		Type:       "keep_default",
	}}

	b, err := json.Marshal(&invites)
	if err != nil {
		t.Fatalf("Marshal to JSON error: %v", err)
	}
	resp, err := http.Post(server.URL, "application/json", bytes.NewBuffer(b))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("got %d, want %d", resp.StatusCode, http.StatusOK)
	}
	expected := fmt.Sprintf("Pushed 1")
	actual, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(actual), expected) {
		t.Errorf("got %s, want %s", actual, expected)
	}
}
