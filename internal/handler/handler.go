package handler

// note that validation ops or so should be executed in handler layer to keep the service layer clean and reusable.
// "validate (like format or null value) in handler and verify (like whether a regcode is legal) in service"

import (
	abstract "github.com/atomreforge/daizy-night-server/internal/abstract/interface"
)

type InterfaceHandlerComplex interface {
	NewHandlerComplex(
		svcUser abstract.InterfaceServiceUser,
		svcCode abstract.InterfaceServiceCode,
		svcAdmin abstract.InterfaceServiceAdmin,
		svcHealth abstract.InterfaceServiceHealth,
		svcNotification abstract.InterfaceServiceNotification,
	) *HandlerComplex
}

// HandleXXX() is func of HandlerComplex
type HandlerComplex struct {
	ServiceUser         abstract.InterfaceServiceUser
	ServiceCode         abstract.InterfaceServiceCode
	ServiceAdmin        abstract.InterfaceServiceAdmin
	ServiceHealth       abstract.InterfaceServiceHealth
	serviceNotification abstract.InterfaceServiceNotification
}

func NewHandlerComplex(
	svcUser abstract.InterfaceServiceUser,
	svcCode abstract.InterfaceServiceCode,
	svcAdmin abstract.InterfaceServiceAdmin,
	svcHealth abstract.InterfaceServiceHealth,
	svcNotification abstract.InterfaceServiceNotification,
) *HandlerComplex {
	return &HandlerComplex{
		ServiceUser:         svcUser,
		ServiceCode:         svcCode,
		ServiceAdmin:        svcAdmin,
		ServiceHealth:       svcHealth,
		serviceNotification: svcNotification,
	}
}
