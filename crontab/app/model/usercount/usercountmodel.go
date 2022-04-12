package usercount

import (
    "database/sql"
    "errors"
    "fmt"
    "minicode.com/sirius/go-back-server/config/cfgstatus"
    "strings"
    "time"

    "github.com/zeromicro/go-zero/core/stores/builder"
    "github.com/zeromicro/go-zero/core/stores/sqlc"
    "github.com/zeromicro/go-zero/core/stores/sqlx"
    "github.com/zeromicro/go-zero/core/stringx"
)

var (
    userCountFieldNames          = builder.RawFieldNames(&UserCount{})
    userCountRows                = strings.Join(userCountFieldNames, ",")
    userCountRowsExpectAutoSet   = strings.Join(stringx.Remove(userCountFieldNames, "`id`", "`created_at`", "`updated_at`"), ",")
    userCountRowsWithPlaceHolder = strings.Join(stringx.Remove(userCountFieldNames, "`id`", "`created_at`", "`updated_at`"), "=?,") + "=?"
)

type (
    UserCountModel interface {
        Insert(data *UserCount) (sql.Result, error)
        FindOne(id int64) (*UserCount, error)
        FindOneByUin(uin string) (*UserCount, error)
        Update(data *UserCount) error
        UpdateNum(data *UserCount) error
        UpdateNumIncr(uin string, upType int8, incrType int8) error
        Delete(id int64) error
    }

    defaultUserCountModel struct {
        conn  sqlx.SqlConn
        table string
    }

    UserCount struct {
        Id        int64        `db:"id"`
        AppId     int64        `db:"app_id"`     // app标识
        Uin       string       `db:"uin"`        // 用户id
        FollowNum int64        `db:"follow_num"` // 粉丝数量
        FocusNum  int64        `db:"focus_num"`  // 关注数量
        CreatedAt time.Time    `db:"created_at"`
        UpdatedAt time.Time    `db:"updated_at"`
        DeletedAt sql.NullTime `db:"deleted_at"`
    }
)

func NewUserCountModel(conn sqlx.SqlConn) UserCountModel {
    return &defaultUserCountModel{
        conn:  conn,
        table: "`user_count`",
    }
}

func (m *defaultUserCountModel) Insert(data *UserCount) (sql.Result, error) {
    query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?)", m.table, userCountRowsExpectAutoSet)
    ret, err := m.conn.Exec(query, data.AppId, data.Uin, data.FollowNum, data.FocusNum, data.DeletedAt)
    return ret, err
}

func (m *defaultUserCountModel) FindOne(id int64) (*UserCount, error) {
    query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", userCountRows, m.table)
    var resp UserCount
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

func (m *defaultUserCountModel) FindOneByUin(uin string) (*UserCount, error) {
    query := fmt.Sprintf("select %s from %s where `uin` = ? limit 1", userCountRows, m.table)
    var resp UserCount
    err := m.conn.QueryRow(&resp, query, uin)
    switch err {
    case nil:
        return &resp, nil
    case sqlc.ErrNotFound:
        return &resp, nil
    default:
        return nil, err
    }
}

func (m *defaultUserCountModel) Update(data *UserCount) error {
    query := fmt.Sprintf("update %s set %s where `id` = ?", m.table, userCountRowsWithPlaceHolder)
    _, err := m.conn.Exec(query, data.AppId, data.Uin, data.FollowNum, data.FocusNum, data.DeletedAt, data.Id)
    return err
}

func (m *defaultUserCountModel) UpdateNum(data *UserCount) error {
    var separator string
    var setStr string
    if data.FollowNum > 0 {
        setStr = fmt.Sprintf("follow_num = %d", data.FollowNum)
        separator = ","
    }
    if data.FocusNum > 0 {
        setStr += fmt.Sprintf(" %s focus_num = %d", separator, data.FocusNum)
    }

    query := fmt.Sprintf("update %s set %s where `id` = ?", m.table, setStr)
    _, err := m.conn.Exec(query, data.Id)
    return err
}

func (m *defaultUserCountModel) UpdateNumIncr(uin string, upType int8, incrType int8) error {
    var separator string
    var setStr string
    var incrStr string

    switch incrType {
    case cfgstatus.UserBehaviorOperationAddType:
        incrStr = "+"
    case cfgstatus.UserBehaviorOperationReduceType:
        incrStr = "-"
    default:
        return errors.New("参数校验失败")
    }

    if upType == 1 {
        setStr += fmt.Sprintf(" %s focus_num = focus_num %s 1 ", separator, incrStr)
        separator = ","
    }
    if upType == 2 {
        setStr = fmt.Sprintf("follow_num = follow_num %s 1", incrStr)
    }

    query := fmt.Sprintf("update %s set %s where `uin` = ?", m.table, setStr)
    _, err := m.conn.Exec(query, uin)
    return err
}

func (m *defaultUserCountModel) Delete(id int64) error {
    query := fmt.Sprintf("delete from %s where `id` = ?", m.table)
    _, err := m.conn.Exec(query, id)
    return err
}
