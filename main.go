package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	jsonhandler "github.com/apex/log/handlers/json"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/ssmiface"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/tj/go/http/response"

	"github.com/apex/log"
	"github.com/apex/log/handlers/text"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

// TODO: Maybe put env variables in this this config struct too
type handler struct {
	DSN            string // e.g. "bugzilla:secret@tcp(auroradb.dev.unee-t.com:3306)/bugzilla?multiStatements=true&sql_mode=TRADITIONAL"
	Domain         string // e.g. https://dev.case.unee-t.com
	APIAccessToken string // e.g. O8I9svDTizOfLfdVA5ri
	db             *sql.DB
}

// {{DOMAIN}}/api/pending-invitations?accessToken={{API_ACCESS_TOKEN}}
type invite struct {
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

func udomain(svc string, stage string) string {
	if stage == "" {
		return fmt.Sprintf("%s.unee-t.com", svc)
	}
	return fmt.Sprintf("%s.%s.unee-t.com", svc, stage)
}

// New setups the configuration assuming various parameters have been setup in the AWS account
func New() (h handler, err error) {

	cfg, err := external.LoadDefaultAWSConfig(external.WithSharedConfigProfile("uneet-dev"))
	if err != nil {
		return
	}
	ssm := ssm.New(cfg)

	h = handler{
		DSN: fmt.Sprintf("bugzilla:%s@tcp(%s:3306)/bugzilla?multiStatements=true&sql_mode=TRADITIONAL",
			getSecret(ssm, "MYSQL_PASSWORD"),
			udomain("auroradb", getSecret(ssm, "STAGE"))),
		Domain:         fmt.Sprintf("https://%s", udomain("case", getSecret(ssm, "STAGE"))),
		APIAccessToken: getSecret(ssm, "API_ACCESS_TOKEN"),
	}

	log.Infof("Configuration: %v", h)

	h.db, err = sql.Open("mysql", h.DSN)
	if err != nil {
		log.WithError(err).Fatal("error opening database")
		return
	}

	return

}

func main() {

	h, err := New()
	if err != nil {
		log.WithError(err).Fatal("error setting configuration")
		return
	}

	defer h.db.Close()

	addr := ":" + os.Getenv("PORT")
	http.HandleFunc("/favicon.ico", http.NotFound)
	http.Handle("/", http.HandlerFunc(h.handleInvite))
	if err := http.ListenAndServe(addr, nil); err != nil {
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
	_, err = h.db.Exec(fmt.Sprintf(string(sqlscript), invite.ID))
	if err != nil {
		log.WithError(err).Error("running sql failed")
	}
	return
}

func (h handler) processInvite(invites []invite) (result error) {

	for _, invite := range invites {
		// Processing invite one by one. If it fails, we move onto next one.
		log.Infof("Processing invite: %+v", invite)

		// Step 1
		err := h.step1Insert(invite)
		if err != nil {
			result = multierror.Append(result, multierror.Prefix(err, invite.ID))
			continue
		}

		// Step 2
		err = h.runsql("1_process_one_invitation_all_scenario_v3.0.sql", invite)
		if err != nil {
			result = multierror.Append(result, multierror.Prefix(err, invite.ID))
			continue
		}

		// Step 3
		err = h.markInvitesProcessed([]string{invite.ID})
		if err != nil {
			result = multierror.Append(result, multierror.Prefix(err, invite.ID))
			continue
		}

		// Step 4
		err = h.runsql("2_add_invitation_sent_message_to_a_case_v3.0.sql", invite)
		if err != nil {
			result = multierror.Append(result, multierror.Prefix(err, invite.ID))
			continue
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

func (h handler) handleInvite(w http.ResponseWriter, r *http.Request) {

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

	response.OK(w, "Worked")

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

func getSecret(ssmapi ssmiface.SSMAPI, store string) string {
	in := &ssm.GetParameterInput{
		Name:           aws.String(store),
		WithDecryption: aws.Bool(true),
	}
	req := ssmapi.GetParameterRequest(in)
	out, err := req.Send()
	if err != nil {
		return ""
	}
	return aws.StringValue(out.Parameter.Value)
}
