package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	uuid "github.com/satori/go.uuid"
	"github.com/unee-t/invite"
)

type SQSBatch struct {
	ID          string `json:"Id"`
	MessageBody string `json:"MessageBody"`
}

// [{
//   "invitationId": "4jd23hc4jv5",
//     "invitedBy": 1,
//     "invitee": 2,
//     "role": "Agent",
//     "isOccupant": false,
//     "unitId": 8,
//     "type": "replace_default"
// }]

func main() {

	for profileID := 2000; profileID <= 2349; profileID++ {
		invitefilename := fmt.Sprintf("invites-%d.json", profileID)
		var sqs []SQSBatch

		for i := 1924; i < 1925; i++ {
			u1 := uuid.Must(uuid.NewV4())
			ivt := invite.Invite{
				ID:         fmt.Sprintf("%s", u1),
				InvitedBy:  2349,
				Invitee:    profileID,
				Role:       "Agent",
				IsOccupant: false,
				UnitID:     i,
				Type:       "keep_default",
			}
			invitesJSON, _ := json.Marshal(ivt)
			sqs = append(sqs, SQSBatch{
				ID:          fmt.Sprintf("%s", u1),
				MessageBody: string(invitesJSON),
			})
		}

		sqsJSON, _ := json.MarshalIndent(sqs, "", "\t")
		err := ioutil.WriteFile(invitefilename, sqsJSON, 0644)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Wrote %s\n", invitefilename)
	}

}
