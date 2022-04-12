package specialmgo

import (
	"context"
	"fmt"
	"github.com/globalsign/mgo/bson"
	"github.com/zeromicro/go-zero/core/stores/mongo"
)

type SpecialModel interface {
	FindOne(ctx context.Context, id string) (*Special, error)
}

type defaultSpecialModel struct {
	*mongo.Model
}

func NewSpecialModel(url, collection string) SpecialModel {
	return &defaultSpecialModel{
		Model: mongo.MustNewModel(url, collection),
	}
}

func (m *defaultSpecialModel) FindOne(ctx context.Context, id string) (*Special, error) {
	session, err := m.TakeSession()
	if err != nil {
		return nil, err
	}
	oid := bson.ObjectIdHex(id) // 获取 objectId

	defer m.PutSession(session)
	var data Special
	err = m.GetCollection(session).FindId(oid).One(&data)
	fmt.Printf("special.FindOne.id %s ,err %v \n", id, err)
	switch err {
	case nil:
		return &data, nil
	case mongo.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}
