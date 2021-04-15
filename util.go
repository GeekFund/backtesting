package backtesting

import (
	"regexp"
	"strconv"
	"time"
)

//ParseDate 转化日期
func ParseDate(t string) time.Time {
	tm, _ := time.Parse("2006-01-02", t)
	return tm
}

//DateToString 格式化日期格式
func DateToString(t time.Time) string {
	return t.Format("2006-01-02")
}

//ParseFloat32 字符串转float32
func ParseFloat32(s string) float32 {
	v, _ := strconv.ParseFloat(s, 32)
	return float32(v)
}

//FindAllStringSubmatch 查找字符串
func FindAllStringSubmatch(rex string, str string) []string {
	p := regexp.MustCompile(rex)
	r := p.FindAllStringSubmatch(str, -1)
	if len(r) > 0 {
		return r[0]
	}
	return nil
}

//FindAllString 查找字符串
func FindAllString(rex string, str string) []string {
	p := regexp.MustCompile(rex)
	return p.FindAllString(str, -1)
}

//DiffDays 判断相差天数
func DiffDays(t1, t2 time.Time) int {
	return int(t1.Sub(t2).Hours() / 24)
}
