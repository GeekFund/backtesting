package backtesting

import (
	"fmt"
	"math"
	"time"
)

type (
	//Engine 计算引擎
	Engine struct {
		strategy Strategy        //策略参数
		nws      NetWorthList    //净值列表
		trans    TransactionList //交易记录
		shares   float32         //份额
		invest   float32         //投入
		profit   float32         //利润
		balance  float32         //余额
		value    float32         //资产总值
		nav      float32         //当日净值
		rop      float32         //利润率
		date     time.Time       //当日日期
	}
	//Result 运行结果
	Result struct {
		Date      time.Time
		Nav       float32
		Invest    float32
		Balance   float32
		Value     float32
		Shares    float32
		Profit    float32
		Rop       float32
		TransList TransactionList
	}
)

//NewEngine 创建回测引擎
func NewEngine(strategy Strategy, nws NetWorthList) *Engine {
	nws.InitVolaRoc(strategy.VolaDays)
	return &Engine{
		strategy: strategy,
		nws:      nws,
	}
}

//Run 运行回测
func (ctx *Engine) Run() Result {
	for i := 0; i < len(ctx.nws); i++ {
		nw := ctx.nws[i]
		if DiffDays(nw.Date, ctx.strategy.StartDate) < 0 || DiffDays(nw.Date, ctx.strategy.EndDate) > 0 {
			continue
		}
		// log.Println("日期", ctx.strategy.StartDate, ctx.strategy.EndDate, nw.Date)
		ctx.runToday(i, nw)
	}

	return Result{
		Date:      ctx.date,
		Nav:       ctx.nav,
		Invest:    ctx.invest,
		Balance:   ctx.balance,
		Value:     ctx.value,
		Shares:    ctx.shares,
		Profit:    ctx.profit,
		Rop:       ctx.rop,
		TransList: ctx.trans,
	}
}

//RunToday 执行当天的结果
func (ctx *Engine) runToday(i int, nw NetWorth) {
	if nw.Splits > 0 {
		ctx.spilit(nw)
	}
	if nw.Dividends > 0 {
		ctx.dividends(nw)
	} else if ctx.isSellDay(nw) {
		shares := ctx.recoSell(nw)
		ctx.sell(shares, nw)
	} else if ctx.isBuyDay(nw) {
		amount := ctx.recoBuy(nw)
		ctx.fixed(amount, nw)
	} else if ctx.isAppendDay(nw) {
		var amount = ctx.recoAppend(nw)
		ctx.append(amount, nw)
	}
	ctx.refresh(nw)
}

func (ctx *Engine) fixed(amount float32, nw NetWorth) *Transaction {
	if ctx.balance > amount {
		ctx.balance -= amount
	} else {
		ctx.invest += amount
	}
	return ctx.buy(nw, amount, ctx.strategy.TransRate, TransFixed)
}

func (ctx *Engine) buy(nw NetWorth, amount, transrate float32, transType TransType) *Transaction {
	if amount <= 0 {
		return nil
	}
	nav := nw.NAV
	transfee := ParseFloat32(fmt.Sprintf("%.2f", amount*transrate/100))
	shares := ParseFloat32(fmt.Sprintf("%.2f", (amount-transfee)/nav))
	trans := Transaction{
		Date:      nw.Date,
		Amount:    amount,
		NAV:       nav,
		TransFee:  transfee,
		Shares:    shares,
		TransType: transType,
		Args: map[string]interface{}{
			"VolaRoc": nw.VolaRoc,
		},
	}
	ctx.trans.Append(trans)
	ctx.shares += shares
	return &trans
}

func (ctx *Engine) sell(shares float32, nw NetWorth) *Transaction {
	amount := nw.NAV * shares
	ctx.shares -= shares
	trans := Transaction{
		Date:      nw.Date,
		Amount:    amount,
		NAV:       nw.NAV,
		TransFee:  0,
		Shares:    shares,
		TransType: TransSell,
		Args: map[string]interface{}{
			"VolaRoc": nw.VolaRoc,
		},
	}
	ctx.balance += amount
	ctx.trans.Append(trans)
	return &trans
}

func (ctx *Engine) append(amount float32, nw NetWorth) *Transaction {
	if ctx.balance > amount {
		ctx.balance -= amount
	} else {
		ctx.invest += amount
	}
	return ctx.buy(nw, amount, ctx.strategy.TransRate, TransAppend)
}

func (ctx *Engine) dividends(nw NetWorth) *Transaction {
	amount := nw.Dividends * ctx.shares
	return ctx.buy(nw, amount, 0, TransDividends)
}

func (ctx *Engine) spilit(nw NetWorth) {
	ctx.shares = ctx.shares * nw.Splits
}

//recoBuy 推荐购买金额
func (ctx *Engine) recoBuy(nw NetWorth) float32 {
	amount := ctx.strategy.BasicAmount
	if ctx.strategy.FixedMethod == FloatInvest {
		amount = ctx.amountB(nw)
	}
	if amount > ctx.strategy.MaxAmount {
		amount = ctx.strategy.MaxAmount
	} else if amount < ctx.strategy.MinAmount {
		amount = ctx.strategy.MinAmount
	} else if amount < 100 {
		amount = 0
	}
	return float32(int(math.Ceil(float64(amount)/10) * 10))
}

func (ctx *Engine) amountB(nw NetWorth) float32 {
	var amount = ctx.strategy.BasicAmount
	roc := nw.VolaRoc
	if roc == 0 {
		return amount
	}
	//考虑6%的gdp增长
	// roc -= 0.06
	r := amount*roc/100 + amount/-roc/100
	if r > 30 {
		return amount - r*3
	} else if r > 20 {
		return amount - r*2
	} else if r > 0 {
		return amount - r*1.5
	}
	r *= 10 //下跌的话，追加金额是系数的10倍。
	if roc < -40 {
		return amount - r*10
	} else if roc < -30 {
		return amount - r*8
	} else if roc < -25 {
		return amount - r*6
	} else if roc < -20 {
		return amount - r*4
	} else if roc < -10 {
		return amount - r*2
	}
	return amount - r*1.5
}

func (ctx *Engine) recoSell(nw NetWorth) float32 {
	//卖多少
	var last = ctx.trans.LastSell()
	var shares = ctx.shares * 0.1
	//如果30天内，有卖出，则卖出份额在上次卖出份额的基础上增加卖出期间的涨幅
	if last != nil && nw.NAV > last.NAV && nw.Date.Sub(last.Date).Hours()/24 < 30 {
		//当前相当于上次卖出涨幅确定卖出比例
		//如果卖出在一个月以内，则加上较上次卖出涨幅比例+0.1
		shares = last.Shares * (1 + (nw.NAV-last.NAV)/last.NAV)
	}
	return shares
}

func (ctx *Engine) recoAppend(nw NetWorth) float32 {
	return ctx.recoBuy(nw)
}

func (ctx *Engine) isBuyDay(nw NetWorth) bool {
	var last interface{}
	trans := ctx.trans.LastBuy()
	if trans != nil {
		last = trans.Date
	}
	return ctx.strategy.IsBuyDay(nw.Date, last)
}

func (ctx *Engine) isSellDay(nw NetWorth) bool {
	last := ctx.trans.LastSell()
	r := nw.VolaRoc > ctx.strategy.SellPoint
	if last != nil {
		r = r && nw.Date.Sub(last.Date).Hours()/24 > 10
	}
	return r
}

func (ctx *Engine) isAppendDay(nw NetWorth) bool {
	if nw.ROC < -3 && nw.VolaRoc < 10 {
		return true
	}
	return false
}

func (ctx *Engine) refresh(nw NetWorth) {
	ctx.date = nw.Date
	ctx.nav = nw.NAV
	ctx.value = ctx.shares*ctx.nav + ctx.balance
	ctx.profit = ctx.value - ctx.invest
	ctx.rop = ctx.profit / ctx.invest * 100
}

// //volaRoc 周期区间内涨跌幅
// func (ctx *Engine) volaRoc(todayIndex int) float32 {
// 	start := 0
// 	//当前位置往前推
// 	if len(ctx.nws[:todayIndex]) > ctx.strategy.VolaDays {
// 		start = todayIndex - ctx.strategy.VolaDays
// 	}
// 	return ctx.nws.VolaRoc(start, todayIndex)
// }

// //volaRoc 周期区间内涨跌幅
// func (ctx *Engine) volaRocAt(start, end int) float32 {
// 	if len(ctx.nws) < end-start {
// 		end = len(ctx.nws)
// 		start = 0
// 	}
// 	var roc float32 = 0
// 	for _, i := range ctx.nws[start:end] {
// 		roc += i.ROC
// 	}
// 	ctx.roc = roc
// 	return ctx.roc
// }

// //volaRoc 周期区间内涨跌幅
// func (ctx *Engine) rebuildVolaRoc(todayIndex int) float32 {
// 	start := 0
// 	//当前位置往前推
// 	if len(ctx.nws[:todayIndex]) > ctx.strategy.VolaDays {
// 		start = todayIndex - ctx.strategy.VolaDays
// 	}
// 	//计算平滑周期内的涨跌幅
// 	if start > 0 {
// 		ctx.roc = 0
// 		for _, i := range ctx.nws[start:todayIndex] {
// 			ctx.roc += i.ROC
// 			log.Println("重新计算", ctx.roc, len(ctx.nws[start:todayIndex]))
// 		}
// 	} else {
// 		ctx.roc = 0
// 		for _, i := range ctx.nws[:todayIndex] {
// 			ctx.roc += i.ROC
// 			log.Println("重新计算", ctx.roc)
// 		}
// 	}
// 	return ctx.roc
// }
