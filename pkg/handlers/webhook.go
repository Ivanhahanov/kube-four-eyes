package handlers

import (
	"fmt"
	"time"
	"webhook/pkg/coordination"
	"webhook/pkg/storage"
	"webhook/pkg/ws"

	"github.com/gofiber/fiber/v2"
	authorizationv1 "k8s.io/api/authorization/v1"
)

func WriteUserLog(id int64, req authorizationv1.SubjectAccessReview) {
	action := fmt.Sprintf("%s: %s %s %s",
		req.Spec.User,
		req.Spec.ResourceAttributes.Namespace,
		req.Spec.ResourceAttributes.Verb,
		req.Spec.ResourceAttributes.Resource,
	)
	storage.DB().Put(fmt.Sprintf("log/%d", time.Now().Unix()), action, id)
	ws.Ch.Broadcast <- action
}

func Authorize(ctx *fiber.Ctx) error {
	var req authorizationv1.SubjectAccessReview
	ctx.BodyParser(&req)
	if id, ok := coordination.CheckUserAccess(req.Spec.User); ok {
		WriteUserLog(id, req)
		req.Status.Allowed = true

	} else {
		req.Status.Allowed = false
		req.Status.Denied = true
		req.Status.Reason = "Access not granted"
	}

	return ctx.JSON(req)
}
