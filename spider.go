package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/opesun/goquery"
)

var URL = flag.String("url", "https://www.hao123.com/", "初始化网址")
var API = flag.String("api", "https://api.ireoo.com/", "API 接口地址")
var token = flag.String("token", "b910996b-c82e-4558-80bf-83dcac747bee", "API接口验证信息")

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

type NowURL struct {
	Url    string
	Number int
}

func main() {
	LastURL := NowURL{
		Url:    "",
		Number: 0,
	}

	flag.Parse()

	c := colly.NewCollector()
	c.UserAgent = "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.108 Safari/537.36"

	// Find and visit all links
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		href := e.Attr("href")
		reg := regexp.MustCompile(`^(https?):\/\/([\w\-]+(\.[\w\-]+)*\/)*[\w\-]+(\.[\w\-]+)*\/?(\?([\w\-\.,@?^=%&:\/~\+#]*)+)?`)
		r := reg.FindAllString(href, -1)
		if len(r) > 0 {
			// 格式化href
			uri, err := url.Parse(href)
			if err == nil {
				if uri.Host == LastURL.Url {
					LastURL.Number++
					if LastURL.Number < 11 {
						e.Request.Visit(href)
						log.Println("[", e.Request.ID, "] 检索到链接:", href, e.Text)
					} else {
						log.Println("[", e.Request.ID, "] 检索到链接:", href, "同一时间重复次数过多,已经丢弃!")
					}
				} else {
					LastURL.Url = uri.Host
					LastURL.Number = 0
				}
			}
		}
	})

	c.OnResponse(func(resp *colly.Response) {
		_type := strings.FieldsFunc(resp.Headers.Get("Content-Type"), func(s rune) bool {
			if s == ';' {
				return true
			}
			return false
		})

		if _type[0] == "text/html" {
			body := string(resp.Body)

			p, _ := goquery.ParseString(body)
			//fmt.Println(p.Find("title").Text())

			log.Println("[", resp.Request.ID, "] 信息获取完毕:", resp.Request.URL, p.Find("title").Text())

			// 准备提交的数据
			_data := &PostData{}

			_data.Where.Url = resp.Request.URL.String()

			_data.Data.Title = p.Find("title").Text()
			_data.Data.Url = resp.Request.URL.String()
			_data.Data.Content = body
			_data.Data.Timer = time.Now().Unix()

			_data.Other.Upsert = true

			data, err := json.Marshal(_data)
			if err != nil {
				log.Println("[", resp.Request.ID, "] 格式化数据失败! 错误代码:", err)
			}
			// fmt.Println(string(data))

			// 将数据保存到远程服务器
			result, err := Api("intenet/update", data)
			//fmt.Println(result)
			if err != nil {
				log.Println("[", resp.Request.ID, "] 提交到数据库失败! 错误代码:", err)
			} else {
				log.Println("[", resp.Request.ID, "] 提交到数据库成功!", result)
			}
		} else {
			log.Println("[", resp.Request.ID, "] 信息获取完毕:", resp.Request.URL, "不是一个网页地址!")
		}

	})

	c.OnRequest(func(r *colly.Request) {
		log.Println("[", r.ID, "] 开始访问: ", r.URL)
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
	reqest.Header.Set("Content-Type", "application/json")
	reqest.Header.Add("Authorization", *token)

	if err != nil {
		log.Println(err)
	}
	//处理返回结果
	response, _ := client.Do(reqest)
	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Println("http.Do failed,[err=%s][url=%s]", err, url)
	}
	defer response.Body.Close()
	return string(b), err
}
