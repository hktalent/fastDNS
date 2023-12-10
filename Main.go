package main

import (
	"embed"
	_ "embed"
	"github.com/hktalent/fastDNS/option"
)

// 全球 公共的 免费 的 62789 个dns
// https://public-dns.info/nameserver/cn.txt
// https://public-dns.info/nameserver/nameservers.txt
//
//go:embed data/*
var dnsDir embed.FS

func main() {
	option.DoFastDns(&dnsDir)
}
