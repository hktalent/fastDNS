package main

import (
	"embed"
	"github.com/hktalent/fastDNS/option"
	util "github.com/hktalent/go-utils"
	"sync"
)

//go:embed config/*
var config embed.FS

// 全球 公共的 免费 的 62789 个dns
// https://public-dns.info/nameserver/cn.txt
// https://public-dns.info/nameserver/nameservers.txt
//
//go:embed data/*
var dnsDir embed.FS

func main() {
	util.Wg = &sync.WaitGroup{}
	util.DoInit(&config)
	option.DoFastDns(&dnsDir)
	util.Wg.Wait()
	util.CloseAll()
}
