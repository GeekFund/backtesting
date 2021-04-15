package backtesting

import "time"

type (
	//NetWorth 净值信息
	NetWorth struct {
		Date      time.Time
		NAV       float32
		CNAV      float32
		ROC       float32
		Dividends float32
		Splits    float32
		VolaRoc   float32 //周期内涨跌幅
	}
	//NetWorthList 净值列表
	NetWorthList []NetWorth
)

//Today 获取指定当天的净值数据
func (items NetWorthList) Today(today time.Time) (k int, nw NetWorth) {
	for k, nw := range items {
		if today.Sub(nw.Date).Hours() < 24 && today.Sub(nw.Date).Hours() >= 0 {
			return k, nw
		}
	}
	return -1, nw
}

//InitVolaRoc 初始化周期内涨跌幅。避免循环内重复计算
func (items NetWorthList) InitVolaRoc(days int) {
	start := 0
	for i := 0; i < len(items); i++ {
		var roc float32
		//当前位置
		if i > days {
			start = i - days
		}
		for _, t := range (items)[start:i] {
			roc += t.ROC
		}
		nw := &items[i]
		nw.VolaRoc = roc
	}

}
