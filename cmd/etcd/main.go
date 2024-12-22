package main

import (
	"fmt"
	"webhook/pkg/coordination"
	"webhook/pkg/helpers"
	"webhook/pkg/models"
)

func generateRequests() {
	ar := models.AccessRequest{
		Username:   "user",
		Email:      "user@example.com",
		Role:       "cluster-admin",
		TimePeriod: "1h",
		Cluster:    "prod",
	}

	_, err := coordination.NewRequest(ar)
	if err != nil {
		panic("create new request error: " + err.Error())
	}
}

func demo() {
	ar := models.AccessRequest{
		Username:   "user",
		Email:      "user@example.com",
		Role:       "cluster-admin",
		TimePeriod: "1h",
		Cluster:    "prod",
	}

	id, err := coordination.NewRequest(ar)
	if err != nil {
		panic("create new request error: " + err.Error())
	}
	fmt.Println("request id: ", id)
	statuses, err := coordination.GetStatuses(id)
	if err != nil {
		panic("get statuses error: " + err.Error())
	}
	fmt.Println("request statuses after creating: ", statuses)
	requests, err := coordination.GetAllRequests()
	if err != nil {
		panic("get statuses error: " + err.Error())
	}
	fmt.Println("requests list", requests)

	stages := []string{helpers.StatusConnected, helpers.StatusReady, helpers.StatusApproved}
	users := []string{"kitten", "pony"}
	for _, stage := range stages {
		for _, user := range users {
			coordination.ChangeStatus(id, user, stage)
		}
		statuses, err := coordination.GetStatuses(id)
		if err != nil {
			panic("get statuses error: " + err.Error())
		}
		fmt.Println("request statuses after ", stage, statuses)
	}
	statuses, err = coordination.GetStatuses(id)
	if err != nil {
		panic("get statuses error: " + err.Error())
	}
	fmt.Println("final statuses", statuses)
}

func main() {
	generateRequests()
}
