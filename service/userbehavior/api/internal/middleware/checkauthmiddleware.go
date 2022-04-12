package middleware

import (
    "net/http"

    "minicode.com/sirius/go-back-server/service/userbehavior/api/internal/config"
    "minicode.com/sirius/go-back-server/utils/middleware"
)

type CheckAuthMiddleware struct {
    Config config.Config
}

func NewCheckAuthMiddleware(c config.Config) *CheckAuthMiddleware {
    return &CheckAuthMiddleware{
        Config: c,
    }
}

func (m *CheckAuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
    return middleware.GetCheckAuthFun(m.Config.Mode, next)
}
