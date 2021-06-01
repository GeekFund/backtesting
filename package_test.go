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
	start := ParseDate("2015-01-01")
	end := ParseDate("2022-12-01")
	var amount float32 = 10000
	var items PackItemList

	items = append(items, PackItem{
		// Engine:  getEngine("007412", start, end), //景顺长城绩优
		Engine:  getEngine("260108", start, end), //景顺长城新兴成长
		Precent: 20,
		TOF:     Radical,
	})
	items = append(items, PackItem{
		// Engine:  getEngine("006228", start, end), //中欧医疗创新
		Engine:  getEngine("001938", start, end), //中欧时代先锋
		Precent: 20,
		TOF:     Radical,
	})
	items = append(items, PackItem{
		// Engine:  getEngine("163417", start, end), //兴全合宜
		Engine:  getEngine("163406", start, end), //兴全合润
		Precent: 20,
		TOF:     Radical,
	})
	items = append(items, PackItem{
		// Engine:  getEngine("005827", start, end), //易方达蓝筹
		Engine:  getEngine("110011", start, end), //易方达中小盘
		Precent: 20,
		TOF:     Radical,
	})
	e := NewPackEngine(items, start, amount)
	// e.SetBanlance(50000)
	e.Run()

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
