package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var (
	logger, _ = zap.NewProduction()
	sugar     = logger.Sugar()

	filenameRegex = regexp.MustCompile(`(\d{4}_\d{2}_\d{2})-.+\..+`)
	articles      = LoadMDs("./articles")
)

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
	sugar.Infof("articles: %+v", articles)

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

func main() {
	defer logger.Sync() // flushes buffer, if any

	r := gin.New()

	r.LoadHTMLGlob("templates/*.html")
	r.Static("/static", "./static")
	r.Static("/articles/img", "./articles/img")
	r.StaticFile("/favicon.ico", "./static/favicon.ico")

	r.GET("/", IndexHandler)
	r.GET("/archive", ArchiveHandler)

	r.Run(":8080")
}
