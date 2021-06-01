package backtesting_test

import (
	"log"
	"testing"

	. "github.com/geekfund/backtesting"
)

func TestEngine(t *testing.T) {
	nws := getNws("163406")
	var amount float32 = 1000
	start := ParseDate("2010-01-01")
	end := ParseDate("2021-03-01")
	e := NewEngine(Strategy{
		BasicAmount: amount,
		MinAmount:   100,
		MaxAmount:   amount * 10,
		SellPoint:   60,
		AppendPoint: -20,
		StartDate:   start,
		EndDate:     end,
		TransRate:   0.15,
		CycleType:   CycleMonth,
		CycleValue:  15,
		VolaDays:    180,
		FixedMethod: FloatInvest,
	}, nws)
	result := e.Run()
	for i := 0; i < len(result.TransList); i++ {
		t := result.TransList[i]
		if t.TransType != TransSell {
			// continue
		}
		bm := map[int]string{1: "买入", 2: "分红", 3: "追加", 4: "卖出"}
		log.Printf("%s %s 净值=%.4f 金额=%.2f 份额=%.2f 手续费=%.2f", DateToString(t.Date), bm[int(t.TransType)], t.NAV, t.Amount, t.Shares, t.TransFee)
	}
	log.Printf("投资结果 净值=%.4f 本金=%.2f 余额=%.2f 价值=%.2f 份额=%.2f 利润=%.2f 收益率=%.2f",
		result.Nav,
		result.Invest,
		result.Balance,
		result.Value,
		result.Shares,
		result.Profit,
		result.Rop,
	)
}
