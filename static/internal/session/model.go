package session

import (
	"github.com/gin-gonic/gin"
)

type Session struct {
	TenantID int
}

func FromContext(ctx *gin.Context) *Session {
	// TODO:
	return &Session{
		TenantID: 1,
	}
}

const sessionKey = "TODOIamASessionKeyChangeME"

func ToContext(ctx *gin.Context, s *Session) {
	ctx.Set(sessionKey, s)
}
