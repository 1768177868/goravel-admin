package requests

import (
	"goravel/app/models"

	"github.com/goravel/framework/contracts/http"
)

type UserCreate struct {
	Username string           `form:"username" json:"username"`
	Password string           `form:"password" json:"password"`
	Name     string           `form:"name" json:"name"`
	Avatar   string           `form:"avatar" json:"avatar"`
	Alias    string           `form:"alias" json:"alias"`
	Mail     string           `form:"mail" json:"mail"`
	Status   uint8            `form:"status" json:"status"`
	Tags     []models.UserTag `form:"tags" json:"tags"`
}

func (r *UserCreate) Authorize(ctx http.Context) error {
	return nil
}

func (r *UserCreate) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"username": "required|min_len:3|max_len:50",
		"password": "required|min_len:6",
		"name":     "required",
		"mail":     "email",
		"status":   "in:0,1",
	}
}
