package backtesting

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var (
	navAPI       = "http://fund.eastmoney.com/f10/F10DataApi.aspx"
	dividendsAPI = "http://fundf10.eastmoney.com/fhsp_%s.html"
)

//Basic 基金信息
type Basic struct {
	Code      string
	Name      string
	SearchKey string
}

//TianTian 天天基金
type tianTian struct {
}

//TianTian 天天基金
var TianTian = tianTian{}

//GetHistories 获取历史净值数据
func (tt *tianTian) GetHistories(code string, sdate string, edate string) []NetWorth {
	var page = 1
	items := make([]NetWorth, 0)
	for {
		query := map[string]string{}
		query["type"] = "lsjz"
		query["code"] = code
		query["sdate"] = sdate
		query["edate"] = edate
		query["per"] = "40"
		query["page"] = strconv.Itoa(page)
		resp, err := tt.request(http.MethodGet, "f10/F10DataApi.aspx", query, nil)
		if err != nil {
			log.Panic("请求天天基金失败", err)
		}
		dom, err := goquery.NewDocumentFromReader(bytes.NewReader(resp))
		if err != nil {
			log.Panic("解析失败", err, dom)
		}
		trs := dom.Find("table tbody tr")
		trs.Each(func(i int, s *goquery.Selection) {
			if s.Text() == "暂无数据!" {
				return
			}
			rate := strings.Replace(s.Find("td").Eq(3).Text(), "%", "", -1)
			if rate == "" {
				rate = "0"
			}
			navs := s.Find("td").Eq(1).Text()
			cnavs := s.Find("td").Eq(2).Text()
			if cnavs == "" {
				cnavs = navs
			}
			last := NetWorth{
				Date:   ParseDate(s.Find("td").Eq(0).Text()),
				NAV:    ParseFloat32(navs),
				CNAV:   ParseFloat32(cnavs),
				ROC:    ParseFloat32(rate),
				Splits: tt.resolveSplits(strings.TrimSpace(s.Find("td").Eq(6).Text())),
			}
			items = append(items, last)
		})

		if trs.Length() < 40 {
			break
		}
		page++
	}
	for i, j := 0, len(items)-1; i < j; i, j = i+1, j-1 {
		items[i], items[j] = items[j], items[i]
	}
	return tt.fillDividends(code, items)
}

//获取份额拆分信息
func (tt *tianTian) resolveSplits(dividend string) float32 {
	if strings.Contains(dividend, "折算") {
		splits := FindAllStringSubmatch("(\\d+\\.\\d+)", dividend)[0]
		return ParseFloat32(splits)
	}
	return 0
}

func (tt *tianTian) fillDividends(code string, items []NetWorth) []NetWorth {
	if len(items) == 0 {
		return items
	}
	dividends := tt.dividends(code)
	var fq float32 = 1
	var last NetWorth
	for i := 0; i < len(items); i++ {
		his := items[i]
		if d, ok := dividends[DateToString(his.Date)]; ok {
			his.Dividends = d
			fq *= (last.NAV / (last.NAV - d))
		}
		//除权因子 已知上一个除权因子
		//已知当前价格
		//可求出当前复权价格
		//当前价格/当前复权价格 当前复权因子
		//复权因子=分红除权日上一日的单位净值/（分红除权日上一日的单位净值-分红金额）
		//复权因子=分红除权日上一日的单位净值/（分红除权日上一日的单位净值-分红金额）=1.0742/（1.0742-0.045）=1.0437。
		//复权单位净值=复权因子*分红除权日单位净值=1.0424*1.0437=1.0880，与万得和Choice一致。
		//https://zhuanlan.zhihu.com/p/144838984

		// if his.Splits > 0 {
		// 	//如果拆分？
		// 	fq *= (last.NAV / (last.NAV - his.Dividends))
		// }
		//除权因子 已知上一个除权因子
		//已知当前价格
		//可求出当前复权价格
		//当前价格/当前复权价格 当前复权因子，
		// his.RNAV = fq * his.NAV
		items[i] = his
		last = his
	}
	return items
}

func (tt *tianTian) dividends(code string) map[string]float32 {
	u, _ := url.Parse(fmt.Sprintf(dividendsAPI, code))
	resp, err := http.Get(u.String())
	if err != nil {
		log.Panic("请求天天基金失败", err)
	}
	dom, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Panic("解析失败", err, dom)
	}
	trs := dom.Find("table.w782.cfxq tbody tr")
	m := make(map[string]float32, 0)
	trs.Each(func(i int, s *goquery.Selection) {
		if s.Text() == "暂无分红信息!" {
			return
		}
		d := FindAllString("(\\d+\\.\\d+)", s.Find("td").Eq(3).Text())[0]
		m[strings.TrimSpace(s.Find("td").Eq(1).Text())] = ParseFloat32(d)
	})
	return m
}

//GetBasic 获取基础信息
func (tt *tianTian) GetBasic(code string) Basic {
	return Basic{
		Code: code,
	}
}

func (tt *tianTian) GetFundList(page, size int) ([]Basic, error) {
	query := map[string]string{
		"op":   "dy",
		"dt":   "kf",
		"ft":   "all",
		"rs":   "",
		"gs":   "0",
		"sc":   "qjzf",
		"st":   "desc",
		"sd":   DateToString(time.Now()),
		"ed":   DateToString(time.Now()),
		"qdii": "",
		"pi":   fmt.Sprintf("%d", page),
		"pn":   fmt.Sprintf("%d", size),
		"dx":   "0",
	}
	var resp, err = tt.request(http.MethodPost, "data/rankhandler.aspx", query, nil)
	if err != nil {
		return nil, err
	}
	var str = string(resp)
	var start, end = strings.Index(str, "["), strings.Index(str, "]") + 1
	var lines = []string{}
	err = json.Unmarshal([]byte(str[start:end]), &lines)
	if err != nil {
		return nil, err
	}
	var items = make([]Basic, 0)
	for _, line := range lines {
		item := strings.Split(line, ",")
		items = append(items, Basic{
			Code:      item[0],
			Name:      item[1],
			SearchKey: item[2],
		})
	}
	return items, nil
}

func (tt *tianTian) request(method, api string, query map[string]string, body io.Reader) ([]byte, error) {
	u, _ := url.Parse("https://fund.eastmoney.com")
	u.Path = api
	q := url.Values{}
	for k, v := range query {
		q.Add(k, v)
	}
	u.RawQuery = q.Encode()
	req, err := http.NewRequest(method, u.String(), body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Referer", u.String())
	http.DefaultClient.Timeout = 5 * time.Second
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(resp.Body)
}
