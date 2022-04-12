package userpraise

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
    userPraiseFieldNames          = builder.RawFieldNames(&UserPraise{})
    userPraiseRows                = strings.Join(userPraiseFieldNames, ",")
    userPraiseRowsExpectAutoSet   = strings.Join(stringx.Remove(userPraiseFieldNames, "`id`", "`created_at`", "`updated_at`"), ",")
    userPraiseRowsWithPlaceHolder = strings.Join(stringx.Remove(userPraiseFieldNames, "`id`", "`created_at`", "`updated_at`"), "=?,") + "=?"
)

type (
    UserPraiseModel interface {
        Insert(data *UserPraise) (sql.Result, error)
        FindOne(id int64) (*UserPraise, error)
        FindOneByParam(uin string, topicType int64, topicId string) (*UserPraise, error)
        FindUinList(uin string, limit int64, offset int64) (*[]UserPraise, error)
        FindUinCount(uin string) (int64, error)
        Update(data *UserPraise) error
        Delete(id int64) error
        DeleteByParam(uin string, topicType int64, topicId string) error
    }

    defaultUserPraiseModel struct {
        conn  sqlx.SqlConn
        table string
    }

    UserPraise struct {
        Id        int64        `db:"id"`
        AppId     int64        `db:"app_id"`     // app标识
        TopicId   string       `db:"topic_id"`   // 主题id
        TopicType int64        `db:"topic_type"` // 主题类型
        Uin       string       `db:"uin"`        // 用户id
        CreatedAt time.Time    `db:"created_at"`
        UpdatedAt time.Time    `db:"updated_at"`
        DeletedAt sql.NullTime `db:"deleted_at"`
    }
)

func NewUserPraiseModel(conn sqlx.SqlConn) UserPraiseModel {
    return &defaultUserPraiseModel{
        conn:  conn,
        table: "`user_praise`",
    }
}

func (m *defaultUserPraiseModel) Insert(data *UserPraise) (sql.Result, error) {
    query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?)", m.table, userPraiseRowsExpectAutoSet)
    ret, err := m.conn.Exec(query, data.AppId, data.TopicId, data.TopicType, data.Uin, data.DeletedAt)
    return ret, err
}

func (m *defaultUserPraiseModel) FindOne(id int64) (*UserPraise, error) {
    query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", userPraiseRows, m.table)
    var resp UserPraise
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

func (m *defaultUserPraiseModel) FindOneByParam(uin string, topicType int64, topicId string) (*UserPraise, error) {
    query := fmt.Sprintf("select %s from %s where `uin` = ? and `topic_type` = ? and `topic_id` = ? limit 1", userPraiseRows, m.table)
    var resp UserPraise
    err := m.conn.QueryRow(&resp, query, uin, topicType, topicId)
    switch err {
    case nil:
        return &resp, nil
    case sqlc.ErrNotFound:
        return &resp, nil
    default:
        return nil, err
    }
}

func (m *defaultUserPraiseModel) FindUinList(uin string, limit int64, offset int64) (*[]UserPraise, error) {
    query := fmt.Sprintf("select %s from %s where `uin` = ? and topic_type in (1,2) order by created_at desc limit %d offset %d ", userPraiseRows, m.table, limit, offset)
    var resp []UserPraise
    err := m.conn.QueryRows(&resp, query, uin)
    switch err {
    case nil:
        return &resp, nil
    case sqlc.ErrNotFound:
        return nil, nil
    default:
        return nil, err
    }
}

func (m *defaultUserPraiseModel) FindUinCount(uin string) (int64, error) {
    query := fmt.Sprintf("select count(*) as num from %s where `uin` = ? limit 1", m.table)
    var resp int64
    err := m.conn.QueryRow(&resp, query, uin)
    switch err {
    case nil:
        return resp, nil
    case sqlc.ErrNotFound:
        return resp, nil
    default:
        return 0, err
    }
}

func (m *defaultUserPraiseModel) Update(data *UserPraise) error {
    query := fmt.Sprintf("update %s set %s where `id` = ?", m.table, userPraiseRowsWithPlaceHolder)
    _, err := m.conn.Exec(query, data.AppId, data.TopicId, data.TopicType, data.Uin, data.DeletedAt, data.Id)
    return err
}

func (m *defaultUserPraiseModel) Delete(id int64) error {
    query := fmt.Sprintf("delete from %s where `id` = ?", m.table)
    _, err := m.conn.Exec(query, id)
    return err
}

func (m *defaultUserPraiseModel) DeleteByParam(uin string, topicType int64, topicId string) error {
    query := fmt.Sprintf("delete from %s where `uin` = ? and `topic_type` = ? and `topic_id` = ? ", m.table)
    _, err := m.conn.Exec(query, uin, topicType, topicId)
    return err
}
