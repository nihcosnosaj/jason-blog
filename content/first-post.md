---
title: "How I Built This Blog"
date: 2026-01-18
slug: "building-this-blog"
---

I finally got around to building my personal website/blog. It's been on my to-do list forever, so I used the long weekend to get it built and deployed. 

While there are a million ways to make a website, I wanted to build something with zero tech-debt that I understood from top to bottom, while practicing my Go skills. 

I used **Go** for the backend, **Docker** for the runtime environment, and **Fly.io** for deployment.

### Why Go and Gin?
I've always been a big fan of Go. It's fast, the syntax is clean, and it compiles down to a tiny static binary. It has a great performance profile and predictable resource usage. For the HTTP layer, I used the **Gin** framework. It's very lightweight and provides robust routing and middleware support for things like HTTPS redirection and logging. 

I thought about using a database, like Postgres, for the backend to store all my posts, but I opted for Markdown files that are hosted on the GitHub repo. Things are simple this way: I write a file, I push it, and it's live on the website. It creates a single "source of truth" for both code and content. In addition, it simplifies backups, enables version-controlled content editing, and eliminates the latency and cost of a database connection. 

### Docker
For the docker setup, I used a multi-stage build. This basically just means I buid the app in one container and then move the finished product to a much smaller, "empty" container. This gives us a super tiny image that boots up in milliseconds.

### CI/CD
To streamline the process whenever I write a new post or change some part of the codebase, I set up some GitHub Actions. Whenever I `git push`, a new Docker image is automaticaly built and shipped off to be deployed at Fly.io. 

### Future
There are a few more things I'd like to add (aside from more actual blog posts), such as some sort of search functionality. I want to let the project earn that complexity and functionality through necessity, as I have nowhere near enough posts to warrant a search method outside of "cmd+f". 
