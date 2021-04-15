package backtesting

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTransactionList(t *testing.T) {
	items := make(TransactionList, 0)
	items.Append(Transaction{
		Date:      time.Now().AddDate(0, 0, -5),
		Amount:    1000,
		NAV:       1.20,
		TransFee:  1000 * 0.15,
		Shares:    (1000 - 1000*0.15) / 1.2,
		TransType: TransFixed,
	})
	items.Append(Transaction{
		Date:      time.Now().AddDate(0, 0, -5),
		Amount:    800,
		NAV:       1.20,
		TransFee:  800 * 0.15,
		Shares:    (800 - 800*0.15) / 1.2,
		TransType: TransSell,
	})
	last := items.LastSell()
	assert.NotNil(t, last)
	assert.Equal(t, last.Amount, float32(800))
	assert.Len(t, items, 2)
}
