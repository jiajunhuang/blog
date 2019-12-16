package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/getsentry/raven-go"
	"github.com/gin-contrib/sentry"
	"github.com/gin-gonic/gin"
	redis "github.com/go-redis/redis/v7"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/russross/blackfriday"
)

const (
	zsetKey = "blogtopn"
)

var (
	filenameRegex = regexp.MustCompile(`(\d{4}_\d{2}_\d{2})-.+\..+`)
	articles      = LoadMDs("articles")

	db *sqlx.DB

	redisClient *redis.Client

	categoryMap = map[string]string{
		"golang":         "Golang简明教程",
		"python":         "Python教程",
		"data_structure": "数据结构在实际项目中的使用",
	}

	// ErrNotFound means article not found
	ErrNotFound = errors.New("Article Not Found")
	// ErrFailedToLoad failed to load article
	ErrFailedToLoad = errors.New("Failed To Load Article")

	// Prometheus
	totalRequests = promauto.NewCounter(prometheus.CounterOpts{Name: "total_requests_total"})
)

// InitSentry 初始化sentry
func InitSentry() {
	raven.SetDSN(os.Getenv("SENTRY_DSN"))
}

// InitializeDB 初始化数据库连接
func InitializeDB() {
	var err error

	db, err = sqlx.Connect("mysql", os.Getenv("SQLX_URL"))
	if err != nil {
		log.Fatalf("failed to connect to the db: %s", err)
	}
}

// InitializeRedis 初始化Redis
func InitializeRedis() {
	opt, err := redis.ParseURL(os.Getenv("REDIS_URL"))
	if err != nil {
		log.Fatalf("failed to connect to redis db: %s", err)
	}

	// Create client as usually.
	redisClient = redis.NewClient(opt)
}

// Article 就是文章
type Article struct {
	Title       string    `json:"title"`
	Date        string    `json:"date_str"`
	Filename    string    `json:"file_name"`
	DirName     string    `json:"dir_name"`
	PubDate     time.Time `json:"-"`
	Description string    `json:"description"`
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

// RandomN return n articles by random
func (a Articles) RandomN(n int) Articles {
	if n <= 0 {
		return nil
	}

	length := len(a)

	pos := rand.Intn(length - n)
	return a[pos : pos+n]
}

func isBlogApp(c *gin.Context) bool {
	ua := c.GetHeader("User-Agent")
	if strings.HasPrefix(ua, "BlogApp/") {
		return true
	}

	return false
}

func getFilePath(path string) string {
	suffix := ".html"
	if strings.HasSuffix(path, suffix) {
		path = path[:len(path)-len(suffix)]
	}
	return "./" + path
}

// ReadDesc 把简介读出来
func ReadDesc(path string) string {
	path = getFilePath(path)

	file, err := os.Open(path)
	if err != nil {
		log.Printf("failed to read file(%s): %s", path, err)
		return ""
	}
	reader := bufio.NewReader(file)
	reader.ReadLine() // 忽略第一行(标题)
	reader.ReadLine() // 忽略第二行(空行)
	desc := ""
	for i := 0; i < 3; i++ {
		line, _, err := reader.ReadLine()
		if err != nil && err != io.EOF {
			log.Printf("failed to read desc of file(%s): %s", path, err)
			continue
		}
		desc += string(line)
	}

	trimChars := "\n，。：,.:"
	return strings.TrimRight(strings.TrimLeft(desc, trimChars), trimChars) + "..."
}

// ReadTitle 把标题读出来
func ReadTitle(path string) string {
	path = getFilePath(path)

	file, err := os.Open(path)
	if err != nil {
		log.Printf("failed to read file(%s): %s", path, err)
		return ""
	}
	line, _, err := bufio.NewReader(file).ReadLine()
	if err != nil {
		log.Printf("failed to read title of file(%s): %s", path, err)
		return ""
	}
	title := strings.Replace(string(line), "# ", "", -1)

	return title
}

// VisitedArticle is for remember which article had been visited
type VisitedArticle struct {
	URLPath string `json:"url_path"`
	Title   string `json:"title"`
}

func genVisited(urlPath, subTitle string) (string, error) {
	title := ReadTitle(urlPath)
	if title == "" {
		return "", ErrNotFound
	}

	if subTitle != "" {
		title += " - " + subTitle
	}

	visited := VisitedArticle{URLPath: urlPath, Title: title}
	b, err := json.Marshal(visited)
	if err != nil {
		return "", ErrFailedToLoad
	}

	return string(b), nil
}

func getTopVisited(n int) []VisitedArticle {
	visitedArticles := []VisitedArticle{}

	articles, err := redisClient.ZRevRangeByScore(zsetKey, &redis.ZRangeBy{
		Min: "-inf", Max: "+inf", Offset: 0, Count: int64(n),
	}).Result()
	if err != nil {
		log.Printf("failed to get top %d visited articles: %s", n, err)
		return nil
	}

	for _, article := range articles {
		var va VisitedArticle
		if err := json.Unmarshal([]byte(article), &va); err != nil {
			log.Printf("failed to unmarshal article: %s", err)
			continue
		}

		visitedArticles = append(visitedArticles, va)
	}

	return visitedArticles
}

// LoadArticle 把文章的元信息读出来
func LoadArticle(dirname, filename string) *Article {
	match := filenameRegex.FindStringSubmatch(filename)
	if len(match) != 2 {
		return nil
	}

	dateString := strings.Replace(match[1], "_", "-", -1)
	filepath := fmt.Sprintf("./%s/%s", dirname, filename)
	title := ReadTitle(filepath)
	pubDate, err := time.Parse("2006-01-02", dateString)
	if err != nil {
		log.Panicf("failed to parse date: %s", err)
	}
	desc := ReadDesc(filepath)

	return &Article{
		Title:       title,
		Date:        dateString,
		Filename:    filename,
		DirName:     dirname,
		PubDate:     pubDate,
		Description: desc,
	}
}

// LoadMDs 读取给定目录中的所有markdown文章
func LoadMDs(dirname string) Articles {
	files, err := ioutil.ReadDir(dirname)
	if err != nil {
		log.Fatalf("failed to read dir(%s): %s", dirname, err)
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
	topArticles := getTopVisited(15)
	c.HTML(
		http.StatusOK, "index.html", gin.H{
			"isBlogApp":   isBlogApp(c),
			"articles":    articles[:100],
			"totalCount":  len(articles),
			"keywords":    "Golang,Python,Go语言,Dart,Flutter,分布式,高并发,Haskell,C,微服务,软件工程,源码阅读,源码分析",
			"description": "享受技术带来的快乐~分布式系统/高并发处理/Golang/Python/Haskell/C/微服务/Flutter/软件工程/源码阅读与分析",
			"topArticles": topArticles,
		},
	)
}

// ArchiveHandler 全部文章
func ArchiveHandler(c *gin.Context) {
	c.HTML(
		http.StatusOK, "index.html", gin.H{
			"isBlogApp":   isBlogApp(c),
			"articles":    articles,
			"keywords":    "Golang,Python,Go语言,Dart,Flutter,分布式,高并发,Haskell,C,微服务,软件工程,源码阅读,源码分析",
			"description": "享受技术带来的快乐~分布式系统/高并发处理/Golang/Python/Haskell/C/微服务/Flutter/软件工程/源码阅读与分析",
		},
	)
}

func renderArticle(c *gin.Context, status int, path string, subtitle string, randomN int) {
	path = getFilePath(path)
	content, err := ioutil.ReadFile(path)
	if err != nil {
		log.Printf("failed to read file %s: %s", path, err)
		c.Redirect(http.StatusFound, "/404")
		return
	}

	content = blackfriday.Run(
		content,
		blackfriday.WithExtensions(blackfriday.FencedCode),
	)

	recommends := articles.RandomN(randomN)
	topArticles := getTopVisited(15)

	c.HTML(
		status, "article.html", gin.H{
			"isBlogApp":   isBlogApp(c),
			"content":     template.HTML(content),
			"title":       ReadTitle(path),
			"subtitle":    subtitle,
			"recommends":  recommends,
			"topArticles": topArticles,
		},
	)
}

func incrVisited(urlPath, subTitle string) {
	if visited, err := genVisited(urlPath, subTitle); err != nil {
		log.Printf("failed to gen visited: %s", err)
	} else {
		if _, err := redisClient.ZIncrBy(zsetKey, 1, visited).Result(); err != nil {
			log.Printf("failed to incr score of %s: %s", urlPath, err)
		}
	}
}

// PingPongHandler ping pong
func PingPongHandler(c *gin.Context) {
	c.JSON(http.StatusOK, nil)
}

// ArticleHandler 具体文章
func ArticleHandler(c *gin.Context) {
	urlPath := c.Request.URL.Path
	incrVisited(urlPath, "")

	renderArticle(c, http.StatusOK, urlPath, "", 15)
}

// FlutterHandler 渲染Flutter专页
func FlutterHandler(c *gin.Context) {
	renderArticle(c, http.StatusOK, "articles/flutter.md", "", 0)
}

// TutorialPageHandler 教程index
func TutorialPageHandler(c *gin.Context) {
	renderArticle(c, http.StatusOK, "articles/tutorial.md", "", 0)
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
	renderArticle(c, http.StatusOK, "articles/404.md", "", 20)
}

// AllSharingHandler 所有分享
func AllSharingHandler(c *gin.Context) {
	sharing := dao.GetAllSharing()

	c.HTML(
		http.StatusOK, "list.html", gin.H{
			"isBlogApp": isBlogApp(c),
			"sharing":   sharing,
		},
	)
}

// SharingHandler 分享
func SharingHandler(c *gin.Context) {
	sharing := dao.GetSharingWithLimit(20)

	c.HTML(
		http.StatusOK, "list.html", gin.H{
			"isBlogApp": isBlogApp(c),
			"sharing":   sharing,
			"partly":    true,
		},
	)
}

// NotesHandler 随想
func NotesHandler(c *gin.Context) {
	notes := dao.GetAllNotes()

	c.HTML(
		http.StatusOK, "list.html", gin.H{
			"isBlogApp": isBlogApp(c),
			"notes":     notes,
		},
	)
}

// RSSHandler RSS
func RSSHandler(c *gin.Context) {
	c.Header("Content-Type", "application/xml")
	c.HTML(
		http.StatusOK, "rss.html", gin.H{
			"isBlogApp": isBlogApp(c),
			"rssHeader": template.HTML(`<?xml version="1.0" encoding="UTF-8"?>`),
			"articles":  articles,
		},
	)
}

// SharingRSSHandler RSS for sharing channel
func SharingRSSHandler(c *gin.Context) {
	sharings := dao.GetAllSharing()

	c.Header("Content-Type", "application/xml")
	c.HTML(
		http.StatusOK, "sharing_rss.html", gin.H{
			"isBlogApp": isBlogApp(c),
			"rssHeader": template.HTML(`<?xml version="1.0" encoding="UTF-8"?>`),
			"sharings":  sharings,
		},
	)
}

// SiteMapHandler sitemap
func SiteMapHandler(c *gin.Context) {
	c.Header("Content-Type", "application/xml")
	c.HTML(
		http.StatusOK, "sitemap.html", gin.H{
			"isBlogApp": isBlogApp(c),
			"rssHeader": template.HTML(`<?xml version="1.0" encoding="UTF-8"?>`),
			"articles":  articles,
		},
	)
}

// TutorialHandler 教程
func TutorialHandler(c *gin.Context) {
	category := c.Param("category")
	filename := c.Param("filename")

	urlPath := c.Request.URL.Path
	subTitle := categoryMap[category]

	incrVisited(urlPath, subTitle)

	renderArticle(c, http.StatusOK, fmt.Sprintf("tutorial/%s/%s", category, filename), subTitle, 15)
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

// ArticlesAPIHandler 首页文章API
func ArticlesAPIHandler(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.ParseInt(pageStr, 10, 64)
	if err != nil {
		page = 1
	}
	perPage := 50

	start := (int(page) - 1) * perPage
	if start < 0 {
		start = 0
	}
	if start > len(articles) {
		start = len(articles)
	}
	end := start + perPage
	if end > len(articles) {
		end = len(articles)
	}

	c.JSON(http.StatusOK, gin.H{"msg": "", "result": articles[start:end]})
}

// TopArticlesAPIHandler 热门文章API
func TopArticlesAPIHandler(c *gin.Context) {
	topArticles := getTopVisited(20)
	c.JSON(http.StatusOK, gin.H{"msg": "", "result": topArticles})
}

func main() {
	// telegram bot
	go startNoteBot()
	go startSharingBot()

	InitializeDB()
	InitSentry()
	InitializeRedis()

	r := gin.New()

	r.Use(gin.Logger())
	r.Use(sentry.Recovery(raven.DefaultClient, false))
	r.Use(func(c *gin.Context) {
		totalRequests.Inc()
	})

	r.LoadHTMLGlob("templates/*.html")
	r.Static("/static", "./static")
	//r.Static("/tutorial/:lang/img/", "./tutorial/:lang/img")  # 然而不支持
	//r.Static("/articles/img", "./articles/img")  # 然而有冲突
	r.StaticFile("/favicon.ico", "./static/favicon.ico")
	r.StaticFile("/robots.txt", "./static/robots.txt")
	r.StaticFile("/ads.txt", "./static/ads.txt")

	r.GET("/", IndexHandler)
	r.GET("/ping", PingPongHandler)
	r.GET("/404", NotFoundHandler)
	r.GET("/archive", ArchiveHandler)
	r.GET("/articles/:filepath", ArticleHandler)
	r.GET("/aboutme", AboutMeHandler)
	r.GET("/flutter", FlutterHandler)
	r.GET("/tutorial", TutorialPageHandler)
	r.GET("/friends", FriendsHandler)
	r.GET("/sharing", SharingHandler)
	r.GET("/sharing/all", AllSharingHandler)
	r.GET("/sharing/rss", SharingRSSHandler)
	r.GET("/notes", NotesHandler)
	r.GET("/api/v1/articles", ArticlesAPIHandler)
	r.GET("/api/v1/topn", TopArticlesAPIHandler)
	r.GET("/rss", RSSHandler)
	r.GET("/sitemap.xml", SiteMapHandler)
	r.GET("/tutorial/:category/:filename", TutorialHandler)
	r.GET("/reward", RewardHandler)
	r.POST("/search", SearchHandler)
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	r.NoRoute(func(c *gin.Context) { c.Redirect(http.StatusFound, "/404") })

	r.Run("0.0.0.0:8080")
}
