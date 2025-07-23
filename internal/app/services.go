package app

import (
	"github.com/cynxees/ra-server/internal/service/virtualmachineservice"
)

type Services struct {
	VirtualMachineService *virtualmachineservice.Service
}

func NewServices(repos *Repos) *Services {
	return &Services{
		VirtualMachineService: &virtualmachineservice.Service{
			VirtualMachineRepo: repos.VirtualMachineRepo,
		},
	}
}
