package graph

import (
	"fmt"
	"github.com/go-pg/pg/v9"
	"github.com/vickywane/api/graph/model"
)

type Resolver struct{
	DB *pg.DB
}

func (r *mutationResolver) GetUserByField(field, value string) (*model.User, error) {
	user := model.User{}

	err := r.DB.Model(&user).Where(fmt.Sprintf("%v = ?", field), value).First()

	return &user, err
}

func (r *mutationResolver) UpdateUser(user *model.User) (*model.User, error) {
	_, err := r.DB.Model(user).Where("id = ?", user.ID).Update()
	return user, err
}