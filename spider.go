package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gocolly/colly"
	"github.com/opesun/goquery"
)

var URL = flag.String("url", "https://www.hao123.com/", "初始化网址，默认: https://www.hao123.com/")
var API = flag.String("api", "https://api.qiyi.io/", "API 接口地址，如: https://api.qiyi.io/")
var token = flag.String("token", "7d73c01d-d16d-45a5-878f-708567945502", "API 接口验证信息，如: 7d73c01d-d16d-45a5-878f-708567945502")

type PostData struct {
	Where struct {
		Url string `json:"url"`
	} `json:"where"`
	Data struct {
		Title   string `json:"title"`
		Url     string `json:"url"`
		Content string `json:"content"`
		Timer   int64  `json:"timer"`
	} `json:"data"`
	Other struct {
		Upsert bool `json:"upsert"`
	} `json:"other"`
}

func main() {
	flag.Parse()

	c := colly.NewCollector()
	c.UserAgent = "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.108 Safari/537.36"

	// Find and visit all links
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		e.Request.Visit(e.Attr("href"))
		//fmt.Println("[", e.Request.ID, "]检索到链接:", e.Attr("href"), e.Text)
	})

	c.OnResponse(func(resp *colly.Response) {
		//fmt.Println(string(resp.Body))
		//fmt.Println(resp.Request.URL, string(resp.Body), resp.Request.ID)
		//_type := resp.Headers.Get("Content-Type")
		//code := strings.FieldsFunc(_type, func(s rune) bool {
		//	if s == '=' {
		//		return true
		//	}
		//	return false
		//})
		//fmt.Println(code)

		//body := ""
		//if len(code) >= 2 {
		//	encode := strings.FieldsFunc(code[1], func(s rune) bool {
		//		if s == ';' {
		//			return true
		//		}
		//		return false
		//	})
		//	body = mahonia.NewDecoder(encode[0]).ConvertString(string(resp.Body))
		//} else {
		//	body = string(resp.Body)
		//}
		body := string(resp.Body)

		p, _ := goquery.ParseString(body)
		//fmt.Println(p.Find("title").Text())

		// 准备提交的数据
		_data := &PostData{}

		_data.Where.Url = resp.Request.URL.String()

		_data.Data.Title = p.Find("title").Text()
		_data.Data.Url = resp.Request.URL.String()
		_data.Data.Content = p.Html()
		_data.Data.Timer = time.Now().Unix()

		_data.Other.Upsert = true

		data, err := json.Marshal(_data)
		if err != nil {
			panic(err)
		}
		// fmt.Println(string(data))

		// 将数据保存到远程服务器
		result, err := Api("intenet/update", data)
		//fmt.Println(result)
		if err != nil {
			fmt.Println("[", resp.Request.ID, "]信息获取完毕:", resp.Request.URL, p.Find("title").Text(), "保存失败! 错误代码:", err)
		} else {
			fmt.Println("[", resp.Request.ID, "]信息获取完毕:", resp.Request.URL, p.Find("title").Text(), "保存成功!", result)
		}

	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("[", r.ID, "]开始访问: ", r.URL)
	})

	c.Visit(*URL)
}

func Api(url string, data []byte) (string, error) {
	client := &http.Client{}
	//生成要访问的url
	body := bytes.NewReader(data)
	//提交请求
	reqest, err := http.NewRequest("POST", *API+url, body)

	//增加header选项
	reqest.Header.Add("Authorization", *token)

	if err != nil {
		panic(err)
	}
	//处理返回结果
	response, _ := client.Do(reqest)
	defer response.Body.Close()
	b, err := ioutil.ReadAll(response.Body)
	//if err != nil {
	//	log.Println("http.Do failed,[err=%s][url=%s]", err, url)
	//}
	return string(b), err
}
