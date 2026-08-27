package authctx

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const principalKey = "auth_principal"

type Principal struct {
	UserID uuid.UUID
	Email  string
	Role   string
	JTI    string
}

func Set(c *gin.Context, p Principal) {
	c.Set(principalKey, p)
}

func Get(c *gin.Context) (Principal, bool) {
	v, ok := c.Get(principalKey)
	if !ok {
		return Principal{}, false
	}
	p, ok := v.(Principal)
	return p, ok
}

func MustGet(c *gin.Context) Principal {
	p, ok := Get(c)
	if !ok {
		panic("auth principal missing from context")
	}
	return p
}
