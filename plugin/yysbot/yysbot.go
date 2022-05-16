package yysbot

import (
	"github.com/FloatTech/ZeroBot-Plugin/plugin/manager/timer"
	sql "github.com/FloatTech/sqlite"
	"github.com/FloatTech/zbputils/control"
	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
)

var (
	urlRoot = "https://yys.res.netease.com/pc/zt/20170731172708/data/picture/"
	db      = &sql.Sqlite{}
	clock   timer.Clock
)

type pictureInfo struct {
	OriName string `db:"oriName"`
	Num     int    `db:"num"`
}

//下载html页面，更新壁纸链接的数据库
func upDateYysWallPaperDb() {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://yys.163.com/media/picture.html", nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("Accept-Charset", "utf-8")
	req.Header.Set("authority", "yys.163.com")
	req.Header.Set("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	req.Header.Set("accept-language", "zh-CN,zh;q=0.9")
	req.Header.Set("cache-control", "max-age=0")
	req.Header.Set("cookie", "topbarnewsshow=1; _ns=NS1.2.256081119.1652349683; _nietop_foot=%u51B3%u6218%uFF01%u5E73%u5B89%u4EAC%7Cmoba.163.com%2C%u9634%u9633%u5E08%uFF1A%u767E%u95FB%u724C%7Cssr.163.com%2C%u9634%u9633%u5E08%7Cyys.163.com; timing_user_id=time_JLMOLaa5tu")
	req.Header.Set("sec-ch-ua", `" Not A;Brand";v="99", "Chromium";v="101", "Google Chrome";v="101"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"Windows"`)
	req.Header.Set("sec-fetch-dest", "document")
	req.Header.Set("sec-fetch-mode", "navigate")
	req.Header.Set("sec-fetch-site", "none")
	req.Header.Set("sec-fetch-user", "?1")
	req.Header.Set("upgrade-insecure-requests", "1")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/101.0.4951.54 Safari/537.36")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	doc, _ := goquery.NewDocumentFromReader(resp.Body)
	rec := regexp.MustCompile(`^https://yys.res.netease.com/pc/zt/20170731172708/data/picture/(([\d]+)/([\d])/((\d+)x[\d]+)\.jpg)`)
	doc.Find("body > div.main.picture > div.wrap > div.scroll >div>div>div.item>div").Each(func(i int, selection *goquery.Selection) {

		first, _ := selection.Find("div>a:first-child").Attr("href")
		last, _ := selection.Find("div>a:last-child").Attr("href")
		//fmt.Println(first)
		if rec.MatchString(first) && rec.MatchString(last) {
			firstRes := rec.FindStringSubmatch(first)
			lastRes := rec.FindStringSubmatch(last)
			f, _ := strconv.Atoi(firstRes[len(firstRes)-1])
			l, _ := strconv.Atoi(lastRes[len(lastRes)-1])
			if f < l {
				firstRes = lastRes
			}
			info := pictureInfo{OriName: firstRes[1], Num: i}
			//fmt.Println(info)
			db.Insert("picInfo", &info)

		}

	})
}

//
func getImgUrlFromDb() string {
	var picInfo pictureInfo
	num, _ := db.Count("picInfo")
	num = rand.Intn(num)

	db.Find("picInfo", &picInfo, "where num = "+strconv.Itoa(num))
	return picInfo.OriName
}
func init() {

	engine := control.Register("yysbot", &control.Options{
		DisableOnDefault: false,
		Help: "阴阳师群聊插件\n" +
			"- 阴阳师壁纸",
		PrivateDataFolder: "",
		PublicDataFolder:  "YysBot",
		OnEnable:          nil,
		OnDisable:         nil,
	})
	go func() {
		db.DBPath = engine.DataFolder() + "yysWallPaper.db"
		clock = timer.NewClock(db)
		err := db.Create("picInfo", &pictureInfo{})
		if err != nil {
			panic(err)
		}
	}()
	upDateYysWallPaperDb()

	engine.OnFullMatch("阴阳师壁纸", zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		imgUrl := getImgUrlFromDb()
		log.Println(imgUrl)
		ctx.SendChain(message.Image(urlRoot + imgUrl))

	})
	engine.OnFullMatch("更新阴阳师壁纸数据", zero.SuperUserPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		upDateYysWallPaperDb()
		ctx.SendChain(message.Text("阴阳师壁纸数据更新完毕"))
	})

}
