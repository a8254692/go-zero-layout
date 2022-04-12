package producecount

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/builder"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/core/stringx"
)

var (
	produceCountFieldNames          = builder.RawFieldNames(&ProduceCount{})
	produceCountRows                = strings.Join(produceCountFieldNames, ",")
	produceCountRowsExpectAutoSet   = strings.Join(stringx.Remove(produceCountFieldNames, "`id`", "`created_at`", "`updated_at`"), ",")
	produceCountRowsWithPlaceHolder = strings.Join(stringx.Remove(produceCountFieldNames, "`id`", "`created_at`", "`updated_at`"), "=?,") + "=?"
)

type (
	ProduceCountModel interface {
		Insert(data *ProduceCount) (sql.Result, error)
		FindOne(id int64) (*ProduceCount, error)
		FindOneByParam(topicType int64, topicId string) (*ProduceCount, error)
		FindListByParam(topicType int64, topicIds []string) (*[]ProduceCount, error)
		Update(data *ProduceCount) error
		Delete(id int64) error
	}

	defaultProduceCountModel struct {
		conn  sqlx.SqlConn
		table string
	}

	ProduceCount struct {
		Id         int64        `db:"id"`
		AppId      int64        `db:"app_id"`      // app标识
		TopicId    string       `db:"topic_id"`    // 主题id
		TopicType  int64        `db:"topic_type"`  // 主题类型
		CommentNum int64        `db:"comment_num"` // 评论数量
		PraiseNum  int64        `db:"praise_num"`  // 点赞数量
		ShareNum   int64        `db:"share_num"`   // 分享数量
		UpdatedAt  time.Time    `db:"updated_at"`
		DeletedAt  sql.NullTime `db:"deleted_at"`
		CreatedAt  time.Time    `db:"created_at"`
	}
)

func NewProduceCountModel(conn sqlx.SqlConn) ProduceCountModel {
	return &defaultProduceCountModel{
		conn:  conn,
		table: "`produce_count`",
	}
}

func (m *defaultProduceCountModel) Insert(data *ProduceCount) (sql.Result, error) {
	query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?, ?, ?)", m.table, produceCountRowsExpectAutoSet)
	ret, err := m.conn.Exec(query, data.AppId, data.TopicId, data.TopicType, data.CommentNum, data.PraiseNum, data.ShareNum, data.DeletedAt)
	return ret, err
}

func (m *defaultProduceCountModel) FindOne(id int64) (*ProduceCount, error) {
	query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", produceCountRows, m.table)
	var resp ProduceCount
	err := m.conn.QueryRow(&resp, query, id)
	switch err {
	case nil:
		return &resp, nil
	case sqlc.ErrNotFound:
		return &resp, nil
	default:
		return nil, err
	}
}

func (m *defaultProduceCountModel) FindOneByParam(topicType int64, topicId string) (*ProduceCount, error) {
	query := fmt.Sprintf("select %s from %s where `topic_type` = ? and `topic_id` = ? limit 1", produceCountRows, m.table)
	var resp ProduceCount
	err := m.conn.QueryRow(&resp, query, topicType, topicId)
	switch err {
	case nil:
		return &resp, nil
	case sqlc.ErrNotFound:
		return &resp, nil
	default:
		return nil, err
	}
}

func (m *defaultProduceCountModel) FindListByParam(topicType int64, topicIds []string) (*[]ProduceCount, error) {
	topicIdStr := strings.Join(topicIds, "','")

	query := fmt.Sprintf("select %s from %s where `topic_type` = ? and `topic_id` in ('%s') ", produceCountRows, m.table, topicIdStr)
	var resp []ProduceCount
	err := m.conn.QueryRows(&resp, query, topicType)
	switch err {
	case nil:
		return &resp, nil
	case sqlc.ErrNotFound:
		return &resp, nil
	default:
		return nil, err
	}
}

func (m *defaultProduceCountModel) Update(data *ProduceCount) error {
	query := fmt.Sprintf("update %s set %s where `id` = ?", m.table, produceCountRowsWithPlaceHolder)
	_, err := m.conn.Exec(query, data.AppId, data.TopicId, data.TopicType, data.CommentNum, data.PraiseNum, data.ShareNum, data.DeletedAt, data.Id)
	return err
}

func (m *defaultProduceCountModel) Delete(id int64) error {
	query := fmt.Sprintf("delete from %s where `id` = ?", m.table)
	_, err := m.conn.Exec(query, id)
	return err
}
