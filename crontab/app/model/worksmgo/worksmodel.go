package worksmgo

import (
	"context"
	"github.com/zeromicro/go-zero/core/stores/mongo"
)

type WorksModel interface {
	FindOne(ctx context.Context, id string) (*Works, error)
	FindList(limit int64, offset int64) (*[]Works, error)
}

type defaultWorksModel struct {
	*mongo.Model
}

func NewWorksModel(url, collection string) WorksModel {
	return &defaultWorksModel{
		Model: mongo.MustNewModel(url, collection),
	}
}

func (m *defaultWorksModel) FindOne(ctx context.Context, id string) (*Works, error) {
	session, err := m.TakeSession()
	if err != nil {
		return nil, err
	}

	defer m.PutSession(session)
	var data Works

	err = m.GetCollection(session).FindId(id).One(&data)
	switch err {
	case nil:
		return &data, nil
	case mongo.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *defaultWorksModel) FindList(limit int64, offset int64) (*[]Works, error) {
	session, err := m.TakeSession()
	if err != nil {
		return nil, err
	}

	defer m.PutSession(session)
	var data []Works

	err = m.GetCollection(session).Find(nil).Skip(int(offset)).Limit(int(limit)).All(&data)
	switch err {
	case nil:
		return &data, nil
	case mongo.ErrNotFound:
		return &data, nil
	default:
		return nil, err
	}
}
