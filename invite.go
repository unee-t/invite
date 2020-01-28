//go:generate -command asset go run asset.go
//go:generate asset -wrap=esql invite_user_to_a_case.sql
//go:generate asset -wrap=esql add_invitation_sent_message_to_a_case_v3.0.sql
//go:generate asset -wrap=esql invite_user_in_a_role_in_a_unit.sql

package invite

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"time"

	"github.com/apex/log"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/endpoints"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/tj/go/http/response"
// This is a hardcoded variable <-- should be moved
	"github.com/unee-t-ins/env"
// END This is a hardcoded variable <-- should be moved
)

// These get autofilled by goreleaser
var pingPollingFreq = 5 * time.Second

type handler struct {
	DSN            string
	Domain         string
	APIAccessToken string
	DB             *sql.DB
	Env            env.Env
	Log            *log.Entry
}

// Invite loosely models ut_invitation_api_data. JSON binding come from MEFE API /api/pending-invitations
type Invite struct {
	ID               string `json:"_id"` // mefe_invitation_id (must be unique)
	MefeInvitationID int    `json:"mefeInvitationIdIntValue"`
	InvitedBy        int    `json:"invitedBy"`
	Invitee          int    `json:"invitee"`
	Role             string `json:"role"`
	IsOccupant       bool   `json:"isOccupant"`
	CaseID           int    `json:"caseId"`
	UnitID           int    `json:"unitId"`
	Type             string `json:"type"` // invitation type
}

// New setups the configuration assuming various parameters have been setup in the AWS account
func New(ctx context.Context) (h handler, err error) {

	var logWithRequestID *log.Entry

	ctxObj, ok := lambdacontext.FromContext(ctx)
	if ok {
		logWithRequestID = log.WithFields(log.Fields{
			"requestID": ctxObj.AwsRequestID,
		})
	} else {
		logWithRequestID = log.WithFields(log.Fields{})
	}

	cfg, err := external.LoadDefaultAWSConfig(external.WithSharedConfigProfile("ins-dev"))
	if err != nil {
		log.WithError(err).Fatal("setting up credentials")
		return
	}
	cfg.Region = endpoints.ApSoutheast1RegionID
	e, err := env.New(cfg)
	if err != nil {
		log.WithError(err).Warn("error getting unee-t env")
	}

	// Check for CASE_HOST override
	var casehost string
	val, ok := os.LookupEnv("CASE_HOST")
	if ok {
		log.Infof("CASE_HOST overridden by local env: %s", val)
		casehost = val
	} else {
		casehost = fmt.Sprintf("https://%s", e.Udomain("case"))
	}

	h = handler{
		DSN:            e.BugzillaDSN(),
		Domain:         casehost,
		APIAccessToken: e.GetSecret("API_ACCESS_TOKEN"),
		Env:            e,
		Log:            logWithRequestID,
	}

	h.Log.Infof("h.Env.Code is %d, Frontend URL: %v", h.Env.Code, h.Domain)

	h.DB, err = sql.Open("mysql", h.DSN)
	if err != nil {
		log.WithError(err).Fatal("error opening database")
		return
	}

	microservicecheck := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "microservice",
			Help: "Version with DB ping check",
		},
		[]string{
			"commit",
		},
	)

	version := os.Getenv("UP_COMMIT")

	go func() {
		for {
			if h.DB.Ping() == nil {
				microservicecheck.WithLabelValues(version).Set(1)
			} else {
				microservicecheck.WithLabelValues(version).Set(0)
			}
			time.Sleep(pingPollingFreq)
		}
	}()

	err = prometheus.Register(microservicecheck)
	if err != nil {
		log.WithError(err).Warn("prom already registered")
		return h, nil
	}
	return
}

func (h handler) BasicEngine() http.Handler {

	app := mux.NewRouter()
	app.HandleFunc("/fail", fail).Methods("GET")
	app.Handle("/metrics", promhttp.Handler()).Methods("GET")

	// Pulls data from MEFE (doesn't really need to be protected, since input is already trusted)
	app.HandleFunc("/", h.handlePull).Methods("GET")

	// Push a POST of a JSON payload of the invite (ut_invitation_api_data)
	app.HandleFunc("/", h.handlePush).Methods("POST")

	if os.Getenv("UP_STAGE") == "" {
		// local dev, get around permissions
		return app
	}
	return env.Protect(app, h.APIAccessToken)
}

func (h handler) lookupRoleID(roleName string) (IDRoleType int, err error) {
	err = h.DB.QueryRow("SELECT id_role_type FROM ut_role_types WHERE role_type=?", roleName).Scan(&IDRoleType)
	return IDRoleType, err
}

func (h handler) step1Insert(invite Invite) (err error) {
	roleID, err := h.lookupRoleID(invite.Role)
	if err != nil {
		return
	}
	h.Log.Infof("%s role converted to id: %d", invite.Role, roleID)

	_, err = h.DB.Exec(
		`INSERT INTO ut_invitation_api_data (mefe_invitation_id,
			mefe_invitation_id_int_value,
			bzfe_invitor_user_id,
			bz_user_id,
			user_role_type_id,
			is_occupant,
			bz_case_id,
			bz_unit_id,
			invitation_type,
			is_mefe_only_user,
			user_more
		) VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
		invite.ID,
		invite.MefeInvitationID,
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

func esql(a asset) asset {
	return a
}

func (h handler) runsql(sqlfile asset, invite Invite) (err error) {
	execSQL := fmt.Sprintf(sqlfile.Content, invite.ID, invite.MefeInvitationID, h.Env.Code)
	h.Log.WithFields(log.Fields{
		"invite":  invite,
		"sqlfile": sqlfile.Name,
		"env":     h.Env.Code,
		//		"SQL":     execSQL,
	}).Info("exec sql")
	_, err = h.DB.Exec(execSQL)
	if err != nil {
		h.Log.WithError(err).Error("execing sql failed")
	}
	return
}

func (h handler) inviteUsertoUnit(invites []Invite) (result error) {
	for _, invite := range invites {

		ctx := h.Log.WithFields(log.Fields{
			"invite": invite,
		})

		// Insert into ut_invitation_api_data

		err := h.step1Insert(invite)
		if err != nil {
			ctx.WithError(err).Error("failed to run step1Insert")
			return err
		}

		ctx.Info("step1")

		// Run invite_user_in_a_role_in_a_unit.sql
		err = h.runsql(invite_user_in_a_role_in_a_unit, invite)
		if err != nil {
			ctx.WithError(err).Error("failed to run invite_user_in_a_role_in_a_unit.sql")
			return err
		}
		ctx.Info("runsql")

	}
	return result
}

func (h handler) queue(invites []Invite) error {
// This is a hardcoded variable <-- should be moved
	var queue = fmt.Sprintf("https://sqs.ap-southeast-1.amazonaws.com/%s/invites", h.Env.AccountID)
	h.Log.Infof("%d invites to queue: %s", len(invites), queue)
// END This is a hardcoded variable
	client := sqs.New(h.Env.Cfg)

	// We can queue as much as 10 at a time
	for len(invites) > 0 {
		n := 10
		if n > len(invites) {
			n = len(invites)
		}
		chunk := invites[:n]
		invites = invites[n:]
		h.Log.Infof("Chunk: %+v", chunk)

		var entries []sqs.SendMessageBatchRequestEntry
		for _, v := range chunk {
			entry := sqs.SendMessageBatchRequestEntry{Id: aws.String(v.ID)}
			MessageBody, err := json.Marshal(v)
			if err != nil {
				return err
			}
			entry.MessageBody = aws.String(string(MessageBody))
			entries = append(entries, entry)
		}

		req := client.SendMessageBatchRequest(&sqs.SendMessageBatchInput{
			Entries:  entries,
			QueueUrl: aws.String(queue),
		})

		h.Log.Infof("Entries: %#v", entries)

		resp, err := req.Send(context.TODO())
		if err != nil {
			h.Log.WithError(err).Error("failed to queue")
			return err
		}
		log.Infof("Queued SQS resp: %s", len(chunk), resp)

	}

	return nil
}

func (h handler) processInvites(invites []Invite) (result error) {

	log.Infof("Number of invites to process: %d", len(invites))

	// Detect if we are running on AWS lambda, as opposed to just locally
	if s := os.Getenv("UP_STAGE"); s != "" {
		log.Info("Running in a Lambda context, will add invites to queue")
		// Instead of processing the invites here, we queue them and process them because
		// * we want to prevent race conditions, queue lambda has concurrency of 1
		// * we want to prevent API gateway timeouts (lambda without APIGW has a timeout of max 15 mins!)
		return h.queue(invites)
	}

	log.Warn("Processing locally, assuming no time outs and the request is uninterrupted")

	for num, invite := range invites {

		ctx := h.Log.WithFields(log.Fields{
			"num":    num,
			"invite": invite,
		})
		err := h.ProcessInvite(invite)

		if err != nil {
			ctx.WithError(err).Error("failed to run mark invite as processed")
			result = multierror.Append(result, multierror.Prefix(err, invite.ID))
		}
	}
	return result
}

func (h handler) ProcessInvite(invite Invite) (result error) {

	dt, err := h.checkProcessedDatetime(invite)
	if err == nil && dt.Valid {
		h.Log.Warnf("already processed %s", time.Since(dt.Time))
		err = h.markInvitesProcessed([]string{invite.ID})
		if err != nil {
			h.Log.WithError(err).Error("failed to mark invite as processed")
			return err
		}
		// Stop processing
		return nil
	}

	h.Log.WithField("id", invite.MefeInvitationID).Infof("Checking if invitation exists already")
	_, err = h.checkIfInvitationExistsAlready(invite)
	// If there is an error, we know that invite ID does not exist in the ut_invitation_api_data table already
	if err != nil {
		h.Log.Infof("Step 1, inserting %s", invite.ID)
		err = h.step1Insert(invite)
		if err != nil {
			h.Log.WithError(err).Error("failed to run step1Insert")
			return err
		}
	} else {
		h.Log.Infof("Skipping Step 1, as %s appears to be inserted already", invite.ID)
	}

	if invite.CaseID == 0 { // if there is no CaseID, invite user to a role in the unit
		err = h.runsql(invite_user_in_a_role_in_a_unit, invite)
		if err != nil {
			h.Log.WithError(err).Error("failed to invite_user_in_a_role_in_a_unit")
			return err
		}
	} else { // if there is a CaseID then invite to a case
		h.Log.Info("Inviting to case")
		err = h.runsql(invite_user_to_a_case, invite)
		if err != nil {
			h.Log.WithError(err).Error("failed to invite_user_to_a_case")
			return err
		}
	}

	dtProcessCheck, err := h.checkProcessedDatetime(invite)
	// If there is an error, there is no record
	if err != nil {
		h.Log.WithError(err).Errorf("no process datetime: %s", invite.ID)
		return err
	}

	h.Log.Infof("Step 3, telling frontend we are done, since %s was processed %s ago", invite.ID,
		time.Since(dtProcessCheck.Time))

	err = h.markInvitesProcessed([]string{invite.ID})
	if err != nil {
		h.Log.WithError(err).Error("failed to run mark invite as processed")
		return err
	}

	if invite.CaseID != 0 {
		h.Log.Infof("Step 4, with case id %d, send a message", invite.CaseID)
		err = h.runsql(add_invitation_sent_message_to_a_case_v3, invite)
		if err != nil {
			h.Log.WithError(err).Error("failed to run add_invitation_sent_message_to_a_case_v3")
			return err
		}
	} else {
		h.Log.Warn("Skipping (Step 4) 2_add_invitation_sent_message_to_a_case_v3.0.sql since CaseID is empty")
	}
	return nil
}

func (h handler) getInvites() (lr []Invite, err error) {
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
		h.Log.WithError(err).Error("marshalling")
		return err
	}

	h.Log.Infof("Marking as done: %s", jids)

	payload := strings.NewReader(string(jids))

	url := h.Domain + "/api/pending-invitations/done?accessToken=" + h.APIAccessToken
	req, err := http.NewRequest("PUT", url, payload)
	if err != nil {
		h.Log.WithError(err).Error("making PUT")
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Cache-Control", "no-cache")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		h.Log.WithError(err).Error("PUT request")
		return err
	}

	if res.StatusCode != 200 {
		log.Warnf("StatusCode is: %d", res.StatusCode)
	}

	// If run in parallel, it is concievable that an input is processed before it can be marked as done
	// hence the "Acted on invitations" can be different from the "Input invitations"

	// defer res.Body.Close()
	// body, err := ioutil.ReadAll(res.Body)
	// if err != nil {
	// 	h.Log.WithError(err).Error("reading body")
	// }

	// i, err := strconv.Atoi(string(body))
	// if err != nil {
	// 	h.Log.WithError(err).Error("cannot convert into integer")
	// }

	//log.Infof("Response: %v", res)
	//log.Infof("Num: %d", i)
	//log.Infof("Body: %s", string(body))
	// if i != len(ids) {
	// 	return fmt.Errorf("Acted on %d invitations, but %d were submitted", i, len(ids))
	// }

	return

}

func (h handler) handlePull(w http.ResponseWriter, r *http.Request) {

	log.Infof("handlePull: %s", r.Header.Get("User-Agent"))

	invites, err := h.getInvites()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// log.Infof("Input %+v", invites)

	err = h.processInvites(invites)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response.OK(w, fmt.Sprintf("Pulled %d", len(invites)))

}

func (h handler) handlePush(w http.ResponseWriter, r *http.Request) {
	// TODO: Update to queue

	h.Log = log.WithFields(log.Fields{
		"requestID": r.Header.Get("X-Request-Id"),
	})

	buf := &bytes.Buffer{}
	tee := io.TeeReader(r.Body, buf)
	defer r.Body.Close()
	dec := json.NewDecoder(tee)

	var invites []Invite
	err := dec.Decode(&invites)

	if err != nil {
		dump, _ := httputil.DumpRequest(r, false)
		h.Log.WithError(err).Errorf("%s\n%v", dump, buf)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.Log.WithField("len(invites)", len(invites)).Info("handlePush")

	if len(invites) < 1 {
		response.BadRequest(w, "Empty payload")
		return
	}

	err = h.inviteUsertoUnit(invites)
	if err != nil {
		h.Log.WithError(err).Error("push inviteUsertoUnit")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response.OK(w, fmt.Sprintf("Pushed %d", len(invites)))

}

func (h handler) runProc(w http.ResponseWriter, r *http.Request) {

	var outArg string
	_, err := h.DB.Exec("CALL ProcName")
	if err != nil {
		h.Log.WithError(err).Error("running proc")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response.OK(w, outArg)

}

func (h handler) checkProcessedDatetime(i Invite) (ProcessedDatetime mysql.NullTime, err error) {
	err = h.DB.QueryRow("SELECT processed_datetime FROM ut_invitation_api_data WHERE mefe_invitation_id_int_value=?", i.MefeInvitationID).Scan(&ProcessedDatetime)
	return ProcessedDatetime, err
}

func (h handler) checkIfInvitationExistsAlready(i Invite) (inviteID string, err error) {
	err = h.DB.QueryRow("SELECT mefe_invitation_id FROM ut_invitation_api_data WHERE mefe_invitation_id_int_value=?",
		i.MefeInvitationID).Scan(&inviteID)
	return inviteID, err
}

func fail(w http.ResponseWriter, r *http.Request) {
	log.Warn("5xx")
	http.Error(w, "5xx", http.StatusInternalServerError)
}
