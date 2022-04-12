package comment

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
	commentFieldNames          = builder.RawFieldNames(&Comment{})
	commentRows                = strings.Join(commentFieldNames, ",")
	commentRowsExpectAutoSet   = strings.Join(stringx.Remove(commentFieldNames, "`id`", "`create_time`", "`update_time`"), ",")
	commentRowsWithPlaceHolder = strings.Join(stringx.Remove(commentFieldNames, "`id`", "`create_time`", "`update_time`"), "=?,") + "=?"
)

type (
	CommentModel interface {
		Insert(data *Comment) (sql.Result, error)
		FindOne(id int64) (*Comment, error)
		FindCommentByUin(uin string ,appId int64 ,topicType int64, topicId string) (*[]Comment, error)
		FindCommentLatest(appId int64 ,topicType int64, topicId string, limit, offset int32) (*[]Comment, error)
		FindCommentHot(appId int64,topicType int64, topicId string, limit, offset int32) (*[]Comment, error)
		CommentCount(appId int64,topicType int64, topicId string) (int64,error)
		UserUnReviewCommentCount(uin string,appId int64,topicType int64, topicId string) (int64, error)
		Update(data *Comment) error
		Delete(id int64) error
	}

	defaultCommentModel struct {
		conn  sqlx.SqlConn
		table string
	}

	Comment struct {
		Id               int64          `db:"id"`
		AppId            int64          `db:"app_id"`             // app
		TopicId          string         `db:"topic_id"`           // 主题id
		TopicType        int64          `db:"topic_type"`         // 主题类型
		Content          sql.NullString `db:"content"`            // 评论内容
		Uin              string         `db:"uin"`                // 评论用户id
		AutoReviewStatus int64          `db:"auto_review_status"` // 自动审核状态 0 - 未审核 1 - 审核不通过  2 - 审核通过
		AutoReviewTime   sql.NullTime   `db:"auto_review_time"`   // 自动审核时间
		Reviewer         string         `db:"reviewer"`           // 审核人
		ReviewTime       sql.NullTime   `db:"review_time"`        // 审核时间
		ReviewStatus     int64          `db:"review_status"`      // 审核状态 0 - 未审核 1 - 审核未通过 2 - 审核通过
		SensitiveEntry   int64          `db:"sensitive_entry"`    // 违规条目
		PraiseCount      int64          `db:"praise_count"`       // 点赞数
		CreatedAt        time.Time      `db:"created_at"`
		UpdatedAt        time.Time      `db:"updated_at"`
		DeletedAt        sql.NullTime   `db:"deleted_at"`
	}
)

func NewCommentModel(conn sqlx.SqlConn) CommentModel {
	return &defaultCommentModel{
		conn:  conn,
		table: "`comment`",
	}
}

func (m *defaultCommentModel) Insert(data *Comment) (sql.Result, error) {
	fieldStr := "`app_id`,`topic_id`,`topic_type`,`content`,`uin`,`created_at`"
	query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?,?)", m.table, fieldStr)
	fmt.Printf("commentModel.insert data %+v ,unixTs %d \n",data,data.CreatedAt.Unix())
	ret, err := m.conn.Exec(query, data.AppId, data.TopicId, data.TopicType, data.Content.String, data.Uin, data.CreatedAt)
	return ret, err
}

func (m *defaultCommentModel) FindOne(id int64) (*Comment, error) {
	query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", commentRows, m.table)
	var resp Comment
	err := m.conn.QueryRow(&resp, query, id)
	switch err {
	case nil:
		return &resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *defaultCommentModel) FindCommentByUin (uin string ,appId int64,topicType int64, topicId string) (*[]Comment, error) {
	querySql := fmt.Sprintf("select %s from %s where uin = ? and  app_id = ? and topic_type =  ? and topic_id = ? and review_status = 0 and deleted_at is null order by created_at desc", commentRows, m.table)

	var resp []Comment
	err := m.conn.QueryRows(&resp, querySql,uin ,appId, topicType, topicId)
	switch err {
	case nil:
		return &resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *defaultCommentModel) FindCommentLatest(appId int64,topicType int64, topicId string, limit, offset int32) (*[]Comment, error) {
	querySql := fmt.Sprintf("select %s from %s where app_id = ? and topic_type =  ? and topic_id = ? and review_status = 2 and deleted_at is null order by created_at desc", commentRows, m.table)
	querySql = fmt.Sprintf("%s limit %d offset %d", querySql, limit, offset)

	var resp []Comment
	err := m.conn.QueryRows(&resp, querySql,appId, topicType, topicId)
	switch err {
	case nil:
		return &resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *defaultCommentModel) FindCommentHot(appId int64,topicType int64, topicId string, limit, offset int32) (*[]Comment, error) {
	querySql := fmt.Sprintf("select %s from %s where app_id = ? and topic_type =  ? and topic_id = ? and review_status = 2 and deleted_at is null order by praise_count,created_at desc", commentRows, m.table)
	querySql = fmt.Sprintf("%s limit %d offset %d", querySql, limit, offset)

	var resp []Comment
	err := m.conn.QueryRows(&resp, querySql,appId, topicType, topicId)
	switch err {
	case nil:
		return &resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}


func (m *defaultCommentModel) Update(data *Comment) error {
	query := fmt.Sprintf("update %s set %s where `id` = ?", m.table, commentRowsWithPlaceHolder)
	_, err := m.conn.Exec(query, data.AppId, data.TopicId, data.TopicType, data.Content, data.Uin, data.Reviewer, data.ReviewTime, data.ReviewStatus,  data.PraiseCount, data.CreatedAt, data.UpdatedAt, data.DeletedAt, data.Id)
	return err
}

func (m *defaultCommentModel) Delete(id int64) error {
	query := fmt.Sprintf("delete from %s where `id` = ?", m.table)
	_, err := m.conn.Exec(query, id)
	return err
}

func (m *defaultCommentModel) CommentCount(appId int64,topicType int64, topicId string) (int64, error) {
	query := fmt.Sprintf("select count(*) as recCount from %s where app_id = ? and topic_type =  ? and topic_id = ? and review_status = 2 and deleted_at is null ",  m.table)
	var resp int64
	err := m.conn.QueryRow(&resp, query,appId,topicType,topicId)

	if err != nil && err != sqlc.ErrNotFound {
		return resp,err
	}

	return resp,nil
}


func (m *defaultCommentModel) UserUnReviewCommentCount(uin string,appId int64,topicType int64, topicId string) (int64, error) {
	// 获取用户未被审核的评论数
	query := fmt.Sprintf("select count(*) as recCount from %s where uin = ? and app_id = ? and topic_type =  ? and topic_id = ? and review_status = 0 and deleted_at is null ",  m.table)
	var resp int64
	err := m.conn.QueryRow(&resp, query,uin,appId,topicType,topicId)

	if err != nil && err != sqlc.ErrNotFound {
		return resp,err
	}

	return resp,nil
}