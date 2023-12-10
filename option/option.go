package option

import (
	"embed"
	"flag"
	"fmt"
	util "github.com/hktalent/go-utils"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type Options struct {
	EnablePing    bool // 开启ping
	PingCnt       int
	PingTimeout   time.Duration
	TopPingDnsNum int //
	OutTopNum     int //
	Thread        int //
	DomainFile    string
	TestStop      int64
	Type          string // 测试模式，默认为 0 = cn，1 = us，3 = all
}

var DnsList = `af	Afghanistan
al	Albania
dz	Algeria
as	American Samoa
ad	Andorra
ao	Angola
aq	Antarctica
ag	Antigua and Barbuda
ar	Argentina
am	Armenia
aw	Aruba
au	Australia
at	Austria
az	Azerbaijan
bs	Bahamas
bh	Bahrain
bd	Bangladesh
bb	Barbados
by	Belarus
be	Belgium
bz	Belize
bj	Benin
bm	Bermuda
bt	Bhutan
bo	Bolivia, Plurinational State of
bq	Bonaire, Sint Eustatius and Saba
ba	Bosnia and Herzegovina
bw	Botswana
br	Brazil
bn	Brunei Darussalam
bg	Bulgaria
bf	Burkina Faso
bi	Burundi
kh	Cambodia
cm	Cameroon
ca	Canada
cv	Cape Verde
ky	Cayman Islands
td	Chad
cl	Chile
cn	China
co	Colombia
cg	Congo
cd	Congo, The Democratic Republic of the
cr	Costa Rica
hr	Croatia
cu	Cuba
cy	Cyprus
cz	Czech Republic
ci	Côte d&#39;Ivoire
dk	Denmark
do	Dominican Republic
ec	Ecuador
eg	Egypt
sv	El Salvador
gq	Equatorial Guinea
ee	Estonia
et	Ethiopia
fi	Finland
fr	France
gf	French Guiana
pf	French Polynesia
ga	Gabon
ge	Georgia
de	Germany
gh	Ghana
gi	Gibraltar
gr	Greece
gl	Greenland
gp	Guadeloupe
gu	Guam
gt	Guatemala
gg	Guernsey
gn	Guinea
hn	Honduras
hk	Hong Kong
hu	Hungary
is	Iceland
in	India
id	Indonesia
ir	Iran, Islamic Republic of
iq	Iraq
ie	Ireland
im	Isle of Man
il	Israel
it	Italy
jm	Jamaica
jp	Japan
je	Jersey
jo	Jordan
kz	Kazakhstan
ke	Kenya
kr	Korea, Republic of
kw	Kuwait
kg	Kyrgyzstan
la	Lao People&#39;s Democratic Republic
lv	Latvia
lb	Lebanon
lr	Liberia
ly	Libya
li	Liechtenstein
lt	Lithuania
lu	Luxembourg
mo	Macao
mk	Macedonia, Republic of
mg	Madagascar
mw	Malawi
my	Malaysia
mv	Maldives
ml	Mali
mt	Malta
mh	Marshall Islands
mq	Martinique
mr	Mauritania
mu	Mauritius
yt	Mayotte
mx	Mexico
md	Moldova, Republic of
mc	Monaco
mn	Mongolia
me	Montenegro
ma	Morocco
mz	Mozambique
mm	Myanmar
na	Namibia
np	Nepal
nl	Netherlands
nc	New Caledonia
nz	New Zealand
ni	Nicaragua
ne	Niger
ng	Nigeria
no	Norway
om	Oman
pk	Pakistan
pw	Palau
ps	Palestine, State of
pa	Panama
pg	Papua New Guinea
py	Paraguay
pe	Peru
ph	Philippines
pl	Poland
pt	Portugal
pr	Puerto Rico
qa	Qatar
ro	Romania
ru	Russian Federation
rw	Rwanda
re	Réunion
vc	Saint Vincent and the Grenadines
sa	Saudi Arabia
sn	Senegal
rs	Serbia
sc	Seychelles
sl	Sierra Leone
sg	Singapore
sk	Slovakia
si	Slovenia
sb	Solomon Islands
so	Somalia
za	South Africa
es	Spain
lk	Sri Lanka
sd	Sudan
sz	Swaziland
se	Sweden
ch	Switzerland
sy	Syrian Arab Republic
tw	Taiwan, Province of China
tj	Tajikistan
tz	Tanzania, United Republic of
th	Thailand
tl	Timor-Leste
tg	Togo
tt	Trinidad and Tobago
tn	Tunisia
tr	Turkey
ug	Uganda
ua	Ukraine
ae	United Arab Emirates
gb	United Kingdom
us	United States
uy	Uruguay
uz	Uzbekistan
ve	Venezuela, Bolivarian Republic of
vn	Viet Nam
vi	Virgin Islands, U.S.
xk	XK
ye	Yemen
zm	Zambia
zw	Zimbabwe
ax	Åland Islands`

func WaitOneFunc4WgParmsChan[T any](wg *sync.WaitGroup, cbk func(x T), cT chan struct{}, parms ...T) {
	for _, x := range parms {
		cT <- struct{}{}
		wg.Add(1)
		go func(p1 T) {
			defer func() {
				wg.Done()
				<-cT
			}()
			cbk(p1)
		}(x)
	}
}
func UpAllList(opt *Options, dnsDir *embed.FS) {
	a := strings.Split(DnsList, "\n")
	a = append(a, "nameservers\t")
	var wg sync.WaitGroup
	var ct = make(chan struct{}, opt.Thread)
	szUP := "https://public-dns.info/nameserver/%s.txt"
	szPwd, _ := os.Getwd()
	WaitOneFunc4WgParmsChan[string](&wg, func(x string) {
		if x = strings.TrimSpace(x); "" != x {
			if x1 := strings.Split(x, "\t"); 0 < len(x1) {
				util.DoUrlCbk(fmt.Sprintf(szUP, x1[0]), "", nil, func(resp *http.Response, szUrl string) {
					if data, err := io.ReadAll(resp.Body); nil == err {
						szU1 := szPwd + "/data/" + x1[0] + ".txt"
						if nil == os.WriteFile(szU1, data, os.ModePerm) {
							log.Println("ok", szU1, len(data))
						}
					}
				})
			}
		}

	}, ct, a...)
	wg.Wait()
}

func ParseOption(dnsDir *embed.FS) *Options {
	var opt = Options{}
	var tmot int64
	bUpdate := false
	flag.BoolVar(&bUpdate, "u", false, "update all dns data")
	flag.BoolVar(&opt.EnablePing, "p", false, "Enable Ping,default:false")
	flag.Int64Var(&tmot, "o", 50, "ping top dns num:50")
	flag.IntVar(&opt.TopPingDnsNum, "k", 200, "ping top dns num:200")
	flag.IntVar(&opt.PingCnt, "c", 2, "ping count:2")
	flag.StringVar(&opt.Type, "m", "cn", "nameservers = all,\n"+DnsList+"\ndefault: cn")
	flag.IntVar(&opt.OutTopNum, "n", 5, "out top dns num:5")
	flag.Int64Var(&opt.TestStop, "y", 1000, "Test Stop dns num:1000")
	flag.IntVar(&opt.Thread, "t", 1024, "Thread num:1024")
	flag.StringVar(&opt.DomainFile, "f", "data/zq/testDomain.txt", "test Domain file name")

	flag.Parse()
	opt.PingTimeout = time.Duration(tmot)
	if bUpdate {
		UpAllList(&opt, dnsDir)
		os.Exit(0)
	}

	return &opt
}
