package coordination

import (
	"strings"
	"webhook/pkg/helpers"
	"webhook/pkg/storage"
)

type policy struct {
	Users     []string
	Threshold int
}

func Policy() *policy {
	return &policy{
		Users:     strings.Split(helpers.GetEnv("POLICY_USERS", "user1@example.com,user2@example.com"), ","),
		Threshold: helpers.GetIntEnv("POLICY_THRESHOLD", 2),
	}
}

func (p *policy) CheckPolicy(rid string) bool {
	approves := 0
	statuses := storage.DB().GetMany(strings.Join([]string{statusType, rid}, sep))
	for _, r := range statuses {
		if string(r.Value) == helpers.StatusApproved {
			approves += 1
		}
	}
	if approves >= p.Threshold {
		return true
	}
	return false
}

// type PolicyController struct{
// 	Policy Policy
// }

// func NewPolicyController() *PolicyController {
// 	return nil
// }

// func (pc *PolicyController)Run(){
// 	cli := storage.NewClient()
// 	cli.Watch(context.Background(), "/")
// }
