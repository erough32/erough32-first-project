package main

import (
	"fmt"
	"io/ioutil"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
)

// HTTP
func cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("server", "NoyPL 1.0")
		c.Writer.Header().Set("Content-Type", "text/html;charset=utf-8")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	store, _ := redis.NewStore(10, "tcp", sessionsConfig("server"), sessionsConfig("password"), []byte("secret"))
	router.Use(sessions.Sessions("NOY_SESSION", store))
	router.Use(cors())
	router.Use(gin.Recovery())
	// 404
	router.NoRoute(func(context *gin.Context) {
		html, _ := ioutil.ReadFile("error/404.html")
		context.String(404, string(html))
	})
	// Status
	router.GET("/status", func(context *gin.Context) {
		context.Writer.Header().Set("Content-Type", "application/json;charset=utf-8")
		html := `{"status":"ok"}`

		context.String(200, string(html))
	})

	router.GET("/ads", ads) // 首頁 ADS (舊版本)

	router.POST("/api/login", loginApi)              // 登入
	router.POST("/ads", ads)                         // 首頁 ADS
	router.POST("/update", update)                   // 獲取APP更新訊息
	router.POST("/api/booklist", BooklistPage)       // 本子列表
	router.POST("/api/booknum", Booknumpage)         // 本子數
	router.POST("/api/userinfo", userselect)         // 用戶資訊
	router.POST("/api/selectSession", selectSession) // 用戶登錄狀態
	router.POST("/api/recommend", recommend)         // 相似推薦算法 (本子相似推薦)
	router.POST("/api/getbookinfo", getBookInfo)     // 獲取本子資訊
	router.POST("/api/adfavorites", adfavorites)     // 修改收藏狀態
	router.POST("/api/favoriteslist", favoriteslist) // 獲取收藏列表
	router.POST("/api/favoritesnum", favoritesnum)   // 獲取收藏數量
	router.POST("/api/history", history)             // 歷史記錄批量查詢

	// V2
	router.POST("/api/ads_v2", AdsV2)                          // ADS V2
	router.POST("/api/home", homeApi)                          // Home
	router.POST("/api/newReg", newRegApi)                      // 註冊
	router.POST("/api/getbooklen", getbooklen)                 // 本子長度
	router.POST("/api/randomBook", randomBook)                 // 隨機本子
	router.POST("/api/search_v2", searchNew)                   // 搜尋API
	router.POST("/api/userinfo_v2", getUserInfoV2)             // 使用者資訊
	router.POST("/api/getDownload", getDownload)               // 下載Token
	router.POST("/api/ftag", ftag)                             // 隨機tag推薦
	router.POST("/api/bulletin", bulletin)                     // 公告板
	router.POST("/api/history_v2", historyV2)                  // 歷史記錄批量查詢
	router.POST("/api/favoriteslist_v2", favoriteslistV2)      // 收藏列表
	router.POST("/api/readLeaderboard", getReadLeaderboard)    // 閱讀榜
	router.POST("/api/favLeaderboard", getFavLeaderboard)      // 收藏榜
	router.POST("/api/favoritesrecommend", favoritesrecommend) // 收藏推薦
	router.POST("/api/login_v2", loginApiv2)                   // 登入
	router.POST("/api/booklist_v2", booklistV2)                // 本子列表
	router.POST("/api/getcredit", getcredit)                   // 獲取使用者積分資訊
	router.POST("/api/bigtaglist", bigtaglist)                 // 大分區
	router.POST("/api/getMsg", getMsg)                         // 獲取回復評論
	router.POST("/api/duplicate_username", duplicateUsername)  // 使用者名稱重複檢查
	router.POST("/api/feedback", feedback)                     // Feedback
	router.POST("/api/reset_name", resetUsername)              // 重設使用者名稱
	router.POST("/api/proportion", proportion)                 // 收藏閱讀比排行榜
	router.POST("/api/badge", badge)                           // 徽章

	// V3
	router.POST("/api/history_v3", historyV3) // 歷史記錄批量查詢

	router.POST("/api/v3/bookinfo", book)                           // BookInfo
	router.POST("/api/v3/community_list", communityList)            // 社區貼文列表
	router.POST("/api/v3/send_community", sendCommunity)            // 傳送貼文
	router.POST("/api/v3/get_community_post", getCommunityPost)     // 獲取貼文
	router.POST("/api/v3/send_community_reply", sendCommunityReply) // 傳送回覆

	// ImageBucket
	router.POST("/api/get_image_bucket_token", getImageBucketToken) // 圖片桶上載 Token

	// Comment
	router.POST("/api/getComment", getComment)
	router.POST("/api/sendComment", sendComment)

	// ImageLib
	router.GET("/api/useI", useI)                          // 使用積分
	router.GET("/api/downloads_credit", downloadCreditApi) // 使用積分（傳入頁數）

	router.POST("/api/seo", seo) // SEO

	router.GET("/cron", Cron) // Cron

	// Run
	fmt.Println("NoyPL Starting ...")
	router.Run(":80")
}
