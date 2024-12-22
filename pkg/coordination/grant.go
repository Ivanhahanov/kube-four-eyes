package coordination

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"webhook/pkg/helpers"
	"webhook/pkg/storage"

	"github.com/gofiber/fiber/v2/log"
)

type Access struct {
	Role    string
	Timeout string
}

func ConvertDurationToMinutes(duration string) (int64, error) {
	var totalMinutes int64
	var numBuffer string

	for i := 0; i < len(duration); i++ {
		char := duration[i]

		// Check if the character is a digit
		if char >= '0' && char <= '9' {
			numBuffer += string(char)
			continue
		}

		// Process the time unit
		if numBuffer == "" {
			return 0, errors.New("invalid duration format")
		}

		value, err := strconv.Atoi(numBuffer)
		if err != nil {
			return 0, err
		}

		switch char {
		case 'm': // Minutes
			totalMinutes += int64(value)
		case 'h': // Hours
			totalMinutes += int64(value * 60)
		case 'd': // Days
			totalMinutes += int64(value * 1440)
		default:
			return 0, fmt.Errorf("unknown time unit '%c'", char)
		}

		// Reset the number buffer
		numBuffer = ""
	}

	if numBuffer != "" {
		return 0, errors.New("invalid duration format")
	}

	return totalMinutes, nil
}

func GrantUserAccess(rid string) error {
	ar, err := GetRequest(rid)
	if err != nil {
		log.Error(err)
	}
	key := strings.Join([]string{"access", ar.Email, rid}, "/")
	seconds, err := ConvertDurationToMinutes(ar.TimePeriod)
	if err != nil {
		return err
	}
	return storage.DB().PutTemporary(key, "ok", seconds*60)
}

func CheckUserAccess(name string) bool {
	name = strings.TrimPrefix(name, helpers.GetEnv("OIDC_PREFIX", "oidc:"))
	key := strings.Join([]string{"access", name}, "/")

	res := storage.DB().GetMany(key)
	for _, res := range res {
		if string(res.Value) == "ok" {
			return true
		}
	}
	return false
}
