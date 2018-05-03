package main

import (
	"encoding/json"
	"net/http"
	"os"

	jsonhandler "github.com/apex/log/handlers/json"
	"github.com/tj/go/http/response"

	"github.com/apex/log"
	"github.com/apex/log/handlers/text"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

// TODO: Maybe put env variables in this this config struct too
type handler struct{ db *sql.DB }

// {{DOMAIN}}/api/pending-invitations?accessToken={{API_ACCESS_TOKEN}}
type listInvitesResponse []struct {
	ID         string `json:"_id"`
	InvitedBy  int    `json:"invitedBy"`
	Invitee    int    `json:"invitee"`
	Role       string `json:"role"`
	IsOccupant bool   `json:"isOccupant"`
	CaseID     int    `json:"caseId"`
	UnitID     int    `json:"unitId"`
	Type       string `json:"type"`
}

func init() {
	if os.Getenv("UP_STAGE") == "" {
		log.SetHandler(text.Default)
	} else {
		log.SetHandler(jsonhandler.Default)
	}
}

func main() {

	db, err := sql.Open("mysql", os.Getenv("DSN"))
	if err != nil {
		log.WithError(err).Fatal("error opening database")
	}

	defer db.Close()

	addr := ":" + os.Getenv("PORT")
	http.Handle("/", handler{db: db})
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.WithError(err).Fatal("error listening")
	}

}

func (h handler) lookupRoleID(roleName string) (id_role_type int, err error) {
	err = h.db.QueryRow("SELECT id_role_type FROM ut_role_types WHERE role_type=?", roleName).Scan(&id_role_type)
	return id_role_type, err
}

func (h handler) set(lr listInvitesResponse) error {

	for _, invite := range lr {
		log.Infof("Processing invite: %+v", invite)

		roleID, err := h.lookupRoleID(invite.Role)
		if err != nil {
			return err
		}
		log.Infof("%s role converted to id: %d", invite.Role, roleID)

		result, err := h.db.Exec(
			`INSERT INTO ut_invitation_api_data (mefe_invitation_id,
			bzfe_invitor_user_id,
			bz_user_id,
			user_role_type_id,
			is_occupant,
			bz_case_id,
			bz_unit_id,
			invitation_type,
			is_mefe_only_user,
			user_more
		) VALUES (?,?,?,?,?,?,?,?,?,?)`,
			invite.ID,
			invite.InvitedBy,
			invite.Invitee,
			roleID,
			invite.IsOccupant,
			invite.CaseID,
			invite.UnitID,
			invite.Type,
			1,
			"Use Unee-T for a faster reply",
		)
		if err != nil {
			return err
		}

		log.Infof("Exec result %v", result)

	}

	return nil

}

func getInvites() (lr listInvitesResponse, err error) {
	resp, err := http.Get(os.Getenv("DOMAIN") + "/api/pending-invitations?accessToken=" + os.Getenv("API_ACCESS_TOKEN"))
	if err != nil {
		return lr, err
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&lr)
	return lr, err
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("X-Robots-Tag", "none")

	lr, err := getInvites()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Infof("Input %+v", lr)

	err = h.set(lr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response.OK(w, "Worked")

}
