# fastDNS
What the hell is this?
I discovered by chance that the impact of DNS on the network is extremely severe.
So I thought, how to choose the best DNS pair?

## What should I pay attention to?
- Use a different IP (VPN) to go out, need to run once and choose the best DNS to set up
- This version does not provide a test domain list for the time being. You can also create your own domain list that you frequently visit. The tool will help you calculate which DNS is optimal.
- The best DNS obtained by the same access list in different regions (regions) is different.
- By default, China’s DNS list is used for testing and filtering.
- By default, ping is not enabled for testing.

## How use
```
vi yourDomainList.txt
go build -o fastDNS Main.go
./fastDNS -h
./fastDNS -f yourDomainList.txt
```
out like:
```
Top 5 DNS IP addresses: 2001
0       0.001001        223.6.6.232
1       0.035905        223.6.6.133
2       0.036049        223.5.5.124
3       0.036390        223.5.5.204
4       0.039645        223.6.6.81
```

# next version
- Start a DNS server
- Monitor and record all domains (DNS) that your system makes requests for when testing the best DNS
- Automatically recalculate optimal DNS when your exit IP changes
- Caching DNS records for the fastest response to local DNS
- When responding to local DNS, the best result will be selected based on the current egress IP.