package backtesting

import (
	"time"
)

type (
	//Strategy 策略
	Strategy struct {
		Code        string      //代码
		BasicAmount float32     //投入基准金额
		MinAmount   float32     //最小投入
		MaxAmount   float32     //最大投入
		SellPoint   float32     //卖出点位
		AppendPoint float32     //追加点位
		StartDate   time.Time   //开始时间
		EndDate     time.Time   //截止时间
		TransRate   float32     //交易费率
		CycleType   CycleType   //周期类型
		CycleValue  int         //周期内值
		VolaDays    int         //统计涨跌幅天数
		FixedMethod FixedMethod //定投方式
	}

	//FixedMethod 定投方式
	FixedMethod int
	//CycleType 周期类型
	CycleType int
)

const (
	//CycleMonth 每月
	CycleMonth CycleType = 1
	//CycleTowWeek 每两周
	CycleTowWeek CycleType = 2
	//CycleWeek 每周
	CycleWeek CycleType = 3
)
const (
	//FixedInvest 定期定额
	FixedInvest FixedMethod = 1
	//FloatInvest 不定期不定额
	FloatInvest FixedMethod = 2
)

//IsBuyDay 是否为投资日
func (s Strategy) IsBuyDay(now time.Time, last interface{}) bool {
	if last == nil {
		return now.Day() >= s.CycleValue
	}
	var t = last.(time.Time)
	now = ParseDate(DateToString(now))
	t = ParseDate(DateToString(t))
	if now.Weekday() == 0 || now.Weekday() == 6 {
		return false
	}
	//判断定投月周期定投
	if s.CycleType == CycleMonth {
		return now.Sub(t).Hours()/24 > 25 && now.Day() >= s.CycleValue
	}
	var d = 0
	//两周定投周期判断
	if s.CycleType == CycleTowWeek {
		d = 14
	}
	//判断单周周期定投
	if s.CycleType == CycleWeek {
		d = 7
	}
	return now.Sub(t).Hours() >= float64(d-1)*24 && int(now.Weekday()) >= s.CycleValue
}
