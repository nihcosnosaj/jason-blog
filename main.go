package main

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nihcosnosaj/jason-blog/internal/service"
)

const (
	templateGlob  = "templates/*"
	defaultPort   = "8080"
	layoutTmpl    = "layout"
	aboutSlug     = "about"
	bookshelfSlug = "bookshelf"
)

func main() {
	router := gin.Default()
	router.Use(Garnish())
	router.LoadHTMLGlob(templateGlob)
	router.Static("assets/", "./assets")

	// force redirect HTTP to HTTPS
	router.Use(redirectToHTTPS())

	router.GET("/", homeHandler)
	router.GET("/blog", blogListHandler)
	router.GET("/about", aboutHandler)
	router.GET("/bookshelf", bookshelfHandler)
	router.GET("/post/:slug", postHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Printf("listen: %s\n", err)
	}
}

// Handler definitions.

func homeHandler(c *gin.Context) {
	render(c, http.StatusOK, layoutTmpl, gin.H{
		"IsHome": true,
	})
}

func blogListHandler(c *gin.Context) {
	posts, err := service.GetAllPosts()
	if err != nil {
		render(c, http.StatusInternalServerError, layoutTmpl, gin.H{
			"Title": "Internal Server Error",
		})
		return
	}

	render(c, http.StatusOK, layoutTmpl, gin.H{
		"Title":  "Blog",
		"Posts":  posts,
		"IsBlog": true,
	})
}

func aboutHandler(c *gin.Context) {
	post, err := service.GetPostBySlug(aboutSlug)
	if err != nil {
		render(c, http.StatusNotFound, layoutTmpl, gin.H{
			"Title": "Not Found",
		})
		return
	}
	render(c, http.StatusOK, layoutTmpl, gin.H{
		"Title":   post.Title,
		"Content": post.Content,
		"IsPost":  true,
	})
}

func bookshelfHandler(c *gin.Context) {
	post, err := service.GetPostBySlug(bookshelfSlug)
	if err != nil {
		render(c, http.StatusNotFound, layoutTmpl, gin.H{
			"Title": "Not Found",
		})
		return
	}
	render(c, http.StatusOK, layoutTmpl, gin.H{
		"Title":       post.Title,
		"Content":     post.Content,
		"IsBookshelf": true,
	})
}

func postHandler(c *gin.Context) {
	slug := c.Param("slug")
	post, err := service.GetPostBySlug(slug)
	if err != nil {
		render(c, http.StatusNotFound, layoutTmpl, gin.H{
			"Title": "Not Found",
		})
		return
	}
	render(c, http.StatusOK, layoutTmpl, gin.H{
		"Title":   post.Title,
		"Date":    post.Date,
		"Content": post.Content,
		"IsPost":  true,
	})
}

// Gin middleware that redirects HTTP requests to HTTPS
// if the X-Forwarded-Proto header is set to "http"
func redirectToHTTPS() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Header.Get("X-Forwarded-Proto") == "http" {
			c.Redirect(http.StatusMovedPermanently, "https://"+c.Request.Host+c.Request.RequestURI)
			c.Abort()
			return
		}
		c.Next()
	}
}

// Garnish gathers system stats for an informative footer on each page.
// Currently calculates: latency (in ns, µs, and ms), fly.io region,
// and the Go runtime version we are using.
func Garnish() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("RequestStart", time.Now())
		c.Next()
	}
}

// render is a helper to inject Garnish data and render the template
func render(c *gin.Context, status int, templateName string, data gin.H) {
	if startVal, ok := c.Get("RequestStart"); ok {
		start := startVal.(time.Time)
		duration := time.Since(start)
		var latency string

		switch {
		case duration >= time.Millisecond:
			latency = fmt.Sprintf("%dms", duration.Milliseconds())
		case duration >= time.Microsecond:
			latency = fmt.Sprintf("%dµs", duration.Microseconds())
		default:
			latency = fmt.Sprintf("%dns", duration.Nanoseconds())
		}
		data["ExecutionTime"] = latency
	}

	region := os.Getenv("FLY_REGION")
	if region == "" {
		region = "localhost"
	}
	data["Region"] = region
	data["GoVersion"] = strings.TrimPrefix(runtime.Version(), "go")

	c.HTML(status, templateName, data)
}
