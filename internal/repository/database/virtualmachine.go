package database

import (
	"context"

	"github.com/cynxees/ra-server/internal/model/entity"
	"gorm.io/gorm"
)

type VirtualMachineRepo struct {
	DB *gorm.DB
}

func NewVirtualMachineRepo(db *gorm.DB) *VirtualMachineRepo {
	return &VirtualMachineRepo{DB: db}
}

func (r *VirtualMachineRepo) Get(ctx context.Context, id int32) (*entity.VirtualMachine, error) {
	var vm entity.VirtualMachine
	err := r.DB.WithContext(ctx).First(&vm, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &vm, nil
}
