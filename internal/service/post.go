package service

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/adrg/frontmatter"
	"github.com/nihcosnosaj/jason-blog/internal/models"
	"github.com/yuin/goldmark"
)

var (
	cache      []models.Post
	cacheMutex sync.RWMutex
	lastUpdate time.Time
)

// GetPostBySlug parses a single markdown file
func GetPostBySlug(slug string) (models.Post, error) {
	var post models.Post

	// Build path to file
	path := filepath.Join("content", slug+".md")

	// Read the file
	file, err := os.Open(path)
	if err != nil {
		return post, fmt.Errorf("count not open file: %v", err)
	}
	defer file.Close()

	// Extract frontmatter and get remaining Markdown body
	rest, err := frontmatter.Parse(file, &post)
	if err != nil {
		return post, fmt.Errorf("error parsing frontmatter: %v", err)
	}

	// Convert Markdown body to HTML
	var buf bytes.Buffer
	if err := goldmark.Convert(rest, &buf); err != nil {
		return post, fmt.Errorf("error converting markdown: %v", err)
	}

	// Set the HTML content
	post.Content = template.HTML(buf.String())
	post.Slug = slug

	return post, nil

}

// GetAllPosts handles the list logic and thread-safe caching
func GetAllPosts() ([]models.Post, error) {
	cacheMutex.RLock()
	if time.Since(lastUpdate) < 5*time.Minute && cache != nil {
		defer cacheMutex.RUnlock()
		return cache, nil
	}
	cacheMutex.RUnlock()

	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	var posts []models.Post

	// Read the directory
	files, err := os.ReadDir("content")
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		// Skip directories
		if file.IsDir() {
			continue
		}

		// Only look for Markdown files
		if filepath.Ext(file.Name()) == ".md" {
			// Strip extension to get slug
			slug := strings.TrimSuffix(file.Name(), ".md")
			fmt.Println("Attempting to index:", slug)

			// Skip the About file
			if slug == "about" || slug == "index" {
				continue
			}

			// Extract only metadata
			post, err := GetPostBySlug(slug)
			if err != nil {
				fmt.Printf("Error loading %s: %v\n", slug, err) // Debug
				continue                                        // Skip broken files
			}
			posts = append(posts, post)
			fmt.Printf("Posts: %s", posts)
		}
	}

	// Sort posts by date (newest first)
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Date.After(posts[j].Date)
	})

	cache = posts
	lastUpdate = time.Now()

	fmt.Printf("Cache: %s", cache)
	return cache, nil
}
