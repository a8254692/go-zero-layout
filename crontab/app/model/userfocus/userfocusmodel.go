package userfocus

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
    userFocusFieldNames          = builder.RawFieldNames(&UserFocus{})
    userFocusRows                = strings.Join(userFocusFieldNames, ",")
    userFocusRowsExpectAutoSet   = strings.Join(stringx.Remove(userFocusFieldNames, "`id`", "`created_at`", "`updated_at`"), ",")
    userFocusRowsWithPlaceHolder = strings.Join(stringx.Remove(userFocusFieldNames, "`id`", "`created_at`", "`updated_at`"), "=?,") + "=?"
)

type (
    UserFocusModel interface {
        Insert(data *UserFocus) (sql.Result, error)
        FindOne(id int64) (*UserFocus, error)
        FindOneByUinFocusUin(uin string, focusUin string) (*UserFocus, error)
        FindUinFocusList(uin string, limit int64, offset int64) (*[]UserFocus, error)
        FindUinFollowList(uin string, limit int64, offset int64) (*[]UserFocus, error)
        FindUinFocusCount(uin string) (int64, error)
        FindUinFollowCount(uin string) (int64, error)
        Update(data *UserFocus) error
        UpdateStatus(uin string, focusUin string, status int64) error
        Delete(id int64) error
        DeleteByUinFocusUin(uin string, focusUin string) error
    }

    defaultUserFocusModel struct {
        conn  sqlx.SqlConn
        table string
    }

    UserFocus struct {
        Id        int64        `db:"id"`
        AppId     int64        `db:"app_id"`    // app标识
        Uin       string       `db:"uin"`       // 用户id
        FocusUin  string       `db:"focus_uin"` // 被关注者id
        Status    int64        `db:"status"`    // 互相关注状态 1-单向关注 2-双向关注
        CreatedAt time.Time    `db:"created_at"`
        UpdatedAt time.Time    `db:"updated_at"`
        DeletedAt sql.NullTime `db:"deleted_at"`
    }
)

func NewUserFocusModel(conn sqlx.SqlConn) UserFocusModel {
    return &defaultUserFocusModel{
        conn:  conn,
        table: "`user_focus`",
    }
}

func (m *defaultUserFocusModel) Insert(data *UserFocus) (sql.Result, error) {
    query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?)", m.table, userFocusRowsExpectAutoSet)
    ret, err := m.conn.Exec(query, data.AppId, data.Uin, data.FocusUin, data.Status, data.DeletedAt)
    return ret, err
}

func (m *defaultUserFocusModel) FindOne(id int64) (*UserFocus, error) {
    query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", userFocusRows, m.table)
    var resp UserFocus
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

func (m *defaultUserFocusModel) FindOneByUinFocusUin(uin string, focusUin string) (*UserFocus, error) {
    var resp UserFocus
    query := fmt.Sprintf("select %s from %s where `uin` = ? and `focus_uin` = ? limit 1", userFocusRows, m.table)
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

func (m *defaultUserFocusModel) FindUinFocusList(uin string, limit int64, offset int64) (*[]UserFocus, error) {
    query := fmt.Sprintf("select %s from %s where `uin` = ? order by created_at desc limit %d offset %d ", userFocusRows, m.table, limit, offset)
    var resp []UserFocus
    err := m.conn.QueryRows(&resp, query, uin)
    switch err {
    case nil:
        return &resp, nil
    case sqlc.ErrNotFound:
        return &resp, nil
    default:
        return nil, err
    }
}

func (m *defaultUserFocusModel) FindUinFollowList(uin string, limit int64, offset int64) (*[]UserFocus, error) {
    query := fmt.Sprintf("select %s from %s where `focus_uin` = ? order by created_at desc limit %d offset %d ", userFocusRows, m.table, limit, offset)
    var resp []UserFocus
    err := m.conn.QueryRows(&resp, query, uin)
    switch err {
    case nil:
        return &resp, nil
    case sqlc.ErrNotFound:
        return &resp, nil
    default:
        return nil, err
    }
}

func (m *defaultUserFocusModel) FindUinFocusCount(uin string) (int64, error) {
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

func (m *defaultUserFocusModel) FindUinFollowCount(uin string) (int64, error) {
    query := fmt.Sprintf("select count(*) as num from %s where `focus_uin` = ? limit 1", m.table)
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

func (m *defaultUserFocusModel) Update(data *UserFocus) error {
    query := fmt.Sprintf("update %s set %s where `id` = ?", m.table, userFocusRowsWithPlaceHolder)
    _, err := m.conn.Exec(query, data.AppId, data.Uin, data.FocusUin, data.Status, data.DeletedAt, data.Id)
    return err
}

func (m *defaultUserFocusModel) UpdateStatus(uin string, focusUin string, status int64) error {
    query := fmt.Sprintf("update %s set status = ? where `uin` = ? and `focus_uin` = ? ", m.table)
    _, err := m.conn.Exec(query, status, uin, focusUin)
    return err
}

func (m *defaultUserFocusModel) Delete(id int64) error {
    query := fmt.Sprintf("delete from %s where `id` = ?", m.table)
    _, err := m.conn.Exec(query, id)
    return err
}

func (m *defaultUserFocusModel) DeleteByUinFocusUin(uin string, focusUin string) error {
    query := fmt.Sprintf("delete from %s where `uin` = ? and `focus_uin` = ? ", m.table)
    _, err := m.conn.Exec(query, uin, focusUin)
    return err
}
