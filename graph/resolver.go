package graph

import (
	"fmt"
	"github.com/go-pg/pg/v9"
	"github.com/vickywane/api/graph/model"
)

type Resolver struct{
	DB *pg.DB
}

func (r *mutationResolver) GetUserField(field, value string) (*model.User, error) {
	user := model.User{}

	err := r.DB.Model(&user).Where(fmt.Sprintf("%v = ?", field), value).First()

	return &user, err
}