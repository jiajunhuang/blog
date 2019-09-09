package main

import (
	"bufio"
	"fmt"
	"html/template"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/getsentry/raven-go"
	"github.com/gin-contrib/sentry"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/russross/blackfriday"
	"go.uber.org/zap"
)

var (
	logger, _ = zap.NewProduction()
	sugar     = logger.Sugar()

	filenameRegex = regexp.MustCompile(`(\d{4}_\d{2}_\d{2})-.+\..+`)
	articles      = LoadMDs("articles")

	db *sqlx.DB

	categoryMap = map[string]string{
		"golang":         "Golang简明教程",
		"python":         "Python教程",
		"data_structure": "数据结构在实际项目中的使用",
	}
)

// InitSentry 初始化sentry
func InitSentry() {
	raven.SetDSN(os.Getenv("SENTRY_DSN"))
}

// InitializeDB 初始化数据库连接
func InitializeDB() {
	var err error

	db, err = sqlx.Connect("sqlite3", os.Getenv("SQLX_URL"))
	if err != nil {
		sugar.Fatalf("failed to connect to the db: %s", err)
	}
}

// Article 就是文章
type Article struct {
	Title    string
	Date     string
	Filename string
	DirName  string
}

// Articles 文章列表
type Articles []Article

func (a Articles) Len() int      { return len(a) }
func (a Articles) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a Articles) Less(i, j int) bool {
	v := strings.Compare(a[i].Date, a[j].Date)
	if v <= 0 {
		return true
	}

	return false
}
func (a Articles) RandomN(n int) Articles {
	if n <= 0 {
		return nil
	}

	length := len(a)

	pos := rand.Intn(length - n)
	return a[pos : pos+n]
}

func getFilePath(path string) string {
	suffix := ".html"
	if strings.HasSuffix(path, suffix) {
		path = path[:len(path)-len(suffix)]
	}
	return "./" + path
}

// ReadTitle 把标题读出来
func ReadTitle(path string) string {
	path = getFilePath(path)

	file, err := os.Open(path)
	if err != nil {
		sugar.Errorf("failed to read file(%s): %s", path, err)
		return ""
	}
	line, _, err := bufio.NewReader(file).ReadLine()
	if err != nil {
		sugar.Errorf("failed to read title of file(%s): %s", path, err)
		return ""
	}
	title := strings.Replace(string(line), "# ", "", -1)

	return title
}

// LoadArticle 把文章的元信息读出来
func LoadArticle(dirname, filename string) *Article {
	match := filenameRegex.FindStringSubmatch(filename)
	if len(match) != 2 {
		return nil
	}

	dateString := match[1]
	filepath := fmt.Sprintf("./%s/%s", dirname, filename)
	title := ReadTitle(filepath)

	return &Article{
		Title:    title,
		Date:     strings.Replace(dateString, "_", "-", -1),
		Filename: filename,
		DirName:  dirname,
	}
}

// LoadMDs 读取给定目录中的所有markdown文章
func LoadMDs(dirname string) Articles {
	files, err := ioutil.ReadDir(dirname)
	if err != nil {
		sugar.Fatalf("failed to read dir(%s): %s", dirname, err)
		return nil
	}

	var articles Articles
	for _, file := range files {
		filename := file.Name()
		if article := LoadArticle(dirname, filename); article != nil {
			articles = append(articles, *article)
		}
	}

	sort.Sort(sort.Reverse(articles))

	return articles
}

// IndexHandler 首页
func IndexHandler(c *gin.Context) {
	c.HTML(
		http.StatusOK, "index.html", gin.H{
			"articles":    articles[:80],
			"totalCount":  len(articles),
			"keywords":    "Golang,Python,Go语言,分布式,高并发,Haskell,C,微服务,软件工程,源码阅读,源码分析",
			"description": "享受技术带来的快乐~分布式系统/高并发处理/Golang/Python/Haskell/C/微服务/软件工程/源码阅读与分析",
		},
	)
}

// ArchiveHandler 全部文章
func ArchiveHandler(c *gin.Context) {
	c.HTML(
		http.StatusOK, "index.html", gin.H{
			"articles":    articles,
			"keywords":    "Golang,Python,Go语言,分布式,高并发,Haskell,C,微服务,软件工程,源码阅读,源码分析",
			"description": "享受技术带来的快乐~分布式系统/高并发处理/Golang/Python/Haskell/C/微服务/软件工程/源码阅读与分析",
		},
	)
}

func renderArticle(c *gin.Context, status int, path string, subtitle string, randomN int) {
	path = getFilePath(path)
	content, err := ioutil.ReadFile(path)
	if err != nil {
		sugar.Errorf("failed to read file %s: %s", path, err)
		c.Redirect(http.StatusFound, "/404")
		return
	}

	content = blackfriday.Run(
		content,
		blackfriday.WithExtensions(blackfriday.FencedCode),
	)

	recommends := articles.RandomN(randomN)

	c.HTML(
		status, "article.html", gin.H{
			"content":    template.HTML(content),
			"title":      ReadTitle(path),
			"subtitle":   subtitle,
			"recommends": recommends,
		},
	)
}

// ArticleHandler 具体文章
func ArticleHandler(c *gin.Context) {
	renderArticle(c, http.StatusOK, c.Request.URL.Path, "", 10)
}

// AboutMeHandler 关于我
func AboutMeHandler(c *gin.Context) {
	renderArticle(c, http.StatusOK, "articles/aboutme.md", "", 0)
}

// FriendsHandler 友链
func FriendsHandler(c *gin.Context) {
	renderArticle(c, http.StatusOK, "articles/friends.md", "", 0)
}

// NotFoundHandler 404
func NotFoundHandler(c *gin.Context) {
	renderArticle(c, http.StatusOK, "articles/404.md", "", 0)
}

// AllSharingHandler 所有分享
func AllSharingHandler(c *gin.Context) {
	sharing := dao.GetAllSharing()

	c.HTML(
		http.StatusOK, "list.html", gin.H{
			"sharing": sharing,
		},
	)
}

// SharingHandler 分享
func SharingHandler(c *gin.Context) {
	sharing := dao.GetSharingWithLimit(20)

	c.HTML(
		http.StatusOK, "list.html", gin.H{
			"sharing": sharing,
			"partly":  true,
		},
	)
}

// NotesHandler 随想
func NotesHandler(c *gin.Context) {
	notes := dao.GetAllNotes()

	c.HTML(
		http.StatusOK, "list.html", gin.H{
			"notes": notes,
		},
	)
}

// RSSHandler RSS
func RSSHandler(c *gin.Context) {
	c.Header("Content-Type", "application/xml")
	c.HTML(
		http.StatusOK, "rss.html", gin.H{
			"rssHeader": template.HTML(`<?xml version="1.0" encoding="UTF-8"?>`),
			"articles":  articles,
		},
	)
}

// SiteMapHandler sitemap
func SiteMapHandler(c *gin.Context) {
	c.Header("Content-Type", "application/xml")
	c.HTML(
		http.StatusOK, "sitemap.html", gin.H{
			"rssHeader": template.HTML(`<?xml version="1.0" encoding="UTF-8"?>`),
			"articles":  articles,
		},
	)
}

// TutorialHandler 教程
func TutorialHandler(c *gin.Context) {
	category := c.Param("category")
	filename := c.Param("filename")

	renderArticle(c, http.StatusOK, fmt.Sprintf("tutorial/%s/%s", category, filename), categoryMap[category], 0)
}

// SearchHandler 搜索
func SearchHandler(c *gin.Context) {
	word := c.PostForm("search")

	c.Redirect(
		http.StatusFound,
		"https://www.google.com/search?q=site:jiajunhuang.com "+word,
	)
}

// RewardHandler 扫码赞赏
func RewardHandler(c *gin.Context) {
	userAgent := c.Request.UserAgent()
	if strings.Contains(userAgent, "MicroMessenger") {
		c.Redirect(http.StatusFound, os.Getenv("WECHAT_PAY_URL"))
		return
	}

	c.Redirect(http.StatusFound, os.Getenv("ALIPAY_URL"))
}

func main() {
	defer logger.Sync() // flushes buffer, if any

	// telegram bot
	go startNoteBot()
	go startSharingBot()

	InitializeDB()
	InitSentry()

	r := gin.New()

	r.Use(gin.Logger())
	r.Use(sentry.Recovery(raven.DefaultClient, false))

	r.LoadHTMLGlob("templates/*.html")
	r.Static("/static", "./static")
	//r.Static("/tutorial/:lang/img/", "./tutorial/:lang/img")  # 然而不支持
	//r.Static("/articles/img", "./articles/img")  # 然而有冲突
	r.StaticFile("/favicon.ico", "./static/favicon.ico")
	r.StaticFile("/robots.txt", "./static/robots.txt")
	r.StaticFile("/ads.txt", "./static/ads.txt")

	r.GET("/", IndexHandler)
	r.GET("/404", NotFoundHandler)
	r.GET("/archive", ArchiveHandler)
	r.GET("/articles/:filepath", ArticleHandler)
	r.GET("/aboutme", AboutMeHandler)
	r.GET("/friends", FriendsHandler)
	r.GET("/sharing", SharingHandler)
	r.GET("/sharing/all", AllSharingHandler)
	r.GET("/notes", NotesHandler)
	r.GET("/rss", RSSHandler)
	r.GET("/sitemap.xml", SiteMapHandler)
	r.GET("/tutorial/:category/:filename", TutorialHandler)
	r.GET("/reward", RewardHandler)
	r.POST("/search", SearchHandler)
	r.NoRoute(func(c *gin.Context) { c.Redirect(http.StatusFound, "/404") })

	r.Run("127.0.0.1:8080")
}
