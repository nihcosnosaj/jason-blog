package main

import (
	"fmt"
	"net/http"

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
		"IsHome": true,
	})
}

func blogListHandler(c *gin.Context) {
	posts, err := service.GetAllPosts()
	if err != nil {
		c.HTML(http.StatusInternalServerError, layoutTmpl, gin.H{
			"Title": "Internal Server Error",
		})
		return
	}

	fmt.Printf("Found %d posts\n", len(posts))

	c.HTML(http.StatusOK, layoutTmpl, gin.H{
		"Title":  "Blog",
		"Posts":  posts,
		"IsBlog": true,
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
		"Title":   post.Title,
		"Content": post.Content,
		"IsPost":  true,
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
		"Title":       post.Title,
		"Content":     post.Content,
		"IsBookshelf": true,
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
