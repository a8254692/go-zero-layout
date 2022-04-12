package coingoods

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/builder"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/core/stringx"

	"minicode.com/sirius/go-back-server/config/cfgstatus"
)

var (
	coinGoodsFieldNames          = builder.RawFieldNames(&CoinGoods{})
	coinGoodsRows                = strings.Join(coinGoodsFieldNames, ",")
	coinGoodsRowsExpectAutoSet   = strings.Join(stringx.Remove(coinGoodsFieldNames, "`id`", "`create_time`", "`update_time`"), ",")
	coinGoodsRowsWithPlaceHolder = strings.Join(stringx.Remove(coinGoodsFieldNames, "`id`", "`create_time`", "`update_time`"), "=?,") + "=?"
)

type (
	CoinGoodsModel interface {
		Insert(data *CoinGoods) (sql.Result, error)
		FindOne(id int64) (*CoinGoods, error)
		FindAll() (*[]CoinGoods, error)
		FindAllByGoodsType(goodsType int64) (*[]CoinGoods, error)
		Update(data *CoinGoods) error
		UpdateStock(goodsId int64, exchangeNum int64) error
		UpdateShowExchangeNum(goodsId int64, exchangeNum int64) error
		Delete(id int64) error
	}

	defaultCoinGoodsModel struct {
		conn  sqlx.SqlConn
		table string
	}

	CoinGoods struct {
		Id            int64          `db:"id"`
		CoinGoodsType int64          `db:"coin_goods_type"` // 商品类型 1实物礼品 2虚拟物品
		Code          string         `db:"code"`            // 商品code
		Name          string         `db:"name"`            // 商品名称
		Stock         int64          `db:"stock"`           // 库存
		CoinNum       int64          `db:"coin_num"`        // 所需兑换的积木数量
		InitNum       int64          `db:"init_num"`        // 已兑换初始值
		CoinLimit     int64          `db:"coin_limit"`      // 每个人的兑换限制 默认是1
		ExchangeNum   int64          `db:"exchange_num"`    // 实际兑换数量
		Sort          int64          `db:"sort"`            // 排序值 升序排列
		Thumbnail     string         `db:"thumbnail"`       // 缩略图
		CoverImg      sql.NullString `db:"cover_img"`       // 详情图片
		Description   sql.NullString `db:"description"`     // 文案
		Status        int64          `db:"status"`          // 状态 1 下架 2 上架 默认下架
		CreatedAt     time.Time      `db:"created_at"`
		UpdatedAt     time.Time      `db:"updated_at"`
		DeletedAt     sql.NullTime   `db:"deleted_at"`
	}
)

func NewCoinGoodsModel(conn sqlx.SqlConn) CoinGoodsModel {
	return &defaultCoinGoodsModel{
		conn:  conn,
		table: "`coin_goods`",
	}
}

func (m *defaultCoinGoodsModel) Insert(data *CoinGoods) (sql.Result, error) {
	query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", m.table, coinGoodsRowsExpectAutoSet)
	ret, err := m.conn.Exec(query, data.CoinGoodsType, data.Code, data.Name, data.Stock, data.CoinNum, data.InitNum, data.CoinLimit, data.Sort, data.Thumbnail, data.CoverImg, data.Description, data.Status, data.CreatedAt, data.UpdatedAt, data.DeletedAt)
	return ret, err
}

func (m *defaultCoinGoodsModel) FindOne(id int64) (*CoinGoods, error) {
	query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", coinGoodsRows, m.table)
	var resp CoinGoods
	err := m.conn.QueryRow(&resp, query, id)
	switch err {
	case nil:
		return &resp, nil
	case sqlc.ErrNotFound:
		return nil, nil
	default:
		return nil, err
	}
}

func (m *defaultCoinGoodsModel) FindAll() (*[]CoinGoods, error) {
	query := fmt.Sprintf("select %s from %s where status = ? order by sort asc", coinGoodsRows, m.table)
	var resp []CoinGoods
	err := m.conn.QueryRows(&resp, query, cfgstatus.CoinGoodsPutAway)
	switch err {
	case nil:
		return &resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *defaultCoinGoodsModel) FindAllByGoodsType(goodsType int64) (*[]CoinGoods, error) {
	query := fmt.Sprintf("select %s from %s where coin_goods_type = ? and status = ? order by sort asc", coinGoodsRows, m.table)
	var resp []CoinGoods
	err := m.conn.QueryRows(&resp, query, goodsType, cfgstatus.CoinGoodsPutAway)
	switch err {
	case nil:
		return &resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *defaultCoinGoodsModel) Update(data *CoinGoods) error {
	query := fmt.Sprintf("update %s set %s where `id` = ?", m.table, coinGoodsRowsWithPlaceHolder)
	_, err := m.conn.Exec(query, data.CoinGoodsType, data.Code, data.Name, data.Stock, data.CoinNum, data.InitNum, data.CoinLimit, data.Sort, data.Thumbnail, data.CoverImg, data.Description, data.Status, data.CreatedAt, data.UpdatedAt, data.DeletedAt, data.Id)
	return err
}

func (m *defaultCoinGoodsModel) UpdateStock(goodsId int64, exchangeNum int64) error {
	query := fmt.Sprintf("update %s set exchange_num = exchange_num + ? where `id` = ?", m.table)
	_, err := m.conn.Exec(query, exchangeNum, goodsId)
	return err
}

func (m *defaultCoinGoodsModel) UpdateShowExchangeNum(goodsId int64, exchangeNum int64) error {
	query := fmt.Sprintf("update %s set init_num = init_num + ? where `id` = ?", m.table)
	_, err := m.conn.Exec(query, exchangeNum, goodsId)
	return err
}

func (m *defaultCoinGoodsModel) Delete(id int64) error {
	query := fmt.Sprintf("delete from %s where `id` = ?", m.table)
	_, err := m.conn.Exec(query, id)
	return err
}
