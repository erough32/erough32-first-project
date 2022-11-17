package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/liuzl/gocc"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserInfo struct {
	Uid      int    `db:"uid"`
	Username string `db:"username"`
	Password string `db:"password"`
	Email    string `db:"email"`
	Avatar   string `db:"avatar"`
	Uhash    string `db:"uhash"`
	Time     int    `db:"time"`
}

type UserInfoS struct {
	Uid      int     `db:"uid"`
	Username string  `db:"username"`
	Email    string  `db:"email"`
	Uhash    string  `db:"uhash"`
	Time     int     `db:"time"`
	Avatar   string  `db:"avatar"`
	Integral float64 `db:"integral"`
	Badge    []uint8 `db:"badge"`
}

type userAvatar struct {
	Uid      int    `db:"uid"`
	Username string `db:"username"`
	Avatar   string `db:"avatar"`
}

type Bid struct {
	Bid int `db:"bid"`
}

type Bnum struct {
	Num int `db:"count(bid)"`
}

type Booktype struct {
	Bid       int    `db:"bid"`
	Bookname  string `db:"bookname"`
	Author    string `db:"author"`
	Pname     string `db:"pname"`
	Ptag      string `db:"ptag"`
	Otag      string `db:"otag"`
	Time      int    `db:"time"`
	Views     int    `db:"views"`
	Favorites int    `db:"favorites"`
	F         bool
	Len       int `db:"len"`
}

type Uid struct {
	Uid int `db:"uid"`
}

type favoritestype struct {
	Bookid []uint8 `db:"bookid"`
}

type CommInfo struct {
	Cid      int    `db:"cid"`
	Bid      int    `db:"bid"`
	Uid      int    `db:"uid"`
	Username string `db:"username"`
	Email    string `db:"email"`
	EmailMd5 string `db:"email_md5"`
	Ip       string `db:"ip"`
	Area     string `db:"area"`
	Reply    int    `db:"reply"`
	Platform string `db:"platform"`
	Time     int    `db:"time"`
	Content  string `db:"content"`
}

type CommInfoUser struct {
	Cid       int    `db:"cid" json:"cid"`
	Uid       int    `db:"uid" json:"uid"`
	Username  string `db:"username" json:"username"`
	EmailMd5  string `db:"email_md5" json:"avatar"`
	AvatarUrl string `json:"avatarUrl"`
	Area      string `db:"area" json:"area"`
	Reply     int    `db:"reply" json:"reply"`
	ReplyText string `json:"reply_text"`
	Platform  string `db:"platform" json:"platform"`
	Time      int    `db:"time" json:"time"`
	Content   string `db:"content" json:"content"`
}

type CommInfoReply struct {
	Cid       int    `db:"cid" json:"cid"`
	Uid       int    `db:"uid" json:"uid"`
	Bid       int    `db:"bid" json:"bid"`
	Username  string `db:"username" json:"username"`
	EmailMd5  string `db:"email_md5" json:"avatar"`
	AvatarUrl string `json:"avatarUrl"`
	Area      string `db:"area" json:"area"`
	Reply     int    `db:"reply" json:"reply"`
	ReplyText string `json:"reply_text"`
	Platform  string `db:"platform" json:"platform"`
	Time      int    `db:"time" json:"time"`
	Content   string `db:"content" json:"content"`
}

type ConfigDb struct {
	Cid    int     `db:"cid"`
	Key    string  `db:"key"`
	Config []uint8 `db:"config"`
}

type AdsDb struct {
	Img string `db:"img" json:"img"`
	Url string `db:"url" json:"url"`
}

type readLeaderboard struct {
	Bid   int `db:"bid"`
	Day   int `db:"day"`
	Week  int `db:"week"`
	Month int `db:"moon"`
}

type readLeaderboardNum struct {
	Num int `db:"count(bid)"`
}

type Commtop struct {
	Username string `json:"Username"`
	Type     string `json:"type"`
	EmailMd5 string `json:"EmailMd5"`
	Time     int    `json:"Time"`
	Content  string `json:"Content"`
	Show     bool   `json:"Show"`
}

type Badge struct {
	Name   string `db:"name"`
	Img    string `db:"img"`
	Minimg string `db:"minimg"`
	Des    string `db:"des"`
}

type userinfoMap struct {
	Email      string
	Integral   float64
	Time       int
	Uhash      string
	Uid        int
	Username   string
	Avatar     string
	CommentLen int
	Badge      []Badge
}

type Sysmsg struct {
	Mid      int    `db:"mid"`
	Touid    int    `db:"touid"`
	Username string `db:"username"`
	Avatar   string `db:"avatar"`
	Msg      string `db:"msg"`
	Time     int    `db:"time"`
}

type Community struct {
	Id       int      `json:"id"`
	Username string   `json:"username"`
	Title    string   `json:"title"`
	Content  string   `json:"content"`
	Image    []string `json:"image"`
	Time     int      `json:"time"`
	Expired  int      `json:"expired"`
	Area     string   `json:"area"`
	Avatar   string   `json:"avatar"`
	Platform string   `json:"platform"`
}

type CommunityDb struct {
	Id       int     `db:"id"`
	Uid      int     `db:"uid"`
	Title    string  `db:"title"`
	Content  string  `db:"content"`
	Image    []uint8 `db:"image"`
	Time     int     `db:"time"`
	Expired  int     `db:"expired"`
	Area     string  `db:"area"`
	EmailMd5 string  `db:"email_md5"`
	Platform string  `db:"platform"`
}

type CommunityReplyDb struct {
	Rid      int    `db:"rid"`
	Uid      int    `db:"uid"`
	Content  string `db:"content"`
	ReplyId  int    `db:"reply"`
	Area     string `db:"area"`
	Time     int    `db:"time"`
	Platform string `db:"platform"`
}

type CommunityReply struct {
	Rid          int    `json:"rid"`
	Username     string `json:"username"`
	Avatar       string `json:"avatar"`
	Content      string `json:"content"`
	ReplyId      int    `json:"reply_id"`
	ReplyContent string `json:"reply"`
	Area         string `json:"area"`
	Time         int    `json:"time"`
	Platform     string `json:"platform"`
}

type Communitytop struct {
	Id       int
	Username string
	Type     string
	Avatar   string
	Time     int
	Expired  int
	Title    string
	Content  string
	Image    []string
	Show     bool
}

var Db *sqlx.DB
var AdsJson []AdsDb     // 廣告列表
var configTags []string // 首頁 Tag 推薦
var downloadCredit int  // 下載積分定價
var topbook int         // 置頂本（-1 不顯示）
var updateInfo map[string]string
var bulletinInfo struct {
	Text string `json:"text"`
	Show bool   `json:"show"`
}
var commtop Commtop           // 置頂評論
var communitytop Communitytop // 置頂社區

func init() {
	database, err := sqlx.Open("mysql", sqlConfig("use"))
	if err != nil {
		log.SetFlags(log.Lshortfile | log.LstdFlags)
		log.Println("open mysql failed,", err)
		return
	}
	database.SetMaxOpenConns(3000)
	database.SetMaxIdleConns(500)

	Db = database

	// Ads
	err = Db.Select(&AdsJson, "SELECT `img`, `url` FROM ads ORDER BY `order`")
	if err != nil {
		log.SetFlags(log.Lshortfile | log.LstdFlags)
		log.Println("exec failed, ", err)
	}

	// Tag
	ReadTags = getReadTags()

	// ConfigTag
	loadConfig()
}

func loadConfig() {
	var configInfo []ConfigDb
	err := Db.Select(&configInfo, "SELECT * FROM `config`")
	if err != nil {
		log.SetFlags(log.Lshortfile | log.LstdFlags)
		log.Println("SqlError", err)
	}

	for i := range configInfo {
		switch configInfo[i].Key {
		case "tag":
			json.Unmarshal(configInfo[i].Config, &configTags)
		case "downloads":
			var configdc map[string]int
			json.Unmarshal(configInfo[i].Config, &configdc)
			downloadCredit = configdc["oi"]
		case "topbook":
			var configtb map[string]int
			json.Unmarshal(configInfo[i].Config, &configtb)
			topbook = configtb["bid"]
		case "bulletin":
			json.Unmarshal(configInfo[i].Config, &bulletinInfo)
		case "commtop":
			json.Unmarshal(configInfo[i].Config, &commtop)
		case "update":
			json.Unmarshal(configInfo[i].Config, &updateInfo)
		case "communitytop":
			var communityConfig struct {
				Show bool `json:"Show"`
				Id   int  `json:"Id"`
			}
			json.Unmarshal(configInfo[i].Config, &communityConfig)
			if communityConfig.Show {
				var communityData []CommunityDb
				err := Db.Select(&communityData, "SELECT `id`, `uid`, `title`, `content`, `image`, `time`, `expired`, `area`, `email_md5`, `platform` FROM community WHERE id = ?", communityConfig.Id)
				if err != nil {
					log.SetFlags(log.Lshortfile | log.LstdFlags)
					log.Println("SqlError", err)
				} else {
					var userinfo []UserInfoS
					err = Db.Select(&userinfo, "SELECT `username`, `email`, `avatar` FROM `uid` = ?", communityData[0].Uid)
					if err != nil {
						log.SetFlags(log.Lshortfile | log.LstdFlags)
						log.Println("SqlError", err)
					}
					var avatar string
					if userinfo[0].Avatar != "default.webp" {
						avatar = userinfo[0].Avatar
					} else {
						avatar = GetMD5HashCode([]byte(userinfo[0].Email))
					}

					var imagelist []string = []string{}
					json.Unmarshal(communityData[0].Image, &imagelist)

					communitytop = Communitytop{
						Show:     true,
						Id:       communityData[0].Id,
						Username: userinfo[0].Username,
						Type:     communityData[0].Area,
						Avatar:   avatar,
						Time:     communityData[0].Time,
						Expired:  communityData[0].Expired,
						Title:    communityData[0].Title,
						Image:    imagelist,
						Content:  communityData[0].Content,
					}
				}
			} else {
				communitytop = Communitytop{Show: false}
			}
		}
	}
}

func loginSelect(user string, pass string) int {
	defer func() {
		err := recover()
		if err != nil {
			return // 未查詢到訊息時拋出錯誤
		}
	}()

	if SQLInject(user) || SQLInject(pass) {
		return -1
	}

	var userInfo []UserInfo
	err := Db.Select(&userInfo, "SELECT uid,username,password,email,avatar FROM user WHERE (`username`=? OR `email` = ?) AND `password`=?", user, user, pass)
	if err != nil {
		log.SetFlags(log.Lshortfile | log.LstdFlags)
		log.Println("exec failed, ", err)
		return 0
	}

	return userInfo[0].Uid
}

func selectYQ(uhash string) int {
	defer func() {
		err := recover()
		if err != nil {
			return // 拋出錯誤
		}
	}()
	var uid []Uid
	err := Db.Select(&uid, "SELECT uid FROM user WHERE `uhash` = ?", uhash)
	if err != nil {
		log.SetFlags(log.Lshortfile | log.LstdFlags)
		log.Println(err)
	}
	return uid[0].Uid
}

func regSql(user string, pass string, email string, fuid int) {
	var uidInfo []Uid
	Db.Select(&uidInfo, "select uid from user order by uid desc limit 1")
	Db.Exec("INSERT INTO user (`uid`, `username`, `password`, `email`, `avatar`, `time`, `uhash`, `fuid`) VALUES (?, ?, ?, ?, 'default.webp', ?, ?, ?)", uidInfo[0].Uid+1, user, pass, email, time.Now().Unix(), GetSHA256HashCode([]byte(fmt.Sprint(uidInfo[0].Uid + 1)))[:9], fuid)
	row, _ := Db.Query("update user set integral=integral+2 where uid = ?", fuid)
	defer func() {
		row.Close()
		err := recover()
		if err != nil {
			return // 拋出錯誤
		}
	}()
}

func searchUser(user string) int {
	defer func() {
		err := recover()
		if err != nil {
			return // 未查詢到訊息時拋出錯誤
		}
	}()

	var userInfo []Uid
	err := Db.Select(&userInfo, "SELECT uid FROM user WHERE `username`=?", user)
	if err != nil {
		log.SetFlags(log.Lshortfile | log.LstdFlags)
		log.Println("exec failed, ", err)
		return 0
	}

	return userInfo[0].Uid
}

func getUserInfo(uid int) []string {
	defer func() {
		err := recover()
		if err != nil {
			return // 未查詢到訊息時拋出錯誤
		}
	}()

	var userInfo []UserInfo
	err := Db.Select(&userInfo, "SELECT uid,username,password,email,avatar,uhash,time FROM user WHERE `uid` = ?", uid)
	if err != nil {
		log.SetFlags(log.Lshortfile | log.LstdFlags)
		log.Println("exec failed, ", err)
		return nil
	}

	return []string{fmt.Sprint(userInfo[0].Uid), userInfo[0].Username, userInfo[0].Email, userInfo[0].Avatar, userInfo[0].Uhash, fmt.Sprint(userInfo[0].Time)}
}

func bookList(pageinfo int) string {
	defer func() {
		err := recover()
		if err != nil {
			return // 拋出錯誤
		}
	}()

	pagenum := (pageinfo - 1) * 20

	var datainfo []Booktype
	err := Db.Select(&datainfo, "select * from book order by bid desc limit ?,20", pagenum)
	if err != nil {
		log.SetFlags(log.Lshortfile | log.LstdFlags)
		log.Println(err)
	}
	liststr := "["
	for _, i := range datainfo {
		jsoninfo, _ := json.Marshal(i)
		liststr = liststr + string(jsoninfo) + ","
	}
	liststr = liststr[:len(liststr)-1] + "]"
	return liststr
}

func booknum() int {
	defer func() {
		err := recover()
		if err != nil {
			return // 拋出錯誤
		}
	}()

	var bnum []Bnum
	err := Db.Select(&bnum, "select count(bid) from book")
	if err != nil {
		log.SetFlags(log.Lshortfile | log.LstdFlags)
		log.Println(err)
	}

	return bnum[0].Num
}

func bookListSqlV2(pageinfo int) ([]Booktype, int) {
	defer func() {
		err := recover()
		if err != nil {
			return // 拋出錯誤
		}
	}()

	pagenum := (pageinfo - 1) * 20

	var datainfo []Booktype
	err := Db.Select(&datainfo, "select * from book order by bid desc limit ?,20", pagenum)
	if err != nil {
		log.SetFlags(log.Lshortfile | log.LstdFlags)
		log.Println(err)
	}

	var bnum []Bnum
	err = Db.Select(&bnum, "select count(bid) from book")
	if err != nil {
		log.SetFlags(log.Lshortfile | log.LstdFlags)
		log.Println(err)
	}

	return datainfo, bnum[0].Num
}

func searchBookNew(bookname string, pageinfo int, searchType string, sortType string, btags string) map[string]interface{} {
	defer func() {
		err := recover()
		if err != nil {
			return
		}
	}()

	pagenum := (pageinfo - 1) * 20

	var bookinfo []Booktype
	var bnum []Bnum

	// Sort
	if sortType == "" || sortType == "bid" {
		sortType = " order by bid desc"
	} else if sortType == "views" {
		sortType = " order by views desc"
	} else if sortType == "favorites" {
		sortType = " order by favorites desc"
	} else {
		sortType = ""
	}

	// block tag
	var btagSQl string = ""
	if btags != "" {
		s2t, err := gocc.New("s2t")
		if err != nil {
			log.Fatal(err)
		}
		out, err := s2t.Convert(btags)
		if err != nil {
			log.Fatal(err)
		}
		var blocktags []string
		json.Unmarshal([]byte(out), &blocktags)
		btagSQl = " AND ptag NOT REGEXP '" + strings.Join(blocktags, "|") + "'"
	}

	if searchType == "" || searchType == "de" {
		var data struct {
			Name string `db:"name"`
		}
		mongoCollectionTags.FindOne(ctx, bson.D{{Key: "name", Value: bookname}}).Decode(&data)
		if data.Name != "" {
			searchType = "tag"
		}
	}

	if searchType == "" || searchType == "de" {
		// Default
		searchList := searchAny(bookname)
		searchListLen := len(searchList)
		if searchListLen == 0 {
			bookinfo = []Booktype{}
		} else {
			search20 := searchList[pagenum:operator3(pagenum+20 >= searchListLen, searchListLen, pagenum+20)]

			if sortType == " order by bid desc" {
				sortType = " order by field(bid, " + strings.Join(search20, ",") + ")"
			}

			err := Db.Select(&bookinfo, "SELECT * FROM book WHERE `bid` in ("+strings.Join(search20, ",")+")"+btagSQl+" "+sortType)
			if err != nil {
				log.SetFlags(log.Lshortfile | log.LstdFlags)
				log.Println(err)
			}
		}
		bnum = []Bnum{{Num: searchListLen}}
	} else if searchType == "tag" {
		// Tag
		s2t, err := gocc.New("s2t")
		if err != nil {
			log.Fatal(err)
		}
		out, err := s2t.Convert(bookname)
		if err != nil {
			log.Fatal(err)
		}
		taglist := strings.Split(strings.TrimSpace(out), " ")
		tagStr := strings.Join(taglist, " +")

		// Search
		err = Db.Select(&bookinfo, "SELECT * FROM book WHERE MATCH(`pname`,`ptag`,`otag`) AGAINST (? IN BOOLEAN MODE)"+btagSQl+" "+sortType+" limit ?,20", "+"+tagStr, pagenum)
		if err != nil {
			log.SetFlags(log.Lshortfile | log.LstdFlags)
			log.Println(err)
		}
		err = Db.Select(&bnum, "SELECT count(bid) FROM book WHERE MATCH(`pname`,`ptag`,`otag`) AGAINST (? IN BOOLEAN MODE)"+btagSQl, "+"+tagStr)
		if err != nil {
			log.SetFlags(log.Lshortfile | log.LstdFlags)
			log.Println(err)
		}
	} else if searchType == "author" {
		// Author
		// Search
		err := Db.Select(&bookinfo, "SELECT * FROM book WHERE `author` LIKE ?"+btagSQl+" "+sortType+" limit ?,20", "%"+bookname+"%", pagenum)
		if err != nil {
			log.SetFlags(log.Lshortfile | log.LstdFlags)
			log.Println(err)
		}
		err = Db.Select(&bnum, "SELECT count(bid) FROM book WHERE `author` LIKE ?"+btagSQl, "%"+bookname+"%")
		if err != nil {
			log.SetFlags(log.Lshortfile | log.LstdFlags)
			log.Println(err)
		}
	}

	return map[string]interface{}{"Info": bookinfo, "len": bnum[0].Num}
}

func bidGetBook(bid string) []Booktype {
	defer func() {
		err := recover()
		if err != nil {
			return
		}
	}()

	var bookinfo []Booktype

	bidInt, _ := strconv.Atoi(bid)
	err := Db.Select(&bookinfo, "SELECT * FROM book WHERE bid = ?", bidInt)
	if err != nil {
		log.SetFlags(log.Lshortfile | log.LstdFlags)
		log.Println(err)
	}
	return bookinfo
}

func getDbUserInfo(uid string) userinfoMap {
	defer func() {
		err := recover()
		if err != nil {
			return
		}
	}()

	var userinfo []UserInfoS
	err := Db.Select(&userinfo, "SELECT uid, username, email, avatar, time, uhash, integral, badge FROM user WHERE `uid`=?", uid)
	if err != nil {
		log.SetFlags(log.Lshortfile | log.LstdFlags)
		log.Println("SQL Error:", err)
	}

	type commentLen struct {
		Len int `db:"len"`
	}

	var commentInfo []commentLen
	err = Db.Select(&commentInfo, "SELECT COUNT(cid) as len FROM comment WHERE `reply` in (SELECT cid FROM comment WHERE `uid` = ?)", uid)
	if err != nil {
		log.SetFlags(log.Lshortfile | log.LstdFlags)
		log.Println("SQL Error:", err)
	}

	type sysmsgNum struct {
		Len int `db:"len"`
	}

	var sysmsgLen []sysmsgNum = []sysmsgNum{}
	err = Db.Select(&sysmsgLen, "SELECT count(mid) as len FROM sysmsg WHERE `touid`=?", uid)
	if err != nil {
		log.SetFlags(log.Lshortfile | log.LstdFlags)
		log.Println(err)
	}

	var data struct {
		Bid    int `db:"bid"`
		Reply  int `db:"reply"`
		Sysmsg int `db:"sysmsg"`
	}
	mongoCollectionMsg.FindOne(ctx, bson.D{{Key: "uid", Value: userinfo[0].Uid}}).Decode(&data)

	userAvatar := userinfo[0].Avatar
	if userAvatar == "default.webp" {
		userAvatar = ""
	}

	badgeId := []string{}
	json.Unmarshal(userinfo[0].Badge, &badgeId)
	var badgeInfo []Badge = []Badge{}
	Db.Select(&badgeInfo, "SELECT name, img, minimg, des FROM badge WHERE id in ("+strings.Join(badgeId, ",")+")")

	return userinfoMap{
		Email:      userinfo[0].Email,
		Integral:   userinfo[0].Integral,
		Time:       userinfo[0].Time,
		Uhash:      userinfo[0].Uhash,
		Uid:        userinfo[0].Uid,
		Username:   userinfo[0].Username,
		Avatar:     userAvatar,
		CommentLen: commentInfo[0].Len + sysmsgLen[0].Len - data.Reply - data.Sysmsg,
		Badge:      badgeInfo,
	}
}

func getCommentReply(uid string, page, num int) gin.H {
	var commentInfo []CommInfoReply
	err := Db.Select(&commentInfo, "SELECT cid, bid, uid, username, email_md5, area, platform, reply, time, content FROM comment WHERE `reply` in (SELECT cid FROM comment WHERE `uid` = ?) ORDER BY `cid` DESC LIMIT ?,40", uid, ((page - 1) * 40))
	if err != nil {
		log.SetFlags(log.Lshortfile | log.LstdFlags)
		log.Println("SQL Error:", err)
	}

	var cidlist []string
	for _, j := range commentInfo {
		cidlist = append(cidlist, fmt.Sprint(j.Reply))
	}

	var commmap map[int]string = make(map[int]string)

	var commentInfoR []CommInfoUser
	err = Db.Select(&commentInfoR, "SELECT cid, content FROM comment WHERE `cid` in ("+strings.Join(cidlist, ",")+")")
	if err != nil {
		log.SetFlags(log.Lshortfile | log.LstdFlags)
		log.Println("SQL Error:", err)
	}
	for _, j := range commentInfoR {
		commmap[j.Cid] = j.Content
	}
	for i := range commentInfo {
		commentInfo[i].ReplyText = commmap[commentInfo[i].Reply]
	}

	var uidList []string
	for _, j := range commentInfo {
		uidList = append(uidList, fmt.Sprint(j.Uid))
	}
	var userinfo []userAvatar = []userAvatar{}
	if len(uidList) != 0 {
		err = Db.Select(&userinfo, "SELECT uid, username, avatar FROM user WHERE `uid` in ("+strings.Join(uidList, ",")+") AND `avatar` <> 'default.webp'")
		if err != nil {
			log.SetFlags(log.Lshortfile | log.LstdFlags)
			log.Println("SqlError", err)
		}
	}
	var userAvatarMap map[int]string = make(map[int]string)
	for _, j := range userinfo {
		userAvatarMap[j.Uid] = j.Avatar
	}
	for i := range commentInfo {
		commentInfo[i].AvatarUrl = userAvatarMap[commentInfo[i].Uid]
	}

	type commentLen struct {
		Len int `db:"len"`
	}

	var commentNum []commentLen
	err = Db.Select(&commentNum, "SELECT COUNT(cid) as len FROM comment WHERE `reply` in (SELECT cid FROM comment WHERE `uid` = ?)", uid)
	if err != nil {
		log.SetFlags(log.Lshortfile | log.LstdFlags)
		log.Println("SQL Error:", err)
	}

	var bookmap map[int]Booktype = make(map[int]Booktype)

	if len(commentInfo) != 0 {
		var bidlist []string
		for _, j := range commentInfo {
			bidlist = append(bidlist, fmt.Sprint(j.Bid))
		}

		var bookinfo []Booktype
		err := Db.Select(&bookinfo, "SELECT * FROM book WHERE `bid` in ("+strings.Join(bidlist, ",")+")")
		if err != nil {
			log.SetFlags(log.Lshortfile | log.LstdFlags)
			log.Println(err)
		}
		for _, j := range bookinfo {
			bookmap[j.Bid] = j
		}
	}

	uidInt, _ := strconv.Atoi(uid)
	mongoCollectionMsg.UpdateOne(ctx, bson.D{{Key: "uid", Value: uidInt}}, bson.D{{Key: "$set", Value: bson.D{{Key: "reply", Value: commentNum[0].Len}}}}, options.Update().SetUpsert(true))

	return gin.H{
		"comment":    commentInfo,
		"book":       bookmap,
		"commentLen": commentNum[0].Len - num,
	}
}

func getSysMsg(uid string, page, num int) gin.H {
	var sysmsg []Sysmsg
	err := Db.Select(&sysmsg, "SELECT * FROM sysmsg WHERE `touid`=? ORDER BY `mid` DESC LIMIT ?,40", uid, (page-1)*40)
	if err != nil {
		log.SetFlags(log.Lshortfile | log.LstdFlags)
		log.Println(err)
	}

	type sysmsgNum struct {
		Len int `db:"len"`
	}

	var sysmsgLen []sysmsgNum = []sysmsgNum{}
	err = Db.Select(&sysmsgLen, "SELECT count(mid) as len FROM sysmsg WHERE `touid`=?", uid)
	if err != nil {
		log.SetFlags(log.Lshortfile | log.LstdFlags)
		log.Println(err)
	}

	uidInt, _ := strconv.Atoi(uid)
	mongoCollectionMsg.UpdateOne(ctx, bson.D{{Key: "uid", Value: uidInt}}, bson.D{{Key: "$set", Value: bson.D{{Key: "sysmsg", Value: sysmsgLen[0].Len}}}}, options.Update().SetUpsert(true))

	return gin.H{
		"msg": sysmsg,
		"len": sysmsgLen[0].Len - num,
	}
}

func useISql(uid int, num string) map[string]string {
	var row *sql.Rows

	defer func() {
		row.Close()
		err := recover()
		if err != nil {
			return
		}
	}()

	var userinfo []UserInfoS
	err := Db.Select(&userinfo, "SELECT integral FROM user WHERE `uid`=?", uid)
	if err != nil {
		log.SetFlags(log.Lshortfile | log.LstdFlags)
		log.Println("SQL Error:", err)
	}

	if len(userinfo) != 0 {
		nowI := userinfo[0].Integral
		useINum, _ := strconv.ParseFloat(num, 64)
		usedI := nowI - useINum
		if usedI >= 0 {
			row, _ = Db.Query("UPDATE user SET `integral` = ? WHERE `uid` = ?", usedI, uid)
			return map[string]string{"status": "ok"}
		} else {
			return map[string]string{"status": "too_few"}
		}
	} else {
		return map[string]string{"status": "not_found"}
	}
}

func getBookListByList(bookid string) []Booktype {
	defer func() {
		err := recover()
		if err != nil {
			return
		}
	}()

	var bookinfo []Booktype = []Booktype{}

	if bookid != "" {
		err := Db.Select(&bookinfo, "SELECT * FROM book WHERE `bid` in ("+bookid+") order by field(bid, "+bookid+")")
		if err != nil {
			log.SetFlags(log.Lshortfile | log.LstdFlags)
			log.Println(err)
		}
	}

	return bookinfo
}

func getSimilar(bid string) []Booktype {
	defer func() {
		err := recover()
		if err != nil {
			return
		}
	}()

	var data struct {
		Bid  int   `db:"bid"`
		Data []int `db:"data"`
	}
	bidint, _ := strconv.Atoi(bid)
	mongoCollectionBook.FindOne(ctx, bson.D{{Key: "bid", Value: bidint}}).Decode(&data)

	if len(data.Data) != 0 {
		var sliststr []string

		for _, j := range data.Data {
			sliststr = append(sliststr, fmt.Sprint(j))
		}

		var blist []Booktype
		err := Db.Select(&blist, "SELECT * FROM book WHERE `bid` in ("+strings.Join(sliststr, ",")+") order by field(bid, "+strings.Join(sliststr, ",")+")")
		if err != nil {
			log.SetFlags(log.Lshortfile | log.LstdFlags)
			log.Println(err)
		}

		return blist
	} else {
		return []Booktype{}
	}
}

func favoritesADSql(bid, uid string) bool {
	defer func() {
		err := recover()
		if err != nil {
			return
		}
	}()

	var userBookList []favoritestype
	uidInt, _ := strconv.Atoi(uid)
	err := Db.Select(&userBookList, "SELECT bookid FROM `favorites` WHERE uid=?", uidInt)
	if err != nil {
		log.SetFlags(log.Lshortfile | log.LstdFlags)
		log.Println(err)
	}

	var booklist []int
	if userBookList != nil {
		json.Unmarshal([]byte(userBookList[0].Bookid), &booklist)
	} else {
		booklist = []int{}
	}

	bid_int, err := strconv.Atoi(bid)
	if err != nil {
		log.SetFlags(log.Lshortfile | log.LstdFlags)
		log.Println(err)
	}

	var d bool

	var booklistJson []byte
	if intInSlice(bid_int, booklist) {
		booklist = SliceRemoveDuplicates(delIntSlice(booklist, bid_int))
		booklistJson, _ = json.Marshal(booklist)
		d = false
	} else {
		booklist = SliceRemoveDuplicates(append(booklist, bid_int))
		booklistJson, _ = json.Marshal(booklist)
		addFavorites(string(booklistJson))
		d = true
	}

	uidInt, _ = strconv.Atoi(uid)
	_, err = Db.Exec("REPLACE INTO `favorites` (`uid`, `bookid`) VALUES (?, ?)", uidInt, string(booklistJson))
	if err != nil {
		log.SetFlags(log.Lshortfile | log.LstdFlags)
		log.Println(err)
	}

	return d
}

func favoritesSelectSql(uid string) []int {
	defer func() {
		err := recover()
		if err != nil {
			return
		}
	}()

	var userBookList []favoritestype
	uidInt, _ := strconv.Atoi(uid)
	err := Db.Select(&userBookList, "SELECT bookid FROM `favorites` WHERE uid=?", uidInt)
	if err != nil {
		log.SetFlags(log.Lshortfile | log.LstdFlags)
		log.Println(err)
	}

	var booklist []int
	if userBookList != nil {
		json.Unmarshal([]byte(userBookList[0].Bookid), &booklist)
	} else {
		booklist = []int{}
	}

	return booklist
}

func favoritesbookList(uid string) []int {
	var userBookList []favoritestype
	uidInt, _ := strconv.Atoi(uid)
	err := Db.Select(&userBookList, "SELECT bookid FROM `favorites` WHERE uid=?", uidInt)
	if err != nil {
		log.SetFlags(log.Lshortfile | log.LstdFlags)
		log.Println(err)
	}

	var booklist []int
	if userBookList != nil {
		json.Unmarshal([]byte(userBookList[0].Bookid), &booklist)
	} else {
		booklist = []int{}
	}

	return booklist
}

func favoritesbookListnum(uid string) int {
	var userBookList []favoritestype
	uidInt, _ := strconv.Atoi(uid)
	err := Db.Select(&userBookList, "SELECT bookid FROM `favorites` WHERE uid=?", uidInt)
	if err != nil {
		fmt.Println(err)
	}

	if userBookList != nil {
		var booklist []int
		json.Unmarshal([]byte(userBookList[0].Bookid), &booklist)
		return len(booklist)
	} else {
		return 0
	}
}

func getRandom(num int) []Booktype {
	defer func() {
		err := recover()
		if err != nil {
			return
		}
	}()

	var booklist []Booktype
	err := Db.Select(&booklist, "SELECT * FROM book ORDER BY RAND() LIMIT ?", num)
	if err != nil {
		log.SetFlags(log.Lshortfile | log.LstdFlags)
		log.Println(err)
	}

	return booklist
}

func getCommentDb(bid, page int) []CommInfoUser {
	defer func() {
		err := recover()
		if err != nil {
			return
		}
	}()

	var commentInfo []CommInfoUser
	err := Db.Select(&commentInfo, "SELECT cid, uid, username, email_md5, area, reply, platform, time, content FROM comment WHERE `bid`=? ORDER BY `cid` DESC LIMIT ?,40", bid, page)
	if err != nil {
		log.SetFlags(log.Lshortfile | log.LstdFlags)
		log.Println("SqlError", err)
	}
	for i := range commentInfo {
		if commentInfo[i].Reply != -1 {
			commentInfo[i].ReplyText = selectComment(commentInfo[i].Reply)
		}
	}

	var uidList []string
	for _, j := range commentInfo {
		uidList = append(uidList, fmt.Sprint(j.Uid))
	}
	var userinfo []userAvatar = []userAvatar{}
	if len(uidList) != 0 {
		err = Db.Select(&userinfo, "SELECT uid, username, avatar FROM user WHERE `uid` in ("+strings.Join(uidList, ",")+") AND `avatar` <> 'default.webp'")
		if err != nil {
			log.SetFlags(log.Lshortfile | log.LstdFlags)
			log.Println("SqlError", err)
		}
	}
	var userAvatarMap map[int]string = make(map[int]string)
	for _, j := range userinfo {
		userAvatarMap[j.Uid] = j.Avatar
	}

	for i := range commentInfo {
		commentInfo[i].AvatarUrl = userAvatarMap[commentInfo[i].Uid]
	}

	return commentInfo
}

func getCommentLen(bid int) int {
	defer func() {
		err := recover()
		if err != nil {
			return
		}
	}()

	type CommentLen struct {
		Len int `db:"len"`
	}

	var commentlen []CommentLen
	err := Db.Select(&commentlen, "SELECT COUNT(cid) as `len` FROM comment WHERE `bid`=?", bid)
	if err != nil {
		log.SetFlags(log.Lshortfile | log.LstdFlags)
		log.Println("SqlError", err)
	}

	return commentlen[0].Len
}

func sendCommentDb(bid, uid int, username, email, emailMd5, ip, area string, reply int, platform, content string) string {
	defer func() {
		err := recover()
		if err != nil {
			return
		}
	}()

	var commentInfo []CommInfo
	err := Db.Select(&commentInfo, "INSERT INTO comment (`bid`, `uid`, `username`, `email`, `email_md5`, `ip`, `area`, `reply`, `platform`, `time`, `content`) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", bid, uid, username, email, emailMd5, ip, area, reply, platform, time.Now().Unix(), content)
	if err != nil {
		log.SetFlags(log.Lshortfile | log.LstdFlags)
		log.Println("SqlError", err)
		return "error"
	}
	return "ok"
}

func selectComment(cid int) string {
	defer func() {
		err := recover()
		if err != nil {
			return
		}
	}()

	var commentInfo []CommInfoUser
	err := Db.Select(&commentInfo, "SELECT username, content FROM comment WHERE `cid`=?", cid)
	if err != nil {
		log.SetFlags(log.Lshortfile | log.LstdFlags)
		log.Println("SqlError", err)
		return "error"
	}
	return commentInfo[0].Username + ": " + commentInfo[0].Content
}

func lastCommentTime(uid int) int64 {
	defer func() {
		err := recover()
		if err != nil {
			return
		}
	}()

	var commentInfo []CommInfoUser
	err := Db.Select(&commentInfo, "SELECT time FROM comment WHERE `uid`=? ORDER BY `cid` DESC LIMIT 0,1", uid)
	if err != nil {
		log.SetFlags(log.Lshortfile | log.LstdFlags)
		log.Println("SqlError", err)
		return 0
	}
	if len(commentInfo) != 0 {
		return int64(commentInfo[0].Time)
	} else {
		return 0
	}
}

func getReadTags() []string {
	var readL []readLeaderboard
	err := Db.Select(&readL, "SELECT `bid` FROM `read_leaderboard` ORDER BY `day` DESC LIMIT 0,10")
	if err != nil {
		log.SetFlags(log.Lshortfile | log.LstdFlags)
		log.Println("SqlError", err)
		return nil
	}

	bidlist := []string{}
	for i := range readL {
		bidlist = append(bidlist, fmt.Sprint(readL[i].Bid))
	}

	var bookinfo []Booktype
	err = Db.Select(&bookinfo, "SELECT `ptag` FROM book WHERE `bid` in ("+strings.Join(bidlist, ",")+")")
	if err != nil {
		fmt.Println("SqlError", err)
		return nil
	}

	taglist := []string{}
	for i := range bookinfo {
		taglist = append(taglist, strings.Split(bookinfo[i].Ptag, " ")...)
	}

	return SliceRemoveDuplicatesString(taglist)
}

func getFs(uid, page int) ([]Booktype, int) {
	pagenum := (page - 1) * 20

	var data struct {
		Uid  int   `db:"uid"`
		Data []int `db:"data"`
	}
	mongoCollectionUser.FindOne(ctx, bson.D{{Key: "uid", Value: uid}}).Decode(&data)

	if len(data.Data) == 0 {
		return []Booktype{}, 0
	}

	fsJsonThisPage := data.Data[pagenum:operator3(pagenum+20 >= len(data.Data), len(data.Data), pagenum+20)]
	var fsJsonStr []string
	for i := range fsJsonThisPage {
		fsJsonStr = append(fsJsonStr, fmt.Sprint(fsJsonThisPage[i]))
	}

	var bookInfo []Booktype
	idStr := strings.Join(fsJsonStr, ",")
	err := Db.Select(&bookInfo, "SELECT * FROM `book` WHERE `bid` in ("+idStr+") ORDER BY field(bid,"+idStr+")")
	if err != nil {
		log.SetFlags(log.Lshortfile | log.LstdFlags)
		log.Println("SqlError", err)
		return []Booktype{}, 0
	}
	return bookInfo, len(data.Data)
}

func getReadLeaderboardCount(ltype string) int {
	var Num []readLeaderboardNum
	err := Db.Select(&Num, "SELECT count(bid) FROM `read_leaderboard` WHERE `"+ltype+"` <> 0")
	if err != nil {
		log.SetFlags(log.Lshortfile | log.LstdFlags)
		log.Println("SqlError", err)
		return 0
	}
	return Num[0].Num
}

func getReadLeaderboardBook(lbType string, page int) []Booktype {
	var lb []readLeaderboard
	err := Db.Select(&lb, "SELECT `bid`, `day`, `week`, `moon` FROM `read_leaderboard` ORDER BY `"+lbType+"` DESC LIMIT ?,20", page)
	if err != nil {
		log.SetFlags(log.Lshortfile | log.LstdFlags)
		log.Println("SqlError", err)
		return nil
	}
	var bookid []string = []string{}
	for i := range lb {
		bookid = append(bookid, fmt.Sprint(lb[i].Bid))
	}
	return getBookListByList(strings.Join(bookid, ","))
}

func getFavLeaderboardCount() int {
	var Num []readLeaderboardNum
	err := Db.Select(&Num, "SELECT count(bid) FROM `fav_leaderboard`")
	if err != nil {
		log.SetFlags(log.Lshortfile | log.LstdFlags)
		log.Println("SqlError", err)
		return 0
	}
	return Num[0].Num
}

func getFavLeaderboardBook(lbType string, page int) []Booktype {
	var lb []readLeaderboard
	err := Db.Select(&lb, "SELECT `bid`, `day`, `week`, `moon` FROM `fav_leaderboard` ORDER BY `"+lbType+"` DESC LIMIT ?,20", page)
	if err != nil {
		log.SetFlags(log.Lshortfile | log.LstdFlags)
		log.Println("SqlError", err)
		return nil
	}
	var bookid []string = []string{}
	for i := range lb {
		bookid = append(bookid, fmt.Sprint(lb[i].Bid))
	}
	return getBookListByList(strings.Join(bookid, ","))
}

func getProportionCount() int64 {
	itemCount, err := mongoCollectionProportion.CountDocuments(ctx, bson.M{})
	if err != nil {
		fmt.Println("MongoErr", err)
	}

	return itemCount
}

func getProportion(page int) []Booktype {
	type mongoDataType struct {
		Bid int     `db:"bid"`
		P   float64 `db:"p"`
	}

	Options := options.Find()
	Options.SetSort(bson.D{{Key: "p", Value: 1}, {Key: "bid", Value: -1}})
	Options.SetLimit(20)
	Options.SetSkip(int64((page - 1) * 20))

	cur, err := mongoCollectionProportion.Find(ctx, bson.D{{}}, Options)
	if err != nil {
		fmt.Println("MongoErr", err)
	}
	defer cur.Close(ctx)

	var uidList []string = []string{}

	for cur.Next(ctx) {
		var data mongoDataType
		err := cur.Decode(&data)
		if err != nil {
			log.Fatal(err)
		}

		uidList = append(uidList, fmt.Sprint(data.Bid))
	}

	var booklist []Booktype
	bidlistStr := strings.Join(uidList, ",")
	err = Db.Select(&booklist, "SELECT bid, bookname, author, pname, ptag, otag, time, views, favorites FROM book WHERE `bid` in ("+bidlistStr+") and `time` < ? ORDER BY field(bid,"+bidlistStr+")", time.Now().Unix()-60*60*24*7)
	if err != nil {
		fmt.Println("SqlServer", err)
	}

	return booklist
}

func getBadge(uid string) []Badge {
	var userinfo []UserInfoS
	err := Db.Select(&userinfo, "SELECT badge FROM user WHERE `uid`=?", uid)
	if err != nil {
		log.SetFlags(log.Lshortfile | log.LstdFlags)
		log.Println("SQL Error:", err)
	}

	badgeId := []string{}
	json.Unmarshal(userinfo[0].Badge, &badgeId)
	var badgeInfo []Badge = []Badge{}
	Db.Select(&badgeInfo, "SELECT name, img, minimg, des FROM badge WHERE id in ("+strings.Join(badgeId, ",")+")")

	return badgeInfo
}

func getCommunityLen() int {
	type CommunityLen struct {
		Len int `db:"len"`
	}

	var communitylen []CommunityLen = []CommunityLen{}
	err := Db.Select(&communitylen, "SELECT COUNT(id) as `len` FROM community WHERE `expired` > ?", time.Now().Unix())
	if err != nil {
		log.SetFlags(log.Lshortfile | log.LstdFlags)
		log.Println("SqlError", err)
	}

	return communitylen[0].Len
}

func getCommunityDb(page int) []Community {
	var community []CommunityDb = []CommunityDb{}
	err := Db.Select(&community, "SELECT `id`, `uid`, `title`, `content`, `image`, `time`, `expired`, `area`, `email_md5`, `platform` FROM community WHERE `expired` > ? ORDER BY `id` DESC LIMIT ?,40", time.Now().Unix(), page)
	if err != nil {
		log.SetFlags(log.Lshortfile | log.LstdFlags)
		log.Println("SqlError", err)
		return []Community{}
	}

	var uidList []string = []string{}
	for _, j := range community {
		uidList = append(uidList, fmt.Sprint(j.Uid))
	}

	var userinfo []UserInfoS
	err = Db.Select(&userinfo, "SELECT `uid`, `username`, `avatar` FROM user WHERE `uid` in ("+strings.Join(uidList, ",")+")")
	if err != nil {
		log.SetFlags(log.Lshortfile | log.LstdFlags)
		log.Println("SqlError", err)
		return []Community{}
	}
	var userinfoMap map[int]map[string]string = make(map[int]map[string]string)
	for _, j := range userinfo {
		var avatar string = ""
		if j.Avatar != "default.webp" {
			avatar = j.Avatar
		}
		userinfoMap[j.Uid] = map[string]string{
			"username": j.Username,
			"avatar":   avatar,
		}
	}

	var imageMap map[int][]string = make(map[int][]string)
	for _, j := range community {
		var imagelist []string = []string{}
		json.Unmarshal(j.Image, &imagelist)
		imageMap[j.Id] = imagelist
	}

	var communityData []Community
	for _, j := range community {
		var avatar string
		if userinfoMap[j.Uid]["avatar"] == "" {
			avatar = j.EmailMd5
		} else {
			avatar = userinfoMap[j.Uid]["avatar"]
		}
		communityData = append(communityData, Community{
			Id:       j.Id,
			Username: userinfoMap[j.Uid]["username"],
			Title:    j.Title,
			Content:  j.Content,
			Image:    imageMap[j.Id],
			Time:     j.Time,
			Expired:  j.Expired,
			Area:     j.Area,
			Avatar:   avatar,
			Platform: j.Platform,
		})
	}

	return communityData
}
