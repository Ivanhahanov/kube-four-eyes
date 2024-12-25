package coordination

import "webhook/pkg/storage"

func GetLog() []string {
	logs := []string{}
	res := storage.DB().GetMany("log")
	for _, r := range res {
		logs = append(logs, string(r.Value))
	}
	return logs
}
