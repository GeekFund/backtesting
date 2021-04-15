package backtesting

import (
	"log"
	"testing"
	"time"
)

func TestStrategy(t *testing.T) {
	var s = Strategy{
		CycleType:  CycleMonth,
		CycleValue: 1,
	}
	now := time.Now()
	var last interface{}
	for i := 0; i < 90; i++ {
		now = now.AddDate(0, 0, 1)
		s.CycleType = CycleMonth
		s.CycleValue = 5
		if s.IsBuyDay(now, last) {
			last = now
			log.Println("月", now.String())
		}
	}
	for i := 0; i < 90; i++ {
		now = now.AddDate(0, 0, 1)
		s.CycleType = CycleTowWeek
		s.CycleValue = 5
		if s.IsBuyDay(now, last) {
			last = now
			log.Println("2周", now.String())
		}
	}
	for i := 0; i < 90; i++ {
		now = now.AddDate(0, 0, 1)
		s.CycleType = CycleWeek
		s.CycleValue = 5
		if s.IsBuyDay(now, last) {
			last = now
			log.Println("1周", now.String())
		}
	}
}
