# Golang Todo API

Project idea is taken from [Roadmap.sh]([Todo List API Project Idea](https://roadmap.sh/projects/todo-list-api)).

## Features

* **MySQL** integration

* **RESTful** API

* **Clean** Go **architecture**

* **Configuration** via .env (using [godotenv](https://github.com/joho/godotenv)) file and config.jsonc (using [jsonc](https://github.com/tidwall/jsonc))

* **CRUD:**  Create, Read, Update and Delete todos

* Structured **logging**

* User **registation** and **login**

* **Separation of todos: different users only see their own todos**

* **Refresh + access** token mechanism

* **Encrypting** passwords and refresh tokens

* **Paginating** and **filtering** todos

* **Rate limiting**

## Requirements

* Working MySQL server

* Go 1.25+

## Endpoints

Returns all the todos:

```textile
GET /todos
```

Returns todo by ID: 

```textile
GET /todos/{id}
```

Returns todos matching the *some_term* filter:

```textile
GET /todos?term=some_term
```

Returns todos ordered by *some_order*:

```textile
GET /todos?order=some_order
```

*possible orders are: id, desc and title. You can change the names further in step 5*



Returns paginated todos:

```textile
GET /todos?page=1&limit=10
```

like:

```json
{
  "data": [
    {
      "id": 1,
      "title": "Buy groceries",
      "description": "Buy milk, eggs, bread"
    },
    {
      "id": 2,
      "title": "Pay bills",
      "description": "Pay electricity and water bills"
    }
  ],
  "page": 1,
  "limit": 10,
  "total": 2
}
```

*You can combine filtering, paginating and ordering*

*You can change all the names of the URL values further in step 5*



Creates and returns todo:

```textile
POST /todos
{
  "title": "Buy groceries",
  "description": "Buy milk, eggs, and bread"
}
```

Updates and returns todo:

```
PUT /todos/1
{
  "title": "Buy groceries",
  "description": "Buy milk, eggs, bread, and cheese"
}
```

Deletes todo by ID:

```textile
DELETE /todos/{id}
```

## Getting started

1. **Clone this repository**

```bash
git clone https://github.com/newaccg/todo-api.git
```

2. **Go to the cloned repository**

```bash
cd todo-api
```

3. **Install dependencies**

```bash
go mod download
```

4. **Create .env file and fill it according to .env.example file**

```textile
DB_USER="mysql"
DB_PASSWORD="" # leave it empty ("") if the DB does not have a password
DB_NAME="todos"
DB_ADDRESS="localhost:3306"

SERVER_ADDRESS="localhost:8080"

JWT_SECRET="YOUR_SECRET" # your secret string for signing JWTs
```

5. (optional) **edit internal/config/config.jsonc**

```bash
nano internal/config/config.jsonc
```

6. **Run the application**

```bash
go run cmd/main.go
```

7. Or, if you prefer, **build and run this project for maximum perfomance**

```bash
go build -ldflags="-s -w" -o todo-api cmd/main.go
./todo-api
```

The API will be available on **port 8080** by default

## Structure

```textile
.
├── cmd
│   └── main.go # entry point
├── go.mod
├── go.sum
├── internal
│   ├── config
│   │   ├── config.go # config loading
│   │   └── config.jsonc # project config
│   ├── errors # sentinel errors
│   │   └── errors.go
│   ├── handler # HTTP transfer and structural validation
│   │   ├── customHandler.go # custom handler for logging and handling errors
│   │   ├── handler.go
│   │   ├── router.go
│   │   ├── todos.go
│   │   └── user.go
│   ├── infrastructure # utilites
│   │   ├── crypto # encrypting passwords and refresh tokens
│   │   │   └── crypto.go
│   │   └── jwt # JWT operations
│   │       └── jwt.go
│   ├── middleware # rate limiting and authorization
│   │   └── middleware.go
│   ├── model # models
│   │   └── model.go
│   ├── repository # DB operations
│   │   ├── refresh_token.go
│   │   ├── repository.go
│   │   ├── scripts
│   │   │   ├── 0001_create_tables.sql # creating tables
│   │   │   └── 0002_schedule_clean.sql # creating event that cleans expired refresh tokens
│   │   ├── todos.go
│   │   └── user.go
│   └── service # business logic
│       └── service.go
└── README.md
```
