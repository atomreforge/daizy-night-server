package service

type ServiceComplex struct {
	ServiceUser         *ServiceUser
	ServiceCode         *ServiceCode
	ServiceAdmin        *ServiceAdmin
	ServiceHealth       *ServiceHealth
	ServiceNotification *ServiceNotification
}

func NewServiceComplex(
	pSvcUser *ServiceUser,
	pSvcCode *ServiceCode,
	pSvcAdmin *ServiceAdmin,
	pSvcHealth *ServiceHealth,
	pSvcNotification *ServiceNotification,
) *ServiceComplex {
	return &ServiceComplex{
		ServiceUser:         pSvcUser,
		ServiceCode:         pSvcCode,
		ServiceAdmin:        pSvcAdmin,
		ServiceHealth:       pSvcHealth,
		ServiceNotification: pSvcNotification,
	}
}
