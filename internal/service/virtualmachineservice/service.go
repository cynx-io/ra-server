package virtualmachineservice

import (
	"github.com/cynxees/ra-server/internal/repository/database"
)

type Service struct {
	VirtualMachineRepo *database.VirtualMachineRepo
}
