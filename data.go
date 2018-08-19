package common

/**
个人基础信息
*/
type People struct {
	DataType      string   `json:"dataType"`      // 区分数据类型
	Id            string   `json:"id"`            // 身份证号码
	Sex           string   `json:"sex"`           // 性别
	Name          string   `json:"name"`          // 姓名
	BirthLocation Location `json:"birthLocation"` // 出生地
	LiveLocation  Location `json:"liveLocation"`  // 现居住地
	MotherId      string   `json:"motherID"`      // 母亲身份证号码
	FatherId      string   `json:"fatherID"`      // 父亲身份证号码
	Childs        []string `json:"chailds"`       // 子女身份证
}

/**
位置
*/
type Location struct {
	Country  string `json:"country"`  // 国家
	Province string `json:"province"` // 省
	City     string `json:"city"`     // 城市
	Town     string `json:"town"`     // 镇
	Detail   string `json:"detail"`   // 详细住址
}
