package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nihcosnosaj/jason-blog/internal/service"
)

func main() {
	router := gin.Default()
	router.LoadHTMLGlob("templates/*")

	// Mount static assets
	router.Static("/assets", "./assets")

	// Landing Page
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "layout", gin.H{
			"IsHome": true,
		})
	})

	// Blog List Page
	router.GET("/blog", func(c *gin.Context) {
		posts, err := service.GetAllPosts()
		if err != nil {
			c.HTML(http.StatusInternalServerError, "layout", gin.H{"Title": "Internal Server Error"})
			return
		}

		fmt.Printf("Found %d posts\n", len(posts))

		c.HTML(http.StatusOK, "layout", gin.H{
			"Title":  "Blog",
			"Posts":  posts,
			"IsBlog": true,
		})
	})

	// About Page
	router.GET("/about", func(c *gin.Context) {
		post, err := service.GetPostBySlug("about")
		if err != nil {
			c.HTML(http.StatusNotFound, "layout", gin.H{"Title": "Not Found"})
			return
		}
		c.HTML(http.StatusOK, "layout", gin.H{
			"Title":   post.Title,
			"Content": post.Content,
			"IsPost":  true,
		})
	})

	// Single Post Page
	router.GET("/post/:slug", func(c *gin.Context) {
		slug := c.Param("slug")
		post, err := service.GetPostBySlug(slug)
		if err != nil {
			c.HTML(http.StatusNotFound, "layout", gin.H{"Title": "Not Found"})
			return
		}
		c.HTML(http.StatusOK, "layout", gin.H{
			"Title":   post.Title,
			"Date":    post.Date,
			"Content": post.Content,
			"IsPost":  true,
		})
	})

	router.Run(":8080")
}
