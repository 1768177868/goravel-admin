package seeders

import (
	"github.com/goravel/framework/facades"

	"goravel/app/models"
)

type PositionSeeder struct {
}

func (s *PositionSeeder) Signature() string {
	return "PositionSeeder"
}

func (s *PositionSeeder) Run() error {
	items := []models.Position{
		{Name: "总经理", Code: "GM", Status: 1, Sort: 1},
		{Name: "部门经理", Code: "MGR", Status: 1, Sort: 2},
		{Name: "普通员工", Code: "STAFF", Status: 1, Sort: 3},
	}
	for i := range items {
		p := items[i]
		exists, _ := facades.Orm().Query().Model(&models.Position{}).Where("code", p.Code).Exists()
		if exists {
			continue
		}
		if err := facades.Orm().Query().Create(&p); err != nil {
			return err
		}
	}
	return nil
}
