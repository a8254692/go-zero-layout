package userfocusmsglog

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
    userFocusMsgLogFieldNames          = builder.RawFieldNames(&UserFocusMsgLog{})
    userFocusMsgLogRows                = strings.Join(userFocusMsgLogFieldNames, ",")
    userFocusMsgLogRowsExpectAutoSet   = strings.Join(stringx.Remove(userFocusMsgLogFieldNames, "`id`", "`created_at`", "`updated_at`"), ",")
    userFocusMsgLogRowsWithPlaceHolder = strings.Join(stringx.Remove(userFocusMsgLogFieldNames, "`id`", "`created_at`", "`updated_at`"), "=?,") + "=?"
)

type (
    UserFocusMsgLogModel interface {
        Insert(data *UserFocusMsgLog) (sql.Result, error)
        FindOne(id int64) (*UserFocusMsgLog, error)
        FindOneByUinFocusUin(uin string, focusUin string) (*UserFocusMsgLog, error)
        Update(data *UserFocusMsgLog) error
        Delete(id int64) error
    }

    defaultUserFocusMsgLogModel struct {
        conn  sqlx.SqlConn
        table string
    }

    UserFocusMsgLog struct {
        Id        int64        `db:"id"`
        AppId     int64        `db:"app_id"`    // app标识
        Uin       string       `db:"uin"`       // 用户id
        FocusUin  string       `db:"focus_uin"` // 被关注者id
        CreatedAt time.Time    `db:"created_at"`
        UpdatedAt time.Time    `db:"updated_at"`
        DeletedAt sql.NullTime `db:"deleted_at"`
    }
)

func NewUserFocusMsgLogModel(conn sqlx.SqlConn) UserFocusMsgLogModel {
    return &defaultUserFocusMsgLogModel{
        conn:  conn,
        table: "`user_focus_msg_log`",
    }
}

func (m *defaultUserFocusMsgLogModel) Insert(data *UserFocusMsgLog) (sql.Result, error) {
    query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?)", m.table, userFocusMsgLogRowsExpectAutoSet)
    ret, err := m.conn.Exec(query, data.AppId, data.Uin, data.FocusUin, data.DeletedAt)
    return ret, err
}

func (m *defaultUserFocusMsgLogModel) FindOne(id int64) (*UserFocusMsgLog, error) {
    query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", userFocusMsgLogRows, m.table)
    var resp UserFocusMsgLog
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

func (m *defaultUserFocusMsgLogModel) FindOneByUinFocusUin(uin string, focusUin string) (*UserFocusMsgLog, error) {
    var resp UserFocusMsgLog
    query := fmt.Sprintf("select %s from %s where `uin` = ? and `focus_uin` = ? limit 1", userFocusMsgLogRows, m.table)
    err := m.conn.QueryRow(&resp, query, uin, focusUin)
    switch err {
    case nil:
        return &resp, nil
    case sqlc.ErrNotFound:
        return &resp, nil
    default:
        return nil, err
    }
}

func (m *defaultUserFocusMsgLogModel) Update(data *UserFocusMsgLog) error {
    query := fmt.Sprintf("update %s set %s where `id` = ?", m.table, userFocusMsgLogRowsWithPlaceHolder)
    _, err := m.conn.Exec(query, data.AppId, data.Uin, data.FocusUin, data.DeletedAt, data.Id)
    return err
}

func (m *defaultUserFocusMsgLogModel) Delete(id int64) error {
    query := fmt.Sprintf("delete from %s where `id` = ?", m.table)
    _, err := m.conn.Exec(query, id)
    return err
}
