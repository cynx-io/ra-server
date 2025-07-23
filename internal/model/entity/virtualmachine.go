package entity

import (
	"github.com/cynxees/cynx-core/src/entity"
	pb "github.com/cynxees/ra-server/api/proto/gen/ra"
)

type VirtualMachine struct {
	entity.EssentialEntity
	Name        string `gorm:"column:name;not null" json:"name"`
	Description string `gorm:"column:description" json:"description"`
	Status      string `gorm:"column:status;default:'inactive'" json:"status"`
	Type        string `gorm:"column:type;not null" json:"type"`
	Resources   string `gorm:"column:resources;type:text" json:"resources"`
	IPAddress   string `gorm:"column:ip_address" json:"ip_address"`
	UserID      int32  `gorm:"column:user_id;not null" json:"user_id"`
	Port        int32  `gorm:"column:port" json:"port"`
}

func (vm VirtualMachine) Response() *pb.VirtualMachine {
	return &pb.VirtualMachine{
		Id:          vm.Id,
		Name:        vm.Name,
		Description: vm.Description,
		Status:      vm.Status,
		Type:        vm.Type,
		Resources:   vm.Resources,
		UserId:      vm.UserID,
		IpAddress:   vm.IPAddress,
		Port:        vm.Port,
	}
}
