package commom

import (
	"Yearning-go/src/lib"
	"Yearning-go/src/model"
	"github.com/jinzhu/gorm"
)

const QueryField = "work_id, username, text, backup, date, real_name, executor, `status`, `type`, `delay`, `source`,`id_c`,`data_base`,`table`,`execute_time`,assigned,current_step,relevant"

type ExecuteStr struct {
	WorkId  string `json:"work_id"`
	Perform string `json:"perform"`
	Page    int    `json:"page"`
	Flag    int    `json:"flag"`
	Text    string `json:"text"`
	Tp      string `json:"tp"`
}

type PageInfo struct {
	Page int    `json:"page"`
	Find Search `json:"find"`
	Tp   string `json:"tp"`
}

type CommonList struct {
	Page    int         `json:"page"`
	Data    interface{} `json:"data"`
	IDC     []string    `json:"idc"`
	Source  interface{} `json:"source"`
	Query   interface{} `json:"query"`
	Auditor interface{} `json:"auditor"`
	Multi   bool        `json:"multi"`
}

type SQLOrder struct {
	List  []model.CoreSqlOrder `json:"data"`
	Multi bool                 `json:"multi"`
	Count int                  `json:"page"`
}

type SQLQuery struct {
	List  []model.CoreQueryOrder `json:"data"`
	Count int                  `json:"page"`
}

func (p *PageInfo) GetSQLOrderList(scopes ...func(*gorm.DB) *gorm.DB) (order SQLOrder) {
	start, offset := lib.Paging(p.Page, 15)
	model.DB().LogMode(true)
	model.DB().Model(model.CoreSqlOrder{}).Select(QueryField).
		Scopes(scopes...).Count(&order.Count).Order("id desc").
		Offset(start).Limit(offset).Find(&order.List)
	order.Multi = model.GloOther.Multi
	return
}

func (p *PageInfo) GetSQLQueryList(scopes ...func(*gorm.DB) *gorm.DB) (query SQLQuery) {
	start, offset := lib.Paging(p.Page, 15)
	model.DB().Model(model.CoreQueryOrder{}).
		Scopes(scopes...).Count(&query.Count).Order("id desc").
		Offset(start).Limit(offset).Find(&query.List)
	return
}

type Search struct {
	Picker   []string `json:"picker"`
	Valve    bool     `json:"valve"`
	Text     string   `json:"text"`
	Explain  string   `json:"explain"`
	WorkId   string   `json:"work_id"`
	Type     int      `json:"type"`
	Status   int      `json:"status"`
	IDC      string   `json:"idc"`
	Source   string   `json:"source"`
	Username string   `json:"username"`
	Dept     string   `json:"dept"`
}

type SQLTest struct {
	Source   string `json:"source"`
	SQL      string `json:"sql"`
	Database string `json:"data_base"`
	IsDML    bool   `json:"is_dml"`
	WorkId   string `json:"work_id"`
}

type QueryOrder struct {
	IDC      string `json:"idc"`
	Source   string `json:"source"`
	Export   uint   `json:"export"`
	Assigned string `json:"assigned"`
	Text     string `json:"text"`
	WorkId   string `json:"work_id"`
	Tp       string `json:"tp"`
}

const (
	ORDER_IS_CREATE     = "???????????????!"
	ORDER_IS_DUP        = "????????????????????????!"
	ORDER_IS_EDIT       = "??????????????????"
	ORDER_IS_DELETE     = "??????????????????"
	ORDER_IS_CLEAR      = "???????????????"
	ORDER_IS_AGREE      = "???????????????"
	ORDER_IS_REJECT     = "???????????????"
	ORDER_IS_ALL_END    = "?????????????????????"
	ORDER_IS_END        = "???????????????"
	ORDER_IS_ALL_CANCEL = "?????????????????????"
	DATA_IS_DELETE      = "??????????????????"
	DATA_IS_EDIT        = "??????????????????"
	DATA_IS_UPDATED     = "???????????????"
)
