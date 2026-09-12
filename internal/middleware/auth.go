package middleware

import (
	"crypto/ed25519"

	abstract "github.com/atomreforge/daizy-night-server/internal/abstract/interface"
	"github.com/atomreforge/daizy-night-server/internal/consts"
	"github.com/atomreforge/daizy-night-server/internal/crypto"
	"github.com/atomreforge/daizy-night-server/internal/model"
	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v5"
	"github.com/labstack/echo/v5"
)

func AuthenJWT(p abstract.InterfaceCrypto) echo.MiddlewareFunc {
	return echojwt.WithConfig(echojwt.Config{
		ContextKey: string(consts.CtxExprKeyJWT),
		KeyFunc:    crypto.NewJWTKeyFunc(p.GetJwtAccessTokenDeckey().(ed25519.PublicKey)),
		NewClaimsFunc: func(c *echo.Context) jwt.Claims {
			return &model.JwtAccessTokenPayload{}
		},
	})
}

func logPrintError(ctx *echo.Context, err error) string {
	return ("error::" + err.Error() +
		"method::" + ctx.Request().Method +
		"path::" + ctx.Request().URL.Path)
}
