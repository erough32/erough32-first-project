package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var ReadTags []string

func loginApi(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json;charset=utf-8")
	var r map[string]string
	user := c.PostForm("user")
	pass := GetSHA256HashCode([]byte(c.PostForm("pass")))

	session := sessions.Default(c)

	loginUid := loginSelect(user, pass)
	if loginUid == 0 {
		r = map[string]string{"status": "error"}
	} else if loginUid == -1 {
		r = map[string]string{"status": "danger"}
	} else {
		tokenInfo := getUserInfo(loginUid)
		if tokenInfo != nil {
			session.Set("token", tokenInfo)
			session.Set("bid", loginUid)
			session.Save()
		}
		r = map[string]string{"status": "ok", "SESSION": "ok"}
	}
	c.JSON(http.StatusOK, r)
}

func loginApiv2(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json;charset=utf-8")
	var r map[string]string
	user := c.PostForm("user")
	pass := c.PostForm("pass")

	session := sessions.Default(c)

	loginUid := loginSelect(user, pass)
	if loginUid == 0 {
		r = map[string]string{"status": "error"}
	} else if loginUid == -1 {
		r = map[string]string{"status": "danger"}
	} else {
		tokenInfo := getUserInfo(loginUid)
		if tokenInfo != nil {
			session.Set("token", tokenInfo)
			session.Set("bid", loginUid)
			session.Save()
		}
		r = map[string]string{"status": "ok", "SESSION": "ok"}
	}
	c.JSON(http.StatusOK, r)
}

func newRegApi(c *gin.Context) {
	user := c.PostForm("user")
	pass := GetSHA256HashCode([]byte(c.PostForm("pass")))
	email := c.PostForm("email")
	uhash := c.PostForm("uhash")
	captchaToken := c.PostForm("captcha")

	params := url.Values{"secret": {"6Ld1uf4cAAAAAOstl042WpziMNs9V8DigZslRX3g"}, "response": {captchaToken}}
	resp, _ := http.PostForm("https://www.google.com/recaptcha/api/siteverify", params)
	body, _ := ioutil.ReadAll(resp.Body)

	var captchaJson map[string]interface{}
	json.Unmarshal(body, &captchaJson)

	if captchaJson["success"] != true {
		c.String(http.StatusOK, "error:captcha")
		return
	}

	if SQLInject(user) || SQLInject(uhash) || SQLInject(email) {
		c.String(http.StatusOK, "error:sql")
		return
	}
	if !VerifyEmailFormat(email) {
		c.String(http.StatusOK, "error:email")
		return
	}
	fuid := selectYQ(uhash)
	if fuid == 0 {
		c.String(http.StatusOK, "error:nuhash")
		return
	}
	if searchUser(user) != 0 {
		c.String(http.StatusOK, "error:username")
		return
	}
	regSql(user, pass, email, fuid)
	c.String(http.StatusOK, "ok")
}

func duplicateUsername(c *gin.Context) {
	username := c.PostForm("username")

	c.Writer.Header().Set("Content-Type", "application/json;charset=utf-8")
	if SQLInject(username) {
		c.JSON(200, gin.H{"status": "sql"})
		return
	}
	if searchUser(username) != 0 {
		c.JSON(200, gin.H{"status": "duplicate"})
	} else {
		c.JSON(200, gin.H{"status": "ok"})
	}
}

func selectSession(c *gin.Context) {
	session := sessions.Default(c)
	info := session.Get("token")
	if info != nil {
		c.String(http.StatusOK, "ok")
	} else {
		c.String(http.StatusOK, "error")
	}
}

func userselect(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json;charset=utf-8")
	session := sessions.Default(c)
	jsonInfo, _ := json.Marshal(session.Get("token"))
	c.String(http.StatusOK, string(jsonInfo))
}

func BooklistPage(c *gin.Context) {
	session := sessions.Default(c)
	info := session.Get("token")
	if info != nil {
		pageinfo := c.PostForm("page")
		if pageinfo == "" {
			pageinfo = "1"
		}
		pagenum, _ := strconv.Atoi(pageinfo)

		c.Writer.Header().Set("Content-Type", "application/json;charset=utf-8")

		re := bookList(pagenum)
		if re == "]" {
			re = `["Error"]`
		}

		session := sessions.Default(c)
		addUser(session.Get("token").([]string)[0])

		c.String(http.StatusOK, re)
	} else {
		c.String(http.StatusOK, `["login"]`)
	}
}

func Booknumpage(c *gin.Context) {
	c.String(http.StatusOK, fmt.Sprint(booknum()))
}

func booklistV2(c *gin.Context) {
	session := sessions.Default(c)
	info := session.Get("token")
	c.Writer.Header().Set("Content-Type", "application/json;charset=utf-8")
	if info != nil {
		pageinfo := c.PostForm("page")
		if pageinfo == "" {
			pageinfo = "1"
		}
		pagenum, _ := strconv.Atoi(pageinfo)

		jsonInfo, count := bookListSqlV2(pagenum)
		if topbook != -1 && pagenum == 1 {
			bookinfo := bidGetBook(fmt.Sprint(topbook))
			bookinfo[0].Bookname = "[置頂]" + bookinfo[0].Bookname
			jsonInfo = append(jsonInfo, bookinfo[0])
		}

		session := sessions.Default(c)
		addUser(session.Get("token").([]string)[0])

		c.JSON(http.StatusOK, gin.H{"info": jsonInfo, "len": count})
	} else {
		c.String(http.StatusOK, `["login"]`)
	}
}

func searchNew(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json;charset=utf-8")
	session := sessions.Default(c)
	info := session.Get("token")
	if info != nil {
		info := c.PostForm("info")
		if info == "" || SQLInject(info) {
			c.JSON(200, gin.H{"status": "nof"})
			return
		}
		searchType := c.PostForm("type")
		sortType := c.PostForm("sort")
		pageinfo := c.PostForm("page")
		btags := c.PostForm("btags")
		if pageinfo == "" {
			pageinfo = "1"
		}
		pagenum, _ := strconv.Atoi(pageinfo)
		searchInfo := searchBookNew(info, pagenum, searchType, sortType, btags)
		c.JSON(200, searchInfo)
	} else {
		c.JSON(200, gin.H{"status": "login"})
	}
}

func recommend(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json;charset=utf-8")
	session := sessions.Default(c)
	info := session.Get("token")
	if info != nil {
		bid := c.PostForm("bid")
		if bid == "" {
			bid = "0"
		}

		booklistinfo := getSimilar(bid)
		jsonStr := make([]map[string]string, len(booklistinfo))
		for i := 0; i < len(booklistinfo); i++ {
			jsonStr[i] = map[string]string{"bid": fmt.Sprint(booklistinfo[i].Bid), "bookname": booklistinfo[i].Bookname}
		}

		c.JSON(http.StatusOK, jsonStr)
	} else {
		c.String(http.StatusOK, `["login"]`)
	}
}

func getBookInfo(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json;charset=utf-8")
	session := sessions.Default(c)
	info := session.Get("token")
	if info != nil {
		bid := c.PostForm("bid")
		uid := session.Get("token").([]string)[0]
		bidint, _ := strconv.Atoi(bid)

		fbool := intInSlice(bidint, favoritesSelectSql(uid))

		if bid != "" {
			bookinfo := bidGetBook(bid)
			if len(bookinfo) != 0 {
				bookinfothis := bookinfo[0]
				bookinfothis.F = fbool
				c.JSON(http.StatusOK, bookinfothis)
			} else {
				bookinfo := bidGetBook("1")[0]
				bookinfo.F = false
				c.JSON(http.StatusOK, bookinfo)
			}
		} else {
			bookinfo := bidGetBook("1")[0]
			bookinfo.F = false
			c.JSON(http.StatusOK, bookinfo)
		}
	} else {
		c.String(http.StatusOK, `["login"]`)
	}
}

func seo(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json;charset=utf-8")
	bid := c.PostForm("bid")

	var bookinfo []Booktype

	bidInt, _ := strconv.Atoi(bid)
	err := Db.Select(&bookinfo, "SELECT * FROM book WHERE bid = ?", bidInt)
	if err != nil {
		fmt.Println(err)
	}

	if len(bookinfo) != 0 {
		bookinfo[0].F = false
		c.JSON(http.StatusOK, bookinfo[0])
	} else {
		c.JSON(http.StatusOK, Booktype{})
	}
}

func adfavorites(c *gin.Context) {
	session := sessions.Default(c)
	info := session.Get("token")
	if info != nil {
		bid := c.PostForm("bid")
		uid := session.Get("token").([]string)[0]

		d := favoritesADSql(bid, uid)
		setFav(bid, d)
		c.String(http.StatusOK, `ok`)
	} else {
		c.String(http.StatusOK, `login`)
	}
}

func randomBook(c *gin.Context) {
	session := sessions.Default(c)
	info := session.Get("token")
	c.Writer.Header().Set("Content-Type", "application/json;charset=utf-8")
	if info != nil {
		booklist := getRandom(20)
		c.JSON(200, booklist)
	} else {
		c.JSON(200, gin.H{"status": "login"})
	}
}

func favoriteslist(c *gin.Context) {
	session := sessions.Default(c)
	info := session.Get("token")
	if info != nil {
		uid := session.Get("token").([]string)[0]
		pageinfo := c.PostForm("page")
		if pageinfo == "" {
			pageinfo = "1"
		}
		pagenum, _ := strconv.Atoi(pageinfo)

		c.Writer.Header().Set("Content-Type", "application/json;charset=utf-8")

		re := favoritesbookList(uid)
		if re == nil {
			c.String(http.StatusOK, `["nof"]`)
			return
		}

		lenre := len(re)
		reverseRe := []string{}
		for i := range re {
			reverseRe = append([]string{fmt.Sprint(re[i])}, reverseRe...)
		}

		if pagenum*20 > lenre {
			reverseRe = reverseRe[(pagenum-1)*20 : lenre]
		} else {
			reverseRe = reverseRe[(pagenum-1)*20 : pagenum*20]
		}

		booklist2 := getBookListByList(strings.Join(reverseRe, ","))
		c.JSON(http.StatusOK, booklist2)
	} else {
		c.String(http.StatusOK, `["login"]`)
	}
}

func favoriteslistV2(c *gin.Context) {
	session := sessions.Default(c)
	info := session.Get("token")
	if info != nil {
		uid := session.Get("token").([]string)[0]
		pageinfo := c.PostForm("page")
		if pageinfo == "" {
			pageinfo = "1"
		}
		pagenum, _ := strconv.Atoi(pageinfo)

		c.Writer.Header().Set("Content-Type", "application/json;charset=utf-8")

		re := favoritesbookList(uid)
		if re == nil {
			c.String(http.StatusOK, `["nof"]`)
			return
		}

		lenre := len(re)
		reverseRe := []string{}
		for i := range re {
			reverseRe = append([]string{fmt.Sprint(re[i])}, reverseRe...)
		}

		if pagenum*20 > lenre {
			reverseRe = reverseRe[(pagenum-1)*20 : lenre]
		} else {
			reverseRe = reverseRe[(pagenum-1)*20 : pagenum*20]
		}

		booklist2 := getBookListByList(strings.Join(reverseRe, ","))
		c.JSON(http.StatusOK, map[string]interface{}{"info": booklist2, "len": lenre})
	} else {
		c.String(http.StatusOK, `["login"]`)
	}
}

func favoritesnum(c *gin.Context) {
	session := sessions.Default(c)
	info := session.Get("token")
	if info != nil {
		c.String(http.StatusOK, fmt.Sprint(favoritesbookListnum(session.Get("token").([]string)[0])))
	} else {
		c.String(http.StatusOK, `["login"]`)
	}
}

func history(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json;charset=utf-8")
	session := sessions.Default(c)
	info := session.Get("token")
	if info != nil {
		bidlist := c.PostForm("list")
		var bookidList []string
		var bookinfoList []Booktype
		json.Unmarshal([]byte(bidlist), &bookidList)
		bookinfoList = getBookListByList(strings.Join(bookidList, ","))

		c.JSON(http.StatusOK, bookinfoList)
	} else {
		c.String(http.StatusOK, `["login"]`)
	}
}

func historyV2(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json;charset=utf-8")
	session := sessions.Default(c)
	info := session.Get("token")
	if info != nil {
		bidlist := c.PostForm("list")
		var bookidList []string
		var bookinfoList []Booktype
		json.Unmarshal([]byte(bidlist), &bookidList)
		if len(bookidList) < 50 {
			bookinfoList = getBookListByList(strings.Join(bookidList, ","))
			bookMap := make(map[int]Booktype)
			for i := 0; i < len(bookinfoList); i++ {
				bookMap[bookinfoList[i].Bid] = bookinfoList[i]
			}
			for i := range bookidList {
				bidNum, _ := strconv.Atoi(bookidList[i])
				if _, ok := bookMap[bidNum]; !ok {
					bookMap[bidNum] = Booktype{Bid: 1, Bookname: "不存在的內容", Author: "NoyAcg", Pname: "Tips", Otag: "Tips", Ptag: "Tips", Time: 0, Views: 0, Favorites: 0, F: false, Len: 0}
				}
			}
			c.JSON(http.StatusOK, bookMap)
		} else {
			c.String(http.StatusOK, `{"status":"toolong"}`)
		}
	} else {
		c.String(http.StatusOK, `["login"]`)
	}
}

func historyV3(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json;charset=utf-8")
	session := sessions.Default(c)
	info := session.Get("token")
	if info != nil {
		bidlist := c.PostForm("list")
		var bookidList []string
		json.Unmarshal([]byte(bidlist), &bookidList)
		if len(bookidList) < 50 {
			bookinfoList := getBookListByList(strings.Join(bookidList, ","))
			c.JSON(http.StatusOK, bookinfoList)
		} else {
			c.String(http.StatusOK, `{"status":"toolong"}`)
		}
	} else {
		c.String(http.StatusOK, `["login"]`)
	}
}

func ftag(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json;charset=utf-8")
	session := sessions.Default(c)
	info := session.Get("token")
	if info != nil {
		var thislist []string
		var rendList []int
		var ftaglen int = len(ReadTags)
		if ftaglen > 20 {
			rendList = generateRandomNumber(0, ftaglen, 20)
		} else {
			rendList = generateRandomNumber(0, ftaglen, ftaglen)
		}
		for i := range rendList {
			thislist = append(thislist, ReadTags[rendList[i]])
		}
		c.JSON(http.StatusOK, thislist)
	} else {
		c.String(http.StatusOK, `["login"]`)
	}
}

func getbooklen(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json;charset=utf-8")
	token := sessions.Default(c).Get("token")
	if token != nil {
		bid := c.PostForm("bid")
		bookinfo := bidGetBook(bid)

		bidInt, _ := strconv.Atoi(bid)

		jsonInfo, _ := json.Marshal(token)
		var jsonInfoObj []string
		json.Unmarshal(jsonInfo, &jsonInfoObj)

		if len(bookinfo) != 0 && readLimit(bidInt, jsonInfoObj[0]) {
			addRead(bid)
			addReadLimit(bidInt, jsonInfoObj[0])
		}

		c.JSON(200, gin.H{"status": "ok", "len": fmt.Sprint(bookinfo[0].Len), "bookname": fmt.Sprint(bookinfo[0].Bookname)})
	}
}

func getUserInfoV2(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json;charset=utf-8")
	token := sessions.Default(c).Get("token")
	if token != nil {
		jsonInfo, _ := json.Marshal(token)
		var jsonInfoObj []string
		json.Unmarshal(jsonInfo, &jsonInfoObj)
		c.JSON(200, getDbUserInfo(jsonInfoObj[0]))
	} else {
		c.JSON(200, gin.H{"status": "login"})
	}
}

func getMsg(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json;charset=utf-8")
	token := sessions.Default(c).Get("token")
	if token != nil {
		page := c.PostForm("page")
		pageInt, err := strconv.Atoi(page)
		if err != nil {
			pageInt = 1
		}

		jsonInfo, _ := json.Marshal(token)
		var jsonInfoObj []string
		json.Unmarshal(jsonInfo, &jsonInfoObj)

		uidInt, _ := strconv.Atoi(jsonInfoObj[0])

		var data struct {
			Bid    int `db:"bid"`
			Reply  int `db:"reply"`
			Sysmsg int `db:"sysmsg"`
		}
		mongoCollectionMsg.FindOne(ctx, bson.D{{Key: "uid", Value: uidInt}}).Decode(&data)

		c.JSON(200, gin.H{
			"status": "ok",
			"sysmsg": getSysMsg(jsonInfoObj[0], pageInt, data.Sysmsg),
			"reply":  getCommentReply(jsonInfoObj[0], pageInt, data.Reply),
		})
	} else {
		c.JSON(200, gin.H{"status": "login"})
	}
}

func useI(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json;charset=utf-8")
	token := c.Query("token")
	if token == GetSHA256HashCode([]byte("imagelib"+fmt.Sprintf(`"%s"`, time.Now().Format("20060102")))) {
		uid := c.Query("uid")
		uidInt, _ := strconv.Atoi(uid)
		num := c.Query("num")

		c.JSON(200, useISql(uidInt, num))
	} else {
		c.JSON(200, gin.H{"status": "token_error"})
	}
}

func getDownload(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json;charset=utf-8")
	token := sessions.Default(c).Get("token")
	if token != nil {
		jsonInfo, _ := json.Marshal(token)
		var jsonInfoObj []string
		json.Unmarshal(jsonInfo, &jsonInfoObj)
		bid := c.PostForm("bid")
		if bid != "" {
			c.JSON(200, gin.H{"status": "ok", "uid": jsonInfoObj[0], "token": GetSHA256HashCode([]byte(jsonInfoObj[0] + bid + fmt.Sprint(time.Now().Format("2006-01-02")) + "noyo"))})
		} else {
			c.JSON(200, gin.H{"status": "bid_not_found"})
		}
	} else {
		c.JSON(200, gin.H{"status": "login"})
	}
}

func bulletin(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json;charset=utf-8")
	token := sessions.Default(c).Get("token")
	if token != nil {
		var jsonInfo map[string]map[string]string
		json.Unmarshal(ReadAll("./config/bulletin.json"), &jsonInfo)
		switch c.PostForm("type") {
		case "app":
			c.JSON(200, jsonInfo["app"])
		case "web":
			c.JSON(200, jsonInfo["web"])
		default:
			c.JSON(200, gin.H{"status": "notfound"})
		}
	} else {
		c.JSON(200, gin.H{"status": "loginout"})
	}
}

func getComment(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json;charset=utf-8")
	session := sessions.Default(c)
	info := session.Get("token")
	if info != nil {

		bid := c.PostForm("bid")
		page := c.PostForm("page")
		bidInt, err := strconv.Atoi(bid)
		if err != nil {
			c.JSON(200, gin.H{"status": "bid_error"})
			return
		}
		pageInt, err := strconv.Atoi(page)
		if err != nil {
			pageInt = 1
		}
		pageTo := (pageInt - 1) * 40
		commInfo := getCommentDb(bidInt, pageTo)
		commLen := getCommentLen(bidInt)

		over := false

		// 評論置頂
		if pageInt == 1 && commtop.Show {
			commInfo = append([]CommInfoUser{{
				Cid:       1,
				Uid:       0,
				Username:  commtop.Username,
				AvatarUrl: "",
				EmailMd5:  commtop.EmailMd5,
				Area:      commtop.Type,
				Reply:     -1,
				ReplyText: "",
				Platform:  "web",
				Time:      commtop.Time,
				Content:   commtop.Content,
			}}, commInfo...)
			commLen++
		}

		if pageInt*40 > commLen {
			over = true
		}

		if commLen != 0 {
			c.JSON(200, gin.H{"status": "ok", "info": commInfo, "len": commLen, "over": over})
		} else {
			c.JSON(200, gin.H{"status": "ok", "info": []string{}, "len": commLen, "over": true})
		}
	} else {
		c.JSON(200, gin.H{"status": "login"})
	}
}

func sendComment(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json;charset=utf-8")
	session := sessions.Default(c)
	info := session.Get("token")
	if info != nil {
		userInfo := getDbUserInfo(info.([]string)[0])

		bid, _ := strconv.Atoi(c.PostForm("bid"))
		content := c.PostForm("content")
		platform := c.PostForm("platform")
		reply, _ := strconv.Atoi(c.PostForm("reply"))
		emailMd5 := md5V(userInfo.Email)
		ip := c.PostForm("user_ip")
		if ip == "" {
			ip = c.ClientIP()
		}
		area := getIpArea(ip)

		if lastCommentTime(userInfo.Uid) < time.Now().Unix()-10 {
			if bid != 0 && content != "" && platform != "" && reply != 0 {
				if isSensitive(content) {
					c.JSON(200, gin.H{"status": "sensitive"})
				} else {
					go func() {
						pushCommToGroup("新評論推送~\n本子連結：https://web.noy.asia/book/"+fmt.Sprint(bid)+"\n\n"+userInfo.Username+": "+content, "")
					}()
					dbStatus := sendCommentDb(bid, userInfo.Uid, userInfo.Username, userInfo.Email, emailMd5, ip, area, reply, platform, content)
					c.JSON(200, gin.H{"status": dbStatus})
				}
			} else {
				c.JSON(200, gin.H{"status": "c_error"})
			}
		} else {
			c.JSON(200, gin.H{"status": "too_fast"})
		}

	} else {
		c.JSON(200, gin.H{"status": "login"})
	}
}

func update(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json;charset=utf-8")
	c.JSON(200, updateInfo)
}

func ads(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json;charset=utf-8")
	c.JSON(200, AdsJson[0])
}

func AdsV2(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json;charset=utf-8")
	c.JSON(200, AdsJson)
}

func Cron(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json;charset=utf-8")
	ip := strings.Split(c.ClientIP(), ".")
	if ip[0] == "127" {
		mode := c.Query("mode")
		switch mode {
		case "getReadTag":
			ReadTags = getReadTags()
		case "loadConfig":
			loadConfig()
		case "resetReadFrequency":
			readFrequencyLimit = make(map[int]map[string]int64)
		}
		c.JSON(200, gin.H{"status": 200})
	} else {
		c.JSON(200, gin.H{"status": 403})
	}
}

func homeApi(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json;charset=utf-8")
	session := sessions.Default(c)
	info := session.Get("token")
	if info != nil {
		uidStr := info.([]string)[0]
		addUser(uidStr)
		uid, _ := strconv.Atoi(uidStr)

		fs, _ := getFs(uid, 1)
		c.JSON(200, gin.H{
			"status":     "ok",
			"readDay":    getReadLeaderboardBook("day", 0),
			"favDay":     getFavLeaderboardBook("day", 0),
			"fs":         fs,
			"tags":       configTags,
			"bulletin":   bulletinInfo,
			"proportion": getProportion(1),
		})
	} else {
		c.JSON(200, gin.H{"status": "login"})
	}
}

func proportion(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json;charset=utf-8")
	session := sessions.Default(c)
	info := session.Get("token")
	if info != nil {
		pageNum, _ := strconv.Atoi(c.PostForm("page"))

		c.JSON(200, gin.H{"status": "ok", "info": getProportion(pageNum), "len": getProportionCount()})
	} else {
		c.JSON(200, gin.H{"status": "login"})
	}
}

func getReadLeaderboard(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json;charset=utf-8")
	session := sessions.Default(c)
	info := session.Get("token")
	if info != nil {
		lbType := c.PostForm("type")
		pageNum, _ := strconv.Atoi(c.PostForm("page"))
		pageNum = (pageNum - 1) * 20
		count := getReadLeaderboardCount(lbType)

		c.JSON(200, gin.H{"status": "ok", "info": getReadLeaderboardBook(lbType, pageNum), "len": count})
	} else {
		c.JSON(200, gin.H{"status": "login"})
	}
}

func getFavLeaderboard(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json;charset=utf-8")
	session := sessions.Default(c)
	info := session.Get("token")
	if info != nil {
		lbType := c.PostForm("type")
		pageNum, _ := strconv.Atoi(c.PostForm("page"))
		pageNum = (pageNum - 1) * 20
		count := getFavLeaderboardCount()

		c.JSON(200, gin.H{"status": "ok", "info": getFavLeaderboardBook(lbType, pageNum), "len": count})
	} else {
		c.JSON(200, gin.H{"status": "login"})
	}
}

func favoritesrecommend(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json;charset=utf-8")
	session := sessions.Default(c)
	info := session.Get("token")
	if info != nil {
		pageNum := c.PostForm("page")
		if pageNum == "" {
			pageNum = "1"
		}
		pageInt, _ := strconv.Atoi(pageNum)
		uid, _ := strconv.Atoi(info.([]string)[0])

		fs, fslen := getFs(uid, pageInt)
		c.JSON(200, gin.H{"status": "ok", "info": fs, "len": fslen})
	} else {
		c.JSON(200, gin.H{"status": "login"})
	}
}

func downloadCreditApi(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json;charset=utf-8")
	token := c.Query("token")
	if token == GetSHA256HashCode([]byte("imagelib"+fmt.Sprintf(`"%s"`, time.Now().Format("20060102")))) {
		uid := c.Query("uid")
		uidInt, _ := strconv.Atoi(uid)
		num := c.Query("num")
		numInt, _ := strconv.Atoi(num)
		bid := c.Query("bid")
		bidInt, _ := strconv.Atoi(bid)

		type mongoDataType struct {
			Uid     int   `db:"uid"`
			BidList []int `db:"bidlist"`
		}

		cur, err := mongoCollectionDownloads.Find(ctx, bson.M{"uid": uidInt, "bidlist": bson.M{"$in": []int{bidInt}}})
		if err != nil {
			fmt.Println("MongoErr", err)
			c.JSON(200, useISql(uidInt, fmt.Sprintf("%.2f", float64(numInt)/float64(downloadCredit))))
			return
		}
		defer cur.Close(ctx)

		var datalist []int = []int{}
		for cur.Next(ctx) {
			var data mongoDataType
			err := cur.Decode(&data)
			if err != nil {
				log.Fatal(err)
			}

			datalist = append(datalist, data.Uid)
		}

		if len(datalist) != 0 {
			c.JSON(200, useISql(uidInt, "0"))
		} else {
			c.JSON(200, useISql(uidInt, fmt.Sprintf("%.2f", float64(numInt)/float64(downloadCredit))))
			mongoCollectionDownloads.UpdateOne(ctx, bson.M{"uid": uidInt}, bson.M{"$push": bson.M{"bidlist": bidInt}}, options.Update().SetUpsert(true))
		}
	} else {
		c.JSON(200, gin.H{"status": "token_error"})
	}
}

func getcredit(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json;charset=utf-8")
	session := sessions.Default(c)
	info := session.Get("token")
	if info != nil {
		bid := c.PostForm("bid")
		userinfo := getDbUserInfo(info.([]string)[0])

		if bid != "" {
			type mongoDataType struct {
				Uid     int   `db:"uid"`
				BidList []int `db:"bidlist"`
			}

			bidInt, _ := strconv.Atoi(bid)

			cur, err := mongoCollectionDownloads.Find(ctx, bson.M{"uid": userinfo.Uid, "bidlist": bson.M{"$in": []int{bidInt}}})
			if err != nil {
				fmt.Println("MongoErr", err)
				cint, _ := strconv.Atoi(httpPost("https://img.noyteam.online/api/getcredit", "bid="+bid))
				c.JSON(200, gin.H{"useri": userinfo.Integral, "booki": fmt.Sprint(float64(cint) / float64(downloadCredit))})
				return
			}
			defer cur.Close(ctx)

			var datalist []int = []int{}
			for cur.Next(ctx) {
				var data mongoDataType
				err := cur.Decode(&data)
				if err != nil {
					log.Fatal(err)
				}

				datalist = append(datalist, data.Uid)
			}

			if len(datalist) != 0 {
				c.JSON(200, gin.H{"useri": userinfo.Integral, "booki": "0"})
			} else {
				cint, _ := strconv.Atoi(httpPost("https://img.noyteam.online/api/getcredit", "bid="+bid))
				c.JSON(200, gin.H{"useri": userinfo.Integral, "booki": fmt.Sprint(float64(cint) / float64(downloadCredit))})
			}
		} else {
			c.JSON(200, gin.H{"useri": userinfo.Integral})
		}

	} else {
		c.JSON(200, gin.H{"status": "login"})
	}
}

func bigtaglist(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json;charset=utf-8")
	session := sessions.Default(c)
	info := session.Get("token")
	if info != nil {
		type bigtaglist struct {
			Tag   string `json:"tag"`
			Cover int    `json:"cover"`
		}

		type mongodata struct {
			Name  string   `db:"name"`
			Tags  []string `db:"tags"`
			Cover int      `db:"cover"`
		}

		cur, err := mongoCollectionBigTag.Find(ctx, bson.D{})
		if err != nil {
			c.JSON(200, gin.H{"status": "error", "data": []bigtaglist{}})
		}

		var bigtag []bigtaglist = []bigtaglist{}
		for cur.Next(ctx) {
			var data mongodata
			err := cur.Decode(&data)
			if err != nil {
				log.Fatal(err)
			}

			bigtag = append(bigtag, bigtaglist{Tag: data.Name, Cover: data.Cover})
		}

		c.JSON(200, gin.H{"status": "error", "data": bigtag})
	} else {
		c.JSON(200, gin.H{"status": "login"})
	}
}

func feedback(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json;charset=utf-8")
	token := sessions.Default(c).Get("token")
	if token != nil {
		jsonInfo, _ := json.Marshal(token)
		var jsonInfoObj []string
		json.Unmarshal(jsonInfo, &jsonInfoObj)
		userinfo := getDbUserInfo(jsonInfoObj[0])
		userip := c.ClientIP()
		userarea := getIpArea(userip)

		var infoJson map[string]string
		info := c.PostForm("info")
		json.Unmarshal([]byte(info), &infoJson)
		infoMsg := ""
		for i, j := range infoJson {
			infoMsg += i + ": " + j + "\n"
		}

		msg := c.PostForm("msg")

		go func() {
			pushCommToGroup("#Feedback\nUID: "+fmt.Sprint(userinfo.Uid)+"   Username: "+userinfo.Username+"\nIP: "+userip+"   Area: "+userarea+"\nTime: "+time.Now().Format("2006-01-02 15:04:05")+"\n\n"+infoMsg+"\n"+msg, "-614850819")
		}()
		c.JSON(200, gin.H{"status": "ok"})
	} else {
		c.JSON(200, gin.H{"status": "login"})
	}
}

func resetUsername(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json;charset=utf-8")
	token := sessions.Default(c).Get("token")
	if token != nil {
		jsonInfo, _ := json.Marshal(token)
		var jsonInfoObj []string
		json.Unmarshal(jsonInfo, &jsonInfoObj)

		newname := c.PostForm("username")

		if SQLInject(newname) {
			c.JSON(200, gin.H{"status": "sql"})
		}

		if searchUser(newname) != 0 {
			c.JSON(200, gin.H{"status": "duplicate"})
		} else {
			uidInt, _ := strconv.Atoi(jsonInfoObj[0])
			if useISql(uidInt, "10")["status"] == "too_few" {
				c.JSON(200, gin.H{"status": "too_few"})
			} else {
				jsonInfo, _ := json.Marshal(token)
				var jsonInfoObj []string
				json.Unmarshal(jsonInfo, &jsonInfoObj)
				userinfo := getDbUserInfo(jsonInfoObj[0])

				rows, _ := Db.Query("UPDATE USER SET username=? WHERE uid = ?", newname, userinfo.Uid)
				defer rows.Close()

				c.JSON(200, gin.H{"status": "ok"})
			}
		}
	} else {
		c.JSON(200, gin.H{"status": "login"})
	}
}

func getImageBucketToken(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json;charset=utf-8")
	token := sessions.Default(c).Get("token")
	if token != nil {
		jsonInfo, _ := json.Marshal(token)
		var jsonInfoObj []string
		json.Unmarshal(jsonInfo, &jsonInfoObj)

		var nowuserinfo []UserInfo
		err := Db.Select(&nowuserinfo, "SELECT uid, username, password, email, avatar, uhash, time FROM user WHERE `uid`=?", jsonInfoObj[0])
		if err != nil {
			c.JSON(200, gin.H{"status": "error"})
		}

		jsonByte, _ := json.Marshal(map[string]string{"u": fmt.Sprint(nowuserinfo[0].Uid), "p": nowuserinfo[0].Password})
		aesKey := time.Now().Format("06-01") + "Cpeuf3BPjfQudgxwkU*Yuxy.h6!"
		tokenBase64 := aesEncrypt([]byte(aesKey), string(jsonByte))

		c.JSON(200, gin.H{"status": "ok", "token": tokenBase64})
	} else {
		c.JSON(200, gin.H{"status": "login"})
	}
}

func badge(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json;charset=utf-8")
	token := sessions.Default(c).Get("token")
	if token != nil {
		jsonInfo, _ := json.Marshal(token)
		var jsonInfoObj []string
		json.Unmarshal(jsonInfo, &jsonInfoObj)

		c.JSON(200, gin.H{"status": "ok", "data": getBadge(jsonInfoObj[0])})
	} else {
		c.JSON(200, gin.H{"status": "login"})
	}
}

func book(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json;charset=utf-8")
	session := sessions.Default(c)
	info := session.Get("token")
	if info != nil {
		bid := c.PostForm("bid")
		uid := info.([]string)[0]
		bidint, _ := strconv.Atoi(bid)

		// Bookinfo
		fbool := intInSlice(bidint, favoritesSelectSql(uid))
		bookinfo := bidGetBook(bid)

		var bookinfothis Booktype

		if bid != "" && len(bookinfo) != 0 {
			bookinfothis = bookinfo[0]
			bookinfothis.F = fbool
		} else {
			bookinfothis = bidGetBook("1")[0]
			bookinfothis.F = false
		}

		// recommend
		if bid == "" {
			bid = "0"
		}
		similarList := getSimilar(bid)
		similarJson := make([]map[string]string, len(similarList))
		for i := 0; i < len(similarList); i++ {
			similarJson[i] = map[string]string{"bid": fmt.Sprint(similarList[i].Bid), "bookname": similarList[i].Bookname}
		}

		// getComment
		commInfo := getCommentDb(bidint, 0)
		commLen := getCommentLen(bidint)

		over := false

		// 評論置頂
		if commtop.Show {
			commInfo = append([]CommInfoUser{{
				Cid:       1,
				Uid:       0,
				Username:  commtop.Username,
				AvatarUrl: "",
				EmailMd5:  commtop.EmailMd5,
				Area:      commtop.Type,
				Reply:     -1,
				ReplyText: "",
				Platform:  "web",
				Time:      commtop.Time,
				Content:   commtop.Content,
			}}, commInfo...)
			commLen++
		}

		if 40 > commLen {
			over = true
		}

		var comment gin.H

		if commLen != 0 {
			comment = gin.H{"info": commInfo, "len": commLen, "over": over}
		} else {
			comment = gin.H{"info": []string{}, "len": commLen, "over": true}
		}

		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"bookinfo":  bookinfothis,
			"recommend": similarJson,
			"comment":   comment,
		})
	} else {
		c.JSON(200, gin.H{"status": "login"})
	}
}

func communityList(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json;charset=utf-8")
	session := sessions.Default(c)
	info := session.Get("token")
	if info != nil {
		page := c.PostForm("page")
		pageInt, err := strconv.Atoi(page)
		if err != nil {
			pageInt = 1
		}
		pageTo := (pageInt - 1) * 40
		communityInfo := getCommunityDb(pageTo)
		communityLen := getCommunityLen()

		over := false

		// 置頂
		if pageInt == 1 && communitytop.Show && communitytop.Expired > int(time.Now().Unix()) {
			communityInfo = append([]Community{{
				Id:       communitytop.Id,
				Username: communitytop.Username,
				Title:    communitytop.Title,
				Content:  communitytop.Content,
				Image:    communitytop.Image,
				Time:     communitytop.Time,
				Expired:  communitytop.Expired,
				Area:     communitytop.Type,
				Avatar:   communitytop.Avatar,
				Platform: "web",
			}}, communityInfo...)
			communityLen++
		}

		if pageInt*40 > communityLen {
			over = true
		}

		if communityLen != 0 {
			c.JSON(200, gin.H{"status": "ok", "info": communityInfo, "len": communityLen, "over": over})
		} else {
			c.JSON(200, gin.H{"status": "ok", "info": []string{}, "len": communityLen, "over": true})
		}
	} else {
		c.JSON(200, gin.H{"status": "login"})
	}
}

func sendCommunity(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json;charset=utf-8")
	session := sessions.Default(c)
	info := session.Get("token")
	if info != nil {
		uid, _ := strconv.Atoi(info.([]string)[0])
		title := c.PostForm("title")
		content := c.PostForm("content")
		image := c.PostForm("image")
		var imageList []string = []string{}
		json.Unmarshal([]byte(image), &imageList)
		ip := c.ClientIP()
		area := getIpArea(ip)
		emailMd5 := GetMD5HashCode([]byte(getDbUserInfo(info.([]string)[0]).Email))
		platform := c.PostForm("platform")

		if title != "" && content != "" && platform != "" {
			uidInt, _ := strconv.Atoi(info.([]string)[0])
			if useISql(uidInt, "0.5")["status"] == "ok" {
				row, err := Db.Query("INSERT INTO `community` (`uid`, `title`, `content`, `image`, `time`, `expired`, `ip`, `area`, `email_md5`, `platform`) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", uid, title, content, imageList, time.Now().Unix(), time.Now().Unix()+60*60*24*7, ip, area, emailMd5, platform)
				if err != nil {
					log.SetFlags(log.Lshortfile | log.LstdFlags)
					log.Println("SqlError", err)
				}
				defer row.Close()
				c.JSON(200, gin.H{"status": "ok"})
			} else {
				c.JSON(200, gin.H{"status": "too_few"})
			}
		} else {
			c.JSON(200, gin.H{"status": "empty"})
		}
	} else {
		c.JSON(200, gin.H{"status": "login"})
	}
}

func sendCommunityReply(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json;charset=utf-8")
	session := sessions.Default(c)
	info := session.Get("token")
	if info != nil {
		uid, _ := strconv.Atoi(info.([]string)[0])
		id, err := strconv.Atoi(c.PostForm("id"))
		if err != nil {
			c.JSON(200, gin.H{"status": "id_err"})
			return
		}
		reply, err := strconv.Atoi(c.PostForm("reply"))
		if err != nil {
			reply = -1
		}
		content := c.PostForm("text")
		ip := c.ClientIP()
		area := getIpArea(ip)
		platform := c.PostForm("platform")
		if content != "" && platform != "" {
			row, err := Db.Query("INSERT INTO `community_reply` (`cid`, `uid`, `content`, `reply`, `ip`, `area`, `time`, `platform`) VALUES (?, ?, ?, ?, ?, ?, ?, ?)", id, uid, content, reply, ip, area, time.Now().Unix(), platform)
			if err != nil {
				log.SetFlags(log.Lshortfile | log.LstdFlags)
				log.Println("SqlError", err)
			}
			defer row.Close()
			c.JSON(200, gin.H{"status": "ok"})
		} else {
			c.JSON(200, gin.H{"status": "empty"})
		}
	} else {
		c.JSON(200, gin.H{"status": "login"})
	}
}

func getCommunityPost(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json;charset=utf-8")
	session := sessions.Default(c)
	info := session.Get("token")
	if info != nil {
		communityId, err := strconv.Atoi(c.PostForm("id"))
		if err != nil {
			c.JSON(200, gin.H{"status": "id_err"})
			return
		}
		var communityInfo []CommunityDb
		err = Db.Select(&communityInfo, "SELECT `id`, `uid`, `title`, `content`, `image`, `time`, `expired`, `area`, `email_md5`, `platform` FROM community WHERE `id` = ?", communityId)
		if err != nil {
			c.JSON(200, gin.H{"status": "get_err"})
			log.SetFlags(log.Lshortfile | log.LstdFlags)
			log.Println("SqlError", err)
			return
		}

		var communityReplyList []CommunityReplyDb = []CommunityReplyDb{}
		err = Db.Select(&communityReplyList, "SELECT `rid`, `uid`, `content`, `reply`, `area`, `time`, `platform` FROM community_reply WHERE `cid` = ?", communityId)
		if err != nil {
			log.SetFlags(log.Lshortfile | log.LstdFlags)
			log.Println("SqlError", err)
		}

		var uidList []string = []string{}
		uidList = append(uidList, fmt.Sprint(communityInfo[0].Uid))
		for _, j := range communityReplyList {
			uidList = append(uidList, fmt.Sprint(j.Uid))
		}

		var userinfo []UserInfoS
		err = Db.Select(&userinfo, "SELECT `uid`, `username`, `email`, `avatar` FROM user WHERE `uid` in ("+strings.Join(uidList, ",")+")")
		if err != nil {
			log.SetFlags(log.Lshortfile | log.LstdFlags)
			log.Println("SqlError", err)
		}

		var userMap map[int]map[string]string = make(map[int]map[string]string)
		for _, j := range userinfo {
			var avatar string = ""
			if j.Avatar != "default.webp" {
				avatar = j.Avatar
			} else {
				avatar = GetMD5HashCode([]byte(j.Email))
			}
			userMap[j.Uid] = map[string]string{
				"username": j.Username,
				"avatar":   avatar,
			}
		}

		var replyIdList []string = []string{}
		for _, j := range communityReplyList {
			replyIdList = append(replyIdList, fmt.Sprint(j.ReplyId))
		}

		var replyContent []CommunityReplyDb = []CommunityReplyDb{}
		if len(replyIdList) != 0 {
			err = Db.Select(&replyContent, "SELECT `rid`, `uid`, `content` FROM community_reply WHERE `rid` in ("+strings.Join(replyIdList, ",")+")")
			if err != nil {
				log.SetFlags(log.Lshortfile | log.LstdFlags)
				log.Println("SqlError", err)
			}
		}

		var replyContentMap map[int]string = make(map[int]string)
		for _, j := range replyContent {
			replyContentMap[j.Rid] = userMap[j.Uid]["username"] + ": " + j.Content
		}

		var replyList []CommunityReply = []CommunityReply{}
		for i := range communityReplyList {
			replyList = append(replyList, CommunityReply{
				Rid:          communityReplyList[i].Rid,
				Username:     userMap[communityReplyList[i].Uid]["username"],
				Avatar:       userMap[communityInfo[0].Uid]["avatar"],
				Content:      communityReplyList[i].Content,
				ReplyContent: replyContentMap[communityReplyList[i].ReplyId],
				ReplyId:      communityReplyList[i].ReplyId,
				Area:         communityReplyList[i].Area,
				Time:         communityReplyList[i].Time,
				Platform:     communityReplyList[i].Platform,
			})
		}

		var imageList []string = []string{}
		json.Unmarshal(communityInfo[0].Image, &imageList)

		c.JSON(200, gin.H{
			"status": "ok",
			"info": Community{
				Id:       communityInfo[0].Id,
				Username: userMap[communityInfo[0].Uid]["username"],
				Title:    communityInfo[0].Title,
				Content:  communityInfo[0].Content,
				Image:    imageList,
				Time:     communityInfo[0].Time,
				Expired:  communityInfo[0].Expired,
				Area:     communityInfo[0].Area,
				Avatar:   userMap[communityInfo[0].Uid]["avatar"],
				Platform: communityInfo[0].Platform,
			},
			"reply": replyList,
		})
	} else {
		c.JSON(200, gin.H{"status": "login"})
	}
}
