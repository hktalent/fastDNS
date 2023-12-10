package option

import (
	"context"
	"embed"
	"fmt"
	"github.com/go-ping/ping"
	util "github.com/hktalent/go-utils"
	"math"
	"net"
	"strings"
	"sync"
	"time"
)

type TopInfo struct {
	Time float64
	Dns  string
}

/*
Provider	Primary DNS	Secondary DNS
Google	8.8.8.8	8.8.4.4
Control D	76.76.2.0	76.76.10.0
Quad9	9.9.9.9	149.112.112.112
OpenDNS Home	208.67.222.222	208.67.220.220
Cloudflare	1.1.1.1	1.0.0.1
CleanBrowsing	185.228.168.9	185.228.169.9
Alternate DNS	76.76.19.19	76.223.122.150
AdGuard DNS	94.140.14.14	94.140.15.15
*/

func PingDns(s string, nCnt int, nTimeout time.Duration) float64 {
	pinger, err := ping.NewPinger(s)
	if err != nil {
		fmt.Printf("x")
		return 0
	}
	pinger.Timeout = nTimeout * time.Millisecond
	pinger.Count = nCnt
	err = pinger.Run() // Blocks until finished.
	if err != nil {
		fmt.Printf("x")
		return 0
	}
	fmt.Printf(".")
	stats := pinger.Statistics() // get s
	return stats.AvgRtt.Seconds()
}

var nMax float64 = math.MaxFloat64

func GetN(n int) []*TopInfo {
	var top []*TopInfo
	for j := 0; j < n; j++ {
		top = append(top, &TopInfo{nMax, ""})
	}
	return top
}

func DoPingAlDns(dnsServers *[]string, topDns *[]*TopInfo, cT chan struct{}, opt *Options) {
	var wg sync.WaitGroup
	WaitOneFunc4WgParmsChan[string](&wg, func(x string) {
		f1 := PingDns(x, opt.PingCnt, opt.PingTimeout)
		UpTop(&wg, topDns, f1, x)
	}, cT, *dnsServers...)
	wg.Wait()
}
func GetTestDomain(dnsServers *[]string, topDns *[]*TopInfo, opt *Options) chan string {
	var out = make(chan string)
	go func() {
		defer close(out)
		if opt.EnablePing {
			for _, x := range *topDns {
				if dnsServer := strings.TrimSpace(x.Dns); "" != dnsServer {
					out <- dnsServer
				}
			}
		} else {
			for _, x := range *dnsServers {
				if x = strings.TrimSpace(x); "" != x {
					out <- x
				}
			}
		}
	}()
	return out
}

func LoadDns(s string, dnsDir *embed.FS) string {
	if data, err := dnsDir.ReadFile("data/" + s + ".txt"); nil == err {
		return string(data)
	}
	if s != "cn" {
		return LoadDns("cn", dnsDir)
	}
	return ""
}
func DoFastDns(dnsDir *embed.FS) {
	opt := ParseOption(dnsDir)
	// 定义需要测试的dns服务器列表 +dnsCnUs
	a11 := LoadDns(opt.Type, dnsDir)
	dnsServers := strings.Split(a11, "\n")

	// 定义一个map来存储dns服务器的ip地址和响应时间
	top := GetN(opt.OutTopNum)
	topDns := GetN(opt.TopPingDnsNum)
	dnsResults := make(map[string]int64)
	var wg sync.WaitGroup
	var cT = make(chan struct{}, opt.Thread)
	var out = util.ReadFile4Line(opt.DomainFile)
	var nCnt int64 = 0
	var nStop int64 = opt.TestStop
	// 计算 ping 再走dns
	if opt.EnablePing {
		DoPingAlDns(&dnsServers, &topDns, cT, opt)
	}
	var domain1 = GetTestDomain(&dnsServers, &topDns, opt)
	var aDnsLst []string
	for dnsServer := range domain1 {
		aDnsLst = append(aDnsLst, dnsServer)
	}
	for domain := range out {
		if nStop < nCnt {
			continue
		}
		nCnt++
		// 遍历dns服务器列表，并测试其响应时间
		for _, dnsServer := range aDnsLst {
			cT <- struct{}{}
			util.WaitFunc4WgParms[interface{}](&wg, []interface{}{dnsServer, *domain, &dnsResults}, func(x ...interface{}) {
				defer func() {
					<-cT
				}()
				var dns, domain5, dnsResults1 = x[0].(string), x[1].(string), x[2].(*map[string]int64)
				var x1, s3 = DoOneDns(dns, domain5, dnsResults1)
				UpTop(&wg, &top, float64(x1), s3)
			})
		}
	}
	wg.Wait()
	fmt.Println("\nTotal number of tests：", (nCnt-1)*int64(len(aDnsLst)), "Top 5 DNS IP addresses: ：")
	for i, x := range top {
		if "" != x.Dns {
			fmt.Printf("%d\t%2f\t%s\n", i, x.Time, x.Dns)
		}
	}
}

var (
	lock   sync.Mutex
	upLock sync.Mutex
	dial   = &net.Dialer{Timeout: 1 * time.Second}
)

func UpTop(wg *sync.WaitGroup, top *[]*TopInfo, n float64, dns string) {
	if -1 == n {
		return
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		upLock.Lock()
		defer upLock.Unlock()
		for _, x := range *top {
			if float64(n)/1e9 < x.Time {
				x.Dns = dns
				x.Time = float64(n) / 1e9
				break
			}
		}
	}()
}
func DoOneDns(dns, domain string, dnsResults *map[string]int64) (int64, string) {
	// 开始计时
	start := time.Now()

	// 创建一个dns客户端
	client := &net.Resolver{
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			return dial.DialContext(ctx, "tcp", dns)
		},
	}
	// 使用dns服务器的ip地址解析一个域名，例如google.com
	a, err := client.LookupHost(context.Background(), domain)

	if err != nil || 0 == len(a) {
		//fmt.Println(dns, domain, err)
		fmt.Printf("x")
		return -1, ""
	}

	// 停止计时
	end := time.Now()
	//fmt.Println("ok", dns, domain, len(a))
	fmt.Printf(".")
	// 计算响应时间
	responseTime := end.Sub(start).Nanoseconds()

	// 将响应时间存储到map中
	lock.Lock()
	defer lock.Unlock()
	(*dnsResults)[dns] += responseTime
	return responseTime, dns
}
