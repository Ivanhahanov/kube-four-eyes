package coordination

import (
	"strings"
	"webhook/pkg/models"
	"webhook/pkg/storage"

	"github.com/google/uuid"
)

const (
	prefix     = "co"
	sep        = "/"
	reqType    = "req"
	statusType = "status"
	onlineType = "online"
)

func NewRequest(ar models.AccessRequest) (string, error) {
	req, _ := encodeRequest(ar)
	id := uuid.New().String()
	err := storage.DB().PutTemporary(strings.Join([]string{reqType, id}, sep), req, 180)
	if err != nil {
		return "", err
	}
	return id, nil
}

func GetAllRequests() (map[string]models.AccessRequest, error) {
	var ars = map[string]models.AccessRequest{}
	res := storage.DB().GetMany(reqType)
	for _, r := range res {
		ar, _ := decodeRequest(r.Value)
		rid := strings.Split(string(r.Key), sep)[1]
		ars[rid] = ar
	}
	return ars, nil
}

func GetRequest(rid string) (models.AccessRequest, error) {
	res := storage.DB().Get(strings.Join([]string{reqType, rid}, sep))
	ar, err := decodeRequest(res)
	if err != nil {
		return ar, err
	}
	return ar, nil
}

func ChangeStatus(requestId, uid, status string) error {
	key := strings.Join([]string{statusType, requestId, uid}, sep)
	return storage.DB().Put(key, status)
}

func GetStatuses(requestId string) (map[string]string, error) {
	var statuses = map[string]string{}
	res := storage.DB().GetMany(strings.Join([]string{statusType, requestId, ""}, sep))
	for _, r := range res {
		uid := strings.Split(string(r.Key), sep)[2]
		statuses[uid] = string(r.Value)
	}
	return statuses, nil
}

func SetOnline(requestId, uid string) error {
	key := strings.Join([]string{onlineType, requestId, uid}, sep)
	return storage.DB().Put(key, "online")
}

func SetOffline(requestId, uid string) {
	key := strings.Join([]string{onlineType, requestId, uid}, sep)
	storage.DB().Delete(key)
}

func GetOnline(requestId string) (map[string]string, error) {
	var statuses = map[string]string{}
	res := storage.DB().GetMany(strings.Join([]string{onlineType, requestId, ""}, sep))
	for _, r := range res {
		uid := strings.Split(string(r.Key), sep)[2]
		statuses[uid] = string(r.Value)
	}
	return statuses, nil
}
