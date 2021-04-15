package backtesting

import (
	"time"
)

type (
	//Transaction 交易记录
	Transaction struct {
		Date      time.Time              //日期
		Amount    float32                //交易金额
		NAV       float32                //净值
		TransFee  float32                //交易费用
		Shares    float32                //交易份额
		TransType TransType              //交易类型
		Args      map[string]interface{} //买入时的参数
	}
	//TransactionList 交易记录
	TransactionList []Transaction

	//TransType 交易类型
	TransType int
)

const (
	//TransFixed 买入
	TransFixed TransType = 1
	//TransDividends 分红买入
	TransDividends TransType = 2
	//TransAppend 追加买入
	TransAppend TransType = 3
	//TransSell 卖出
	TransSell TransType = 4
)

//LastSell 最后一次卖出记录
func (items TransactionList) LastSell() *Transaction {
	return items.LastByType(TransSell)
}

//LastBuy 最后一次买入记录
func (items TransactionList) LastBuy() *Transaction {
	return items.LastByType(TransFixed)
}

//LastByType 获取交易类型的最后一次交易
func (items TransactionList) LastByType(transtype TransType) *Transaction {
	for i := len(items) - 1; i >= 0; i-- {
		item := items[i]
		if item.TransType == transtype {
			return &item
		}
	}
	return nil
}

//Append 最后一次卖出记录
func (items *TransactionList) Append(t Transaction) {
	*items = append(*items, t)
}

//Today 获取指定当天的交易
func (items TransactionList) Today(today time.Time) *Transaction {
	for _, item := range items {
		diff := today.Sub(item.Date)
		if diff.Hours() < 24 && diff.Hours() >= 0 {
			return &item
		}
	}
	return nil
}
