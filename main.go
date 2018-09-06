package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	jsonhandler "github.com/apex/log/handlers/json"
	"github.com/aws/aws-sdk-go-v2/aws/endpoints"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/gorilla/mux"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/tj/go/http/response"
	"github.com/unee-t/env"

	"github.com/apex/log"
	"github.com/apex/log/handlers/text"

	"database/sql"

	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
)

// These get autofilled by goreleaser
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

type handler struct {
	DSN            string // e.g. "bugzilla:secret@tcp(auroradb.dev.unee-t.com:3306)/bugzilla?multiStatements=true&sql_mode=TRADITIONAL"
	Domain         string // e.g. https://dev.case.unee-t.com
	APIAccessToken string // e.g. O8I9svDTizOfLfdVA5ri
	db             *sql.DB
	Code           env.EnvCode
}

// loosely models ut_invitation_api_data. JSON binding come from MEFE API /api/pending-invitations
type invite struct {
	ID         string `json:"_id"` // mefe_invitation_id (must be unique)
	InvitedBy  int    `json:"invitedBy"`
	Invitee    int    `json:"invitee"`
	Role       string `json:"role"`
	IsOccupant bool   `json:"isOccupant"`
	CaseID     int    `json:"caseId"`
	UnitID     int    `json:"unitId"`
	Type       string `json:"type"` // invitation type
}

func init() {
	if os.Getenv("UP_STAGE") == "" {
		log.SetHandler(text.Default)
	} else {
		log.SetHandler(jsonhandler.Default)
	}
}

// New setups the configuration assuming various parameters have been setup in the AWS account
func New() (h handler, err error) {

	cfg, err := external.LoadDefaultAWSConfig(external.WithSharedConfigProfile("uneet-dev"))
	if err != nil {
		log.WithError(err).Fatal("setting up credentials")
		return
	}
	cfg.Region = endpoints.ApSoutheast1RegionID
	e, err := env.New(cfg)
	if err != nil {
		log.WithError(err).Warn("error getting unee-t env")
	}

	// Check for MYSQL_HOST override
	var mysqlhost string
	val, ok := os.LookupEnv("MYSQL_HOST")
	if ok {
		log.Infof("MYSQL_HOST overridden by local env: %s", val)
		mysqlhost = val
	} else {
		mysqlhost = e.Udomain("auroradb")
	}

	// Check for CASE_HOST override
	var casehost string
	val, ok = os.LookupEnv("CASE_HOST")
	if ok {
		log.Infof("CASE_HOST overridden by local env: %s", val)
		casehost = val
	} else {
		casehost = fmt.Sprintf("https://%s", e.Udomain("case"))
	}

	h = handler{
		DSN: fmt.Sprintf("%s:%s@tcp(%s:3306)/bugzilla?multiStatements=true&sql_mode=TRADITIONAL",
			e.GetSecret("MYSQL_USER"),
			e.GetSecret("MYSQL_PASSWORD"),
			mysqlhost),
		Domain:         casehost,
		APIAccessToken: e.GetSecret("API_ACCESS_TOKEN"),
		Code:           e.Code,
	}

	log.Infof("h.Code is %d", h.Code)
	log.Infof("Frontend URL: %v", h.Domain)

	h.db, err = sql.Open("mysql", h.DSN)
	if err != nil {
		log.WithError(err).Fatal("error opening database")
		return
	}

	return

}

func (h handler) BasicEngine() http.Handler {

	app := mux.NewRouter()
	app.HandleFunc("/version", showversion).Methods("GET")
	app.HandleFunc("/health_check", h.ping).Methods("GET")
	app.HandleFunc("/fail", fail).Methods("GET")

	app.HandleFunc("/check", h.processedAlready).Methods("GET")

	// Pulls data from MEFE (doesn't really need to be protected, since input is already trusted)
	app.HandleFunc("/", h.handlePull).Methods("GET")

	// Push a POST of a JSON payload of the invite (ut_invitation_api_data)
	app.HandleFunc("/", h.handlePush).Methods("POST")

	return app
}

func main() {

	h, err := New()
	if err != nil {
		log.WithError(err).Fatal("error setting configuration")
		return
	}

	defer h.db.Close()

	addr := ":" + os.Getenv("PORT")
	app := h.BasicEngine()

	if err := http.ListenAndServe(addr, env.Protect(app, h.APIAccessToken)); err != nil {
		log.WithError(err).Fatal("error listening")
	}

}

func (h handler) lookupRoleID(roleName string) (IDRoleType int, err error) {
	err = h.db.QueryRow("SELECT id_role_type FROM ut_role_types WHERE role_type=?", roleName).Scan(&IDRoleType)
	return IDRoleType, err
}

func (h handler) step1Insert(invite invite) (err error) {
	roleID, err := h.lookupRoleID(invite.Role)
	if err != nil {
		return
	}
	log.Infof("%s role converted to id: %d", invite.Role, roleID)

	_, err = h.db.Exec(
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
	return
}

func (h handler) runsql(sqlfile string, invite invite) (err error) {
	sqlscript, err := ioutil.ReadFile(fmt.Sprintf("sql/%s", sqlfile))
	if err != nil {
		return
	}
	log.Infof("Running %s with invite id %s with env %d", sqlfile, invite.ID, h.Code)
	_, err = h.db.Exec(fmt.Sprintf(string(sqlscript), invite.ID, h.Code))
	if err != nil {
		log.WithError(err).Error("running sql failed")
	}
	return
}

func (h handler) inviteUsertoUnit(invites []invite) (result error) {
	for _, invite := range invites {

		ctx := log.WithFields(log.Fields{
			"invite": invite,
		})

		// Insert into ut_invitation_api_data

		err := h.step1Insert(invite)
		if err != nil {
			ctx.WithError(err).Error("failed to run step1Insert")
			return err
		}

		// Run invite_user_in_a_role_in_a_unit.sql
		err = h.runsql("invite_user_in_a_role_in_a_unit.sql", invite)

		if err != nil {
			ctx.WithError(err).Error("failed to run invite_user_in_a_role_in_a_unit.sql")
			return err
		}

	}
	return result
}

func (h handler) processInvite(invites []invite) (result error) {

	for _, invite := range invites {

		ctx := log.WithFields(log.Fields{
			"invite": invite,
		})

		// Processing invite one by one. If it fails, we move onto next one.

		dt, err := h.checkProcessedDatetime(invite.ID)
		if err == nil {
			ctx.Warnf("already processed %s", time.Since(dt.Time))
			err = h.markInvitesProcessed([]string{invite.ID})
			if err != nil {
				ctx.WithError(err).Error("failed to run mark invite as processed")
				result = multierror.Append(result, multierror.Prefix(err, invite.ID))
			}
			continue
		}

		ctx.Info("Step 1, inserting")

		err = h.step1Insert(invite)
		if err != nil {
			ctx.WithError(err).Error("failed to run step1Insert")
			result = multierror.Append(result, multierror.Prefix(err, invite.ID))
			continue
		}

		ctx.Info("Step 2, running SQL")

		err = h.runsql("1_process_one_invitation_all_scenario_v3.0.sql", invite)
		if err != nil {
			ctx.WithError(err).Error("failed to run 1_process_one_invitation_all_scenario_v3.0.sql")
			result = multierror.Append(result, multierror.Prefix(err, invite.ID))
			continue
		}

		ctx.Info("Step 3, telling frontend we are done")

		err = h.markInvitesProcessed([]string{invite.ID})
		if err != nil {
			ctx.WithError(err).Error("failed to run mark invite as processed")
			result = multierror.Append(result, multierror.Prefix(err, invite.ID))
			continue
		}

		if invite.CaseID != 0 {
			ctx.Infof("Step 4, with case id %d, send a message", invite.CaseID)
			err = h.runsql("2_add_invitation_sent_message_to_a_case_v3.0.sql", invite)
			if err != nil {
				ctx.WithError(err).Error("failed to run 2_add_invitation_sent_message_to_a_case_v3.0.sql")
				result = multierror.Append(result, multierror.Prefix(err, invite.ID))
				continue
			}
		} else {
			ctx.Warn("Skipping (Step 4) 2_add_invitation_sent_message_to_a_case_v3.0.sql since CaseID is empty")
		}
	}
	return result
}

func (h handler) getInvites() (lr []invite, err error) {
	resp, err := http.Get(h.Domain + "/api/pending-invitations?accessToken=" + h.APIAccessToken)
	if err != nil {
		return lr, err
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&lr)
	return lr, err
}

func (h handler) markInvitesProcessed(ids []string) (err error) {

	jids, err := json.Marshal(ids)
	if err != nil {
		log.WithError(err).Error("marshalling")
		return err
	}

	log.Infof("Marking as done: %s", jids)

	payload := strings.NewReader(string(jids))

	url := h.Domain + "/api/pending-invitations/done?accessToken=" + h.APIAccessToken
	req, err := http.NewRequest("PUT", url, payload)
	if err != nil {
		log.WithError(err).Error("making PUT")
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Cache-Control", "no-cache")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.WithError(err).Error("PUT request")
		return err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.WithError(err).Error("reading body")
		return err
	}

	i, err := strconv.Atoi(string(body))
	if err != nil {
		log.WithError(err).Error("reading body")
		return err
	}

	//log.Infof("Response: %v", res)
	//log.Infof("Num: %d", i)
	//log.Infof("Body: %s", string(body))
	if i != len(ids) {
		return fmt.Errorf("Acted on %d invitations, but %d were submitted", i, len(ids))
	}

	return

}

func (h handler) handlePull(w http.ResponseWriter, r *http.Request) {

	log.Info("handlePull")

	w.Header().Set("X-Robots-Tag", "none") // We don't want Google to index us

	invites, err := h.getInvites()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Infof("Input %+v", invites)

	err = h.processInvite(invites)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response.OK(w, fmt.Sprintf("Pulled %d", len(invites)))

}

func (h handler) handlePush(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)
	var invites []invite
	err := decoder.Decode(&invites)

	if err != nil {
		log.WithError(err).Errorf("Input error: %s", r.Body)
		response.BadRequest(w, "Invalid JSON")
		return
	}
	defer r.Body.Close()

	log.Infof("Input %+v", invites)

	log.Infof("Length %d", len(invites))

	if len(invites) < 1 {
		response.BadRequest(w, "Empty payload")
		return
	}

	err = h.inviteUsertoUnit(invites)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response.OK(w, fmt.Sprintf("Pushed %d", len(invites)))

}

func (h handler) runProc(w http.ResponseWriter, r *http.Request) {

	var outArg string
	_, err := h.db.Exec("CALL ProcName")
	if err != nil {
		log.WithError(err).Error("running proc")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response.OK(w, outArg)

}

func (h handler) processedAlready(w http.ResponseWriter, r *http.Request) {
	MefeInvitationID := r.URL.Query().Get("id")
	if MefeInvitationID == "" {
		response.BadRequest(w, "Missing id")
		return
	}

	ctx := log.WithFields(log.Fields{
		"mefe_invitation_id": MefeInvitationID,
	})

	dt, err := h.checkProcessedDatetime(MefeInvitationID)
	if err != nil {
		ctx.WithError(err).Error("checking if processed")
		response.BadRequest(w, "Not processed")
		return
	}

	if dt.Valid {
		response.OK(w, "Got a date")
	} else {
		response.BadRequest(w, "Bad date")
	}
}

func (h handler) checkProcessedDatetime(MefeInvitationID string) (ProcessedDatetime mysql.NullTime, err error) {
	err = h.db.QueryRow("SELECT processed_datetime FROM ut_invitation_api_data WHERE mefe_invitation_id=?", MefeInvitationID).Scan(&ProcessedDatetime)
	return ProcessedDatetime, err
}

func showversion(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%v, commit %v, built at %v", version, commit, date)
}

func fail(w http.ResponseWriter, r *http.Request) {
	log.Warn("5xx")
	http.Error(w, "5xx", http.StatusInternalServerError)
}

func (h handler) ping(w http.ResponseWriter, r *http.Request) {
	err := h.db.Ping()
	if err != nil {
		log.WithError(err).Error("failed to ping database")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	fmt.Fprintf(w, "OK")
}
