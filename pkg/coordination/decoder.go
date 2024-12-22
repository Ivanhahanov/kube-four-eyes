package coordination

import (
	"bytes"
	"encoding/gob"
	"webhook/pkg/models"
)

func decodeRequest(v []byte) (models.AccessRequest, error) {
	var buf bytes.Buffer
	var ar models.AccessRequest
	buf.Write(v)
	err := gob.NewDecoder(&buf).Decode(&ar)
	if err != nil {
		return ar, err
	}
	return ar, nil
}

func encodeRequest(ar models.AccessRequest) (string, error) {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(&ar)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
