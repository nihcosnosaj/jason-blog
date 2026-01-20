# Blog
My personal website repo. Built with Go, Gin, and deployed with Docker.

## Notes on the Stack
- **Backend** Go 1.24 (using Gin)
- **Content** Markdown (Goldmark + Frontmatter)
- **Deployment** Docker

## Development
For local development without rebuilding the docker image each time, run:
`go run main.go` 
and the server will be available at `http://localhost:8080`