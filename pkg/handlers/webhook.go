package handlers

import (
	"webhook/pkg/coordination"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	authorizationv1 "k8s.io/api/authorization/v1"
)

func Authorize(ctx *fiber.Ctx) error {
	var req authorizationv1.SubjectAccessReview
	ctx.BodyParser(&req)
	log.Debug(req.Spec.User)
	if coordination.CheckUserAccess(req.Spec.User) {
		req.Status.Allowed = true
	} else {
		req.Status.Allowed = false
		req.Status.Denied = true
		req.Status.Reason = "Access not granted"
	}

	return ctx.JSON(req)
}
