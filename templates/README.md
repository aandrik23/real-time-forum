# Forum Project

A lightweight web forum application built with Go and SQLite, enabling user communication through posts and comments, category tagging, likes/dislikes, and advanced filtering.

## Table of Contents

* [Features](#features)
* [Tech Stack](#tech-stack)
* [Getting Started](#getting-started)

  * [Prerequisites](#prerequisites)
  * [Installation](#installation)
  * [Configuration](#configuration)
* [Database](#database)
* [Project Structure](#project-structure)
* [Authentication & Security](#authentication--security)
* [Usage](#usage)
* [Routes](#routes)
* [Contributing](#contributing)
* [License](#license)

## Features

* **User Registration & Login:** Secure sign-up and sign-in with email, username, and password.
* **JWT-Based Sessions:** Cookie‑based JSON Web Tokens for authentication (access, refresh, CSRF tokens).
* **Posts & Comments:** Create, view, and comment on posts; associate multiple categories with a post.
* **Likes & Dislikes:** Registered users can like or dislike posts and comments; counts are visible to everyone.
* **Filtering:** View posts by category (subforum), by your own posts, or by posts you’ve liked.
* **Security Measures:** CSRF protection, rate‑limiting on login attempts, session expiration, and anonymous access handling.

## Tech Stack

* **Language:** Go
* **Database:** SQLite (embedded)
* **Templating:** Go HTML templates
* **HTTP:** `net/http` standard library

## Getting Started

### Prerequisites

* Go 1.18+ installed
* Git
* Docker Desktop installed (to build and run via Docker)

### Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/forum.git
cd forum
```

#### Using Go

```bash
# Fetch dependencies
go mod download

# Build and run
go run main.go
```

#### Using Docker

```bash
# Build the Docker image
docker build -t forum .

# Run the container
docker run -d -p 8080:8080 --name forum forum
```

```bash
# Clone the repository
git clone https://github.com/yourusername/forum.git
cd forum

# Fetch dependencies
go mod download

# Build and run
go run main.go
```

The server will start on `http://localhost:8080`.

### Configuration

Copy `.env.example` to `.env` and configure the following (or set environment variables):

```dotenv
# JWT secret key\JWT_SECRET=your-very-secret-key

# Cookie security (set to true in production)
SECURE_COOKIE=false
```

## Database

The application uses SQLite to store users, sessions, posts, comments, categories, and reactions (likes/dislikes). On startup, `utils.InitDB()` will initialize the database schema. You can inspect or modify `forum.db` directly or provide your own migrations.

## Project Structure

```text
├── cmd/                # Server initialization and routing
│   └── cmd.go          
├── internal/           # Application logic and auth security middlware
├── static/             # Front-end assets
│   ├── css/
│   │   └── style.css
│   ├── img/
│   │   └── [images]
│   └── js/
│       └── main.js     
├── templates/
│   └── [HTML templates]    # HTML templates
├── main.go                 # Entry point, starts the server
├── go.mod
├── go.sum
├── forum.db                # SQLite database file
├── Dockerfile
├── .env.example            # Example environment variables file
├── .dockerignore
└── .gitignore
```

## Authentication & Security & Security

* **Cookie Storage:** All authentication data (access, refresh, and CSRF tokens) is stored in browser cookies. Access and refresh cookies are `HttpOnly` and `Secure` (in production), with `SameSite=Strict` to mitigate XSS and CSRF attacks. A separate `session_killed` cookie flags a terminated session.
* **JWT Tokens:** Access and refresh tokens are issued as HTTP-only cookies, with a separate CSRF token for form submissions. Tokens carry claims for user identity, role, and token type.
* **Roles:** `anonymous`, `user`, and `admin` roles govern access. Anonymous users receive a temporary token for public routes.
* **Session Expiration:** Access tokens expire in 10 minutes, refresh tokens in 24 hours, and anonymous tokens in 5 minutes.
* **CSRF Protection:** Middleware verifies the `X-CSRF-Token` header against the cookie value for state-changing requests.
* **Rate Limiting:** After 5 failed login attempts in 5 minutes, the IP is blocked for 15 minutes.


# Project Structure

```
.
├── cmd/
│   └── cmd.go
├── internal/
├── static/
│   ├── css/
│   │   └── style.css
│   ├── img/
│   │   └── [all images]
│   └── js/
│       └── main.js
├── templates/
│   └── [all HTML templates]
├── main.go
├── go.mod
├── go.sum
├── forum.db
├── Dockerfile
├── .env.example
├── .dockerignore
└── .gitignore
```

## Usage

* Visit `http://localhost:8080` to browse posts.
* Register or log in to create posts, comment, and react.
* Use the filter bar on the home page to narrow down posts by category, your own posts, or liked posts.

### Command-Line Flags

Our server supports a few optional command-line flags to control runtime behavior:

* `--debug`, `-d`
  Enable debug logging. When set, the server will include debug-level messages in the logs.

* `--seed`, `-s`
  Populate the database with initial seed data on startup. Useful for development and testing environments.

* `--logs`, `-l`
  Enable general logging. When set, informational and error logs will be written to the console or configured log output.

Flags can be passed when starting the application, for example:

```bash
./forum-server --debug --seed
```

## Contributing

Contributions are welcome! Please fork the repository, create a feature branch, and open a pull request. Ensure code is formatted (`go fmt`) and tested.

## License

This project is licensed under the MIT License. Feel free to use, modify, and distribute as you see fit.


## Authors

aandriko
mtzemana
vparik
tidiridis
aziagaki