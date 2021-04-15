package backtesting

import (
	"log"
	"math"
	"time"
)

type (
	//PackEngine 组合引擎
	PackEngine struct {
		startDate time.Time
		amount    float32         //定额投入
		uprate    float32         //每年投入增长
		items     PackItemList    //投资项目
		month     int             //月份
		balance   float32         //余额
		value     float32         //持仓价值
		invest    float32         //总投入
		now       time.Time       //当前计算日期
		trans     TransactionList //交易列表
	}
	//PackItem 组合引擎
	PackItem struct {
		*Engine
		TOF     TOF //资金类型 资金利用优先级依次为 激进 保守 现金
		Precent int //分配比例，当激进类型不在需要投入时。激进类型比例将被分配到保守和现金中
		Result  Result
	}
	//PackItemList 项目列表
	PackItemList []PackItem
	//TOF 资金类型
	TOF int
)

const (
	//Cash 现金
	Cash TOF = 1
	//Conservative 保守的资金
	Conservative TOF = 2
	//Radical 激进的资金
	Radical TOF = 3
)

/**
组合运转目标
期望通过激进，保守，现金三种不同类型的基金及策略来实现资金的最大收益率及使用效率

组合运行逻辑
组合运行基于单个基金的投资逻辑来实现的。我们可以每月投入固定的资金到组合，组合通过策略来分配你的资金。
当激进标的不再买入的时候会将资金投入到保守及现金当中
当激进标的出现买入机会时，将从现金中抽取资金进行投入到投资目标中
当激进标的出现卖出机会时，将卖出的资金买到现金当中
当现金比例占总资产超过10%时，将多余的现金买入到保守的标的中去
当现金及保守资金超过一定比例后，将自动提高低位时的买入金额上限，动态平衡现金持有
*/

//NewPackEngine 创建组合计算引擎
func NewPackEngine(items PackItemList, startDate time.Time, amount float32) *PackEngine {
	e := &PackEngine{
		amount:    amount,
		items:     items,
		startDate: startDate,
		month:     int(startDate.Month()),
	}
	for _, item := range items {
		item.strategy.BasicAmount = amount * float32(item.Precent) / 100
		item.strategy.MaxAmount = item.strategy.BasicAmount * 10
	}
	return e
}

//Run 运行组合
func (e *PackEngine) Run() {
	days := int(math.Ceil(time.Now().Sub(e.startDate).Hours() / 24))
	for i := 0; i < days; i++ {
		now := e.startDate.AddDate(0, 0, i)
		e.now = now
		if e.investDay(now) {
			e.append()
		}
		e.refresh()
		for k := range e.items {
			item := &e.items[k]

			k, nw := item.nws.Today(now)
			if k < 0 {
				continue
			}
			trans := e.transaction(item, nw)
			if trans == nil {
				continue
			}
			e.trans.Append(*trans)
			bm := map[int]string{1: "买入", 2: "分红", 3: "追加", 4: "卖出"}
			log.Printf("%s %s %s 净值=%.4f 金额=%.2f 份额=%.2f 手续费=%.2f 持仓份额=%.2f 现金=%.2f",
				DateToString(trans.Date),
				item.strategy.Code,
				bm[int(trans.TransType)],
				trans.NAV,
				trans.Amount,
				trans.Shares,
				trans.TransFee,
				item.shares,
				e.cashValue(),
			)
		}
	}

	for _, item := range e.items {
		log.Printf("投资结果 %s %s 净值=%.4f 本金=%.2f 余额=%.2f 价值=%.2f 份额=%.2f 利润=%.2f 收益率=%.2f",
			DateToString(item.date),
			item.strategy.Code,
			item.nav,
			item.invest,
			item.balance,
			item.value,
			item.shares,
			item.profit,
			item.rop,
		)
		//比对与引擎的计算结果
		// log.Println("引擎参数", item.strategy.BasicAmount, item.strategy)
		// g := NewEngine(item.strategy, item.nws)
		// result := g.Run()
		// for i := 0; i < len(result.TransList); i++ {
		// 	pi := item.trans[i]
		// 	ei := result.TransList[i]
		// 	if pi.Date != ei.Date || pi.Amount != ei.Amount || pi.TransType != ei.TransType {
		// 		// log.Println("结果不一致", DateToString(pi.Date), DateToString(ei.Date), pi.Amount, ei.Amount, pi.TransType, ei.TransType)
		// 	}
		// }
	}
	log.Printf("投资结果 %s 本金=%.2f 余额=%.2f 价值=%.2f 利润=%.2f 收益率=%.2f",
		DateToString(e.now),
		e.invest,
		e.balance,
		e.value,
		e.value-e.invest,
		(e.value-e.invest)/e.invest*100,
	)
}

func (e *PackEngine) refresh() {
	e.value = e.balance
	for _, item := range e.items {
		e.value += item.value - item.balance
		if item.balance > 0 {
			// log.Panic("账户余额大于0", item.balance)
		}
	}
}

func (e *PackEngine) append() {
	//每月投入金额到现金
	e.balance += e.amount
	e.invest += e.amount
	e.month++
	if e.month > 12 {
		e.month = 1
		if e.uprate > 0 {
			e.amount *= 1.1
		}
	}
}

func (e *PackEngine) investDay(now time.Time) bool {
	return int(now.Month()) == e.month && now.Day() >= 1
}

func (e *PackEngine) transaction(item *PackItem, nw NetWorth) *Transaction {
	defer item.refresh(nw)
	var trans *Transaction
	if nw.Splits > 0 {
		item.spilit(nw)
	}
	if nw.Dividends > 0 {
		trans = item.dividends(nw)
		return trans
	} else if e.isSellDay(item, nw) {
		shares := e.recoSell(item, nw)
		if shares > item.shares {
			shares = item.shares
		}
		if shares <= 0 {
			return nil
		}
		trans = item.sell(shares, nw)
		e.balance += trans.Amount
		//TODO: 组合卖出将减少项目成本，非组合计算是增加余额，而组合是将余额存入总账
		// item.invest -= trans.Amount
		// item.balance = 0
		return trans
	}
	isBuyDay := item.isBuyDay(nw)
	isAppendDay := item.isAppendDay(nw)
	if isBuyDay == false && isAppendDay == false {
		return nil
	}
	//买入多少，取决于资金类型
	amount := e.recoAmount(item, nw)
	if amount > e.balance {
		amount = e.balance
	}
	if amount <= 0 {
		return nil
	}
	if isBuyDay {
		trans = item.fixed(amount, nw)
	} else if isAppendDay {
		trans = item.append(amount, nw)
	}
	//买入减去账户余额
	e.balance -= trans.Amount
	return trans
}

//recoAmount 计算推荐的买入资金
func (e *PackEngine) recoAmount(item *PackItem, nw NetWorth) float32 {
	amount := item.recoBuy(nw)
	if item.TOF == Radical {
		if amount <= item.strategy.MinAmount {
			return amount
		}
		//如果最近3个月内有卖出，就不在买入
		t := item.trans.LastSell()
		if t != nil && t.Date.AddDate(0, 3, 0).Sub(e.now).Hours() > 0 {
			return 0
		}
		// return amount
		cash := e.cashValue()
		//当前比例+预留资金比例
		pc := float32(item.Precent/100) + (1.0 - float32(e.radicalCount()))
		// pc := float32(item.Precent / 100)
		// amount = cash * pc
		//如果进攻型的下跌超过20%
		if nw.VolaRoc < -25 {
			amount = cash * pc
			// log.Printf("跌幅超30的机会？涨跌幅=%.2f 余额=%.2f 建议买入=%.2f 保守价值=%.2f", nw.VolaRoc, e.balance, amount, cash)
		} else if nw.VolaRoc < -20 {
			amount = amount*3 + cash*0.8*pc
			// log.Printf("跌幅超20的机会？涨跌幅=%.2f 余额=%.2f 建议买入=%.2f 保守价值=%.2f", nw.VolaRoc, e.balance, amount, cash)
		} else if nw.VolaRoc < -15 {
			// log.Println("跌幅超15的机会？", nw.VolaRoc, e.balance)
			amount = amount*2 + cash*0.6*pc
		} else if nw.VolaRoc < -10 {
			// log.Println("跌幅超10的机会？", nw.VolaRoc,cash)
			amount = amount*1.5 + cash*0.4*pc
		} else if nw.VolaRoc < 0 {
			// log.Println("跌幅超10的机会？", nw.VolaRoc,cash)
			amount = amount + cash*0.2*pc
		} else {
			amount = amount + cash*0.1*pc
		}
		return float32(int(math.Ceil(float64(amount/100))) * 100)
	}

	//如果一个月内有交易，则不在交易
	lastnum := len(item.trans) - 1
	if lastnum >= 0 && item.trans[lastnum].Date.AddDate(0, 1, 0).Sub(e.now).Hours() < 0 {
		return 0
	}
	//判断买入时，验证三个月内是否有卖出过。如果有，则不在买入
	sell := item.trans.LastSell()
	if sell != nil && sell.Date.AddDate(0, 3, 0).Sub(e.now).Hours() > 0 {
		return 0
	}
	amount = 0
	if e.balance > e.keepBalance() {
		amount = e.balance - e.keepBalance()
	}
	return amount
}

//获取激进保留s的金额
func (e *PackEngine) keepBalance() float32 {
	var amount float32
	for _, v := range e.items {
		if v.TOF != Radical {
			continue
		}
		_, nw := v.nws.Today(e.now)
		amount += e.recoAmount(&v, nw)
	}
	return amount
}

//获取激进保留s的金额
func (e *PackEngine) cashValue() float32 {
	var amount float32
	for _, v := range e.items {
		if v.TOF != Conservative {
			continue
		}
		amount += v.value
	}
	return amount + e.balance
}

//获取进攻配置比例
func (e *PackEngine) radicalCount() float32 {
	var count float32
	for _, v := range e.items {
		if v.TOF != Conservative {
			count = float32(v.Precent) / 100.00
		}
	}
	return count
}

//是否卖出
func (e *PackEngine) isSellDay(item *PackItem, nw NetWorth) bool {
	if item.TOF == Radical {
		return item.isSellDay(nw)
	}
	//保守型的，就得留资金给进攻型
	keep := e.keepBalance()
	if keep > e.balance*1.2 {
		return true
	}
	return item.isSellDay(nw)
}

//卖出金额
func (e *PackEngine) recoSell(item *PackItem, nw NetWorth) float32 {
	if item.TOF == Radical {
		return item.recoSell(nw)
	}
	//保守型的，就得留资金给进攻型
	keep := e.keepBalance()
	if e.balance < keep {
		// log.Println("余额不足，需要卖出给进攻", e.balance, keep)
		return keep * 1.2 / nw.NAV
	}
	return item.recoSell(nw)
}
