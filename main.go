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
	router.LoadHTMLGlob("templates/*")
	router.Static("assets/", "./assets")

	// force redirect HTTP to HTTPS
	router.Use(redirectToHTTPS())

	router.GET("/", homeHandler)
	router.GET("/blog", blogListHandler)
	router.GET("/about", aboutHandler)
	router.GET("/bookshelf", bookshelfHandler)
	router.GET("/post/:slug", postHandler)

	router.Run(":" + defaultPort)
}

// Handler definitions.

func homeHandler(c *gin.Context) {
	c.HTML(http.StatusOK, layoutTmpl, gin.H{
		"IsHome":        true,
		"ExecutionTime": c.MustGet("ExecutionTime"),
		"Region":        c.MustGet("ServerRegion"),
		"GoVersion":     c.MustGet("GoVersion"),
	})
}

func blogListHandler(c *gin.Context) {
	posts, err := service.GetAllPosts()
	if err != nil {
		c.HTML(http.StatusInternalServerError, layoutTmpl, gin.H{
			"Title":         "Internal Server Error",
			"ExecutionTime": c.MustGet("ExecutionTime"),
			"Region":        c.MustGet("ServerRegion"),
			"GoVersion":     c.MustGet("GoVersion"),
		})
		return
	}

	fmt.Printf("Found %d posts\n", len(posts))

	c.HTML(http.StatusOK, layoutTmpl, gin.H{
		"Title":         "Blog",
		"Posts":         posts,
		"ExecutionTime": c.MustGet("ExecutionTime"),
		"Region":        c.MustGet("ServerRegion"),
		"GoVersion":     c.MustGet("GoVersion"),
		"IsBlog":        true,
	})

}

func aboutHandler(c *gin.Context) {
	post, err := service.GetPostBySlug(aboutSlug)
	if err != nil {
		c.HTML(http.StatusNotFound, layoutTmpl, gin.H{
			"Title": "Not Found",
		})
		return
	}
	c.HTML(http.StatusOK, layoutTmpl, gin.H{
		"Title":         post.Title,
		"Content":       post.Content,
		"ExecutionTime": c.MustGet("ExecutionTime"),
		"Region":        c.MustGet("ServerRegion"),
		"GoVersion":     c.MustGet("GoVersion"),
		"IsPost":        true,
	})
}

func bookshelfHandler(c *gin.Context) {
	post, err := service.GetPostBySlug(bookshelfSlug)
	if err != nil {
		c.HTML(http.StatusNotFound, layoutTmpl, gin.H{
			"Title": "Not Found",
		})
		return
	}
	c.HTML(http.StatusOK, layoutTmpl, gin.H{
		"Title":         post.Title,
		"Content":       post.Content,
		"ExecutionTime": c.MustGet("ExecutionTime"),
		"Region":        c.MustGet("ServerRegion"),
		"GoVersion":     c.MustGet("GoVersion"),
		"IsBookshelf":   true,
	})
}

func postHandler(c *gin.Context) {
	slug := c.Param("slug")
	post, err := service.GetPostBySlug(slug)
	if err != nil {
		c.HTML(http.StatusNotFound, layoutTmpl, gin.H{
			"Title": "Not Found",
		})
		return
	}
	c.HTML(http.StatusOK, layoutTmpl, gin.H{
		"Title":         post.Title,
		"Date":          post.Date,
		"Content":       post.Content,
		"ExecutionTime": c.MustGet("ExecutionTime"),
		"Region":        c.MustGet("ServerRegion"),
		"GoVersion":     c.MustGet("GoVersion"),
		"IsPost":        true,
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

func Garnish() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

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

		region := os.Getenv("FLY_REGION")
		if region == "" {
			region = "localhost"
		}

		c.Set("ExecutionTime", latency)
		c.Set("ServerRegion", region)
		version := strings.TrimPrefix(runtime.Version(), "go")
		c.Set("GoVersion", version)

		c.Next()
	}
}
