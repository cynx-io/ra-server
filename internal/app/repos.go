package app

import (
	"github.com/cynxees/ra-server/internal/repository/database"
)

type Repos struct {
	VirtualMachineRepo *database.VirtualMachineRepo
}

func NewRepos(dependencies *Dependencies) *Repos {
	return &Repos{
		VirtualMachineRepo: database.NewVirtualMachineRepo(dependencies.DatabaseClient.DB),
	}
}
