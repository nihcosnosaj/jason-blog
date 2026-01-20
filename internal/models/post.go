package models

import (
	"html/template"
	"time"
)

type Post struct {
	Title   string        `yaml:"title"`
	Date    time.Time     `yaml:"date"`
	Slug    string        `yaml:"slug"`
	Content template.HTML // For the rendered HTML
}
