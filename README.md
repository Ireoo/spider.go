# spider.go

spider with golang

> version: 1.0.0

## Golang 最新版本

```bash
sudo add-apt-repository ppa:longsleep/golang-backports
sudo apt-get update
sudo apt-get install golang-go
```

## dep 包管理安装

```bash
curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
```

## 打包程序

```bash
// window
go build -o build/spider.exe spider.go

// macos or linux
go build -o build/spider spider.go
```
