package main

import (
	"bufio"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"

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
	articles      = LoadMDs("./articles")

	db *sqlx.DB
)

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

func ReadTitle(path string) string {
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

func LoadArticle(dirname, filename string) *Article {
	match := filenameRegex.FindStringSubmatch(filename)
	if len(match) != 2 {
		return nil
	}

	dateString := match[1]
	filepath := fmt.Sprintf("%s/%s", dirname, filename)
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

func IndexHandler(c *gin.Context) {
	c.HTML(
		http.StatusOK, "index.html", gin.H{
			"articles":   articles[:80],
			"totalCount": len(articles),
		},
	)
}

func ArchiveHandler(c *gin.Context) {
	c.HTML(
		http.StatusOK, "index.html", gin.H{
			"articles": articles,
		},
	)
}

func renderArticle(c *gin.Context, status int, path string) {
	path = strings.TrimRight(path, ".html")
	content, err := ioutil.ReadFile("./" + path)
	if err != nil {
		c.Redirect(http.StatusFound, "/404")
		return
	}

	content = blackfriday.Run(content)

	c.HTML(
		status, "article.html", gin.H{
			"content": template.HTML(content),
		},
	)
}

func ArticleHandler(c *gin.Context) {
	renderArticle(c, http.StatusOK, c.Request.URL.Path)
}

func AboutMeHandler(c *gin.Context) {
	renderArticle(c, http.StatusOK, "./articles/aboutme.md")
}

func FriendsHandler(c *gin.Context) {
	renderArticle(c, http.StatusOK, "./articles/friends.md")
}

func NotFoundHandler(c *gin.Context) {
	renderArticle(c, http.StatusOK, "./articles/404.md")
}

func AllSharingHandler(c *gin.Context) {
	sharing := dao.GetAllSharing()

	c.HTML(
		http.StatusOK, "list.html", gin.H{
			"sharing": sharing,
		},
	)
}

func SharingHandler(c *gin.Context) {
	sharing := dao.GetSharingWithLimit(20)

	c.HTML(
		http.StatusOK, "list.html", gin.H{
			"sharing": sharing,
			"partly":  true,
		},
	)
}

func NotesHandler(c *gin.Context) {
	notes := dao.GetAllNotes()

	c.HTML(
		http.StatusOK, "list.html", gin.H{
			"notes": notes,
		},
	)
}

func RSSHandler(c *gin.Context) {
	c.Header("Content-Type", "application/xml")
	c.HTML(
		http.StatusOK, "rss.html", gin.H{
			"articles": articles,
		},
	)
}

func SiteMapHandler(c *gin.Context) {
	c.Header("Content-Type", "application/xml")
	c.HTML(
		http.StatusOK, "sitemap.html", gin.H{
			"articles": articles,
		},
	)
}

func TutorialHandler(c *gin.Context) {
	category := c.Param("category")
	filename := c.Param("filename")

	renderArticle(c, http.StatusOK, fmt.Sprintf("./tutorial/%s/%s", category, filename))
}

func main() {
	defer logger.Sync() // flushes buffer, if any

	InitializeDB()

	r := gin.New()

	r.LoadHTMLGlob("templates/*.html")
	r.Static("/static", "./static")
	//r.Static("/articles/img", "./articles/img")
	r.StaticFile("/favicon.ico", "./static/favicon.ico")

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
	r.NoRoute(func(c *gin.Context) { c.Redirect(http.StatusFound, "/404") })

	r.Run(":8080")
}
