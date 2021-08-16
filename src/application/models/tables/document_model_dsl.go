package tables

type DocumentModelDsl struct {
	Id             int64      `json:"id" xorm:"id pk autoincr"`
	Mid            int64      `json:"mid" xorm:"comment('模型ID') int(5)"`
	FieldType      int64      `json:"field_type" xorm:"comment('字段类型') int(5)"` // 字段类型ID  命名搞错了 先这样写程序吧
	FormName       string     `json:"form_name"  xorm:"comment('表单名称') varchar(50)"`
	TableField     string     `json:"table_field" xorm:"comment('表字段') varchar(50)"`
	ListOrder      int64      `xorm:"listorder comment('排序值')" json:"listorder"`
	Required       bool       `json:"required"  xorm:"comment('是否必填') tinyint(1)"`
	IsDict         bool       `json:"is_dict"  xorm:"comment('字典数据源') tinyint(1)"`
	DictKey        string     `json:"dict_key" xorm:"comment('字典分类name,启用is_dict后需设置此字段') varchar(100)"`
	Datasource     string     `json:"datasource" xorm:"comment('数据源，链接或json') text"`
	RequiredTips   string     `json:"required_tips"  xorm:"comment('必填字段信息') varchar(100)"`
	Validator      string     `json:"validator"  xorm:"comment('验证器或规则') varchar(100)"`
	Default        string     `json:"default"  xorm:"comment('默认值') varchar(100)"` //默认值
	Status         bool       `json:"status" xorm:"comment('状态 0=禁用 1=启用') tinyint(1)"`
	MainTableField bool       `json:"main_table_field" xorm:"comment('是否为主表字段') tinyint(1)"`
	Component      string     `json:"component" xorm:"comment('自定义组件配置') text"`
	Searchable     bool       `json:"searchable" xorm:"comment('是否可搜索') tinyint(1)"`
	SearchType     uint       `json:"search_type" xorm:"comment('搜索类型 1=精确 2=模糊 3=多值 4=范围') tinyint(2)"`
	Span           uint       `json:"span" xorm:"comment('表单span宽度') tinyint(3)"`
	Sortable       bool       `json:"sortable" xorm:"comment('是否可排序') tinyint(1)"`
	Visible        bool       `json:"visible" xorm:"comment('是否表单可见') tinyint(1)"`
	ListVisible    bool       `json:"list_visible" xorm:"comment('是否列表可见') tinyint(1)"`
	FieldLen       uint       `json:"field_len" xorm:"comment('字段长度') bigint"`
	ListWidth      uint       `json:"list_width" xorm:"comment('列表字段宽度') int(10)"`
	Center         bool       `json:"center" xorm:"comment('是否列表居中显示内容') tinyint(1)"`
	CreatedAt      *LocalTime `xorm:"created" json:"created_at"`
	UpdatedAt      *LocalTime `xorm:"updated" json:"updated_at"`
	DeletedAt      *LocalTime `xorm:"deleted" json:"deleted_at"`
}

type ModelDslFields []DocumentModelDsl

//GetListFields 允许表单显示的列 固定字段不可隐藏
func (m ModelDslFields) GetListFields() []string {
	var fields = []string{"id", "title", "catid", "status"}
	for _, field := range m {
		if field.ListVisible {
			fields = append(fields, field.TableField)
		}
	}
	return fields
}

// GetSearchableFields 构建搜索字段key
func (m ModelDslFields) GetSearchableFields() {
	for _, field := range m {
		if field.Searchable {
			switch field.FieldType {

			}
		}
	}

}
