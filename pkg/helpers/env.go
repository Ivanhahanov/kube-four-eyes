package helpers

import (
	"os"
	"strconv"

	"github.com/gofiber/fiber/v2/log"
)

func GetEnv(env, value string) string {
	variable := os.Getenv(env)
	if variable == "" {
		return value
	}
	return variable
}

func GetIntEnv(env string, value int) int {
	variable := os.Getenv(env)
	if variable == "" {
		return value
	}
	val, err := strconv.Atoi(variable)
	if err != nil {
		log.Errorf("can't convert env '%s' with value '%s' to int: %e", env, variable, err)
	}
	return val
}
