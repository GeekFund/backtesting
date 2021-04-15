package backtesting_test

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"testing"
	"time"

	. "github.com/geekfund/backtesting"
)

func TestPackEngine(t *testing.T) {
	start := ParseDate("2010-01-01")
	end := ParseDate("2021-05-01")
	var amount float32 = 10000
	var items PackItemList

	items = append(items, PackItem{
		Engine:  getEngine("001938", start, end),
		Precent: 20,
		TOF:     Radical,
	})
	items = append(items, PackItem{
		Engine:  getEngine("163406", start, end),
		Precent: 40,
		TOF:     Radical,
	})
	items = append(items, PackItem{
		Engine:  getEngine("110011", start, end),
		Precent: 20,
		TOF:     Radical,
	})
	e := NewPackEngine(items, start, amount)
	e.Run()

	// 	bm := map[int]string{1: "买入", 2: "分红", 3: "追加", 4: "卖出"}
	// log.Printf("%s %s %s 净值=%.4f 金额=%.2f 份额=%.2f 手续费=%.2f 账户余额=%.2f 总资产=%.2f",
	// 	item.strategy.Code,
	// 	DateToString(trans.Date),
	// 	bm[int(trans.TransType)],
	// 	trans.NAV,
	// 	trans.Amount,
	// 	trans.Shares,
	// 	trans.TransFee,
	// 	e.balance,
	// 	e.value,
	// )

	// log.Printf("投资结果 本金=%.2f 余额=%.2f 价值=%.2f 利润=%.2f 收益率=%.2f",
	// 	e.invest,
	// 	e.balance,
	// 	e.value,
	// 	e.value-e.invest,
	// 	(e.value-e.invest)/e.invest*100,
	// )
}

func getNws(code string) NetWorthList {
	key := code + "2010-01-01" + DateToString(time.Now())
	bts, err := ioutil.ReadFile(key)
	var nws NetWorthList
	if err != nil {
		nws = TianTian.GetHistories(code, "2010-01-01", DateToString(time.Now()))
		bts, err = json.Marshal(nws)
		if err != nil {
			log.Panic("缓存写入失败", err)
		}
		ioutil.WriteFile(key, bts, 0644)
		log.Println("历史净值写入文件")
	} else {
		err := json.Unmarshal(bts, &nws)
		if err != nil {
			log.Panic("缓存读取失败", err)
		}
		log.Println("历史净值读取文件")
	}
	return nws
}

func getEngine(code string, start, end time.Time) *Engine {
	st := Strategy{
		Code:        code,
		SellPoint:   60,
		AppendPoint: -20,
		StartDate:   start,
		EndDate:     end,
		TransRate:   0.15,
		CycleType:   CycleMonth,
		CycleValue:  15,
		VolaDays:    180,
		FixedMethod: FloatInvest,
	}
	nws := getNws(code)
	return NewEngine(st, nws)
}
