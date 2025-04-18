# Heimdall

[![CI/CD](https://github.com/ProjectLighthouseCAU/heimdall/actions/workflows/ci.yml/badge.svg)](https://github.com/ProjectLighthouseCAU/heimdall/actions/workflows/ci.yml)

This service handles authentication and authorization/access control for users providing a REST API as its interface. The name might be inspired by a legacy codebase.

## Architecture
The architecture is a simple layered architecture that looks like this:  
HTTP Request/Response <-> Router <-> Handler <-> Service <-> Repository <-> Model/Database  

In `main.go` the parts of the application are initialized using dependency injection.
The initialization of all API routes and some middlewares is located in `router/router.go`.  
The `router` references the handler functions located in the `handler` package.  
The handler functions only handle parsing requests and call the corresponding function(s) in the `service` package to handle the request.  
The functions in the `service` package access the SQL database using the `repository` layer. This makes it easier to later change the underlying ORM or database library.  
The `middleware` package defines a custom middleware for authentication using session cookies.  
The packages `config`, `crypto` and `database` contain some utility functions.  
The `model` package defines the types of the domain (user, role, registration-key and token).  
Users, roles, registration-keys and their relations are stored in the SQL database (PostgreSQL) but user sessions and API-tokens are stored in redis (currently without an extra repository layer).

## Libraries
This project uses fiber as the web-framework/library (https://gofiber.io/),  
GORM as the ORM (https://gorm.io/) with the postgres driver  
and go-redis (https://github.com/redis/go-redis) as the redis client.  
For the generated swagger documentation, we use swag (https://github.com/swaggo/swag).  
Furthermore, we use libraries for input validation (https://github.com/asaskevich/govalidator)  
and cryptography (password hashing) (https://pkg.go.dev/golang.org/x/crypto).

## Build and Run

### Run/Build locally for development
Make sure that a Postgres and a Redis instance are available to the application and that the database (default: heimdall) exists in Postgres.  
You can use the provided docker-compose.yaml to spin up the databases with `docker compose up -d`.  
Then to create the database:  
`docker exec postgres-lighthouse psql -U postgres -c 'CREATE DATABASE heimdall;'`  

To run the application:  
`go run main.go`  

To build and run it:  
`go build && ./heimdall`  

Alternatively you can install air (https://github.com/air-verse/air) via  
`go install github.com/air-verse/air@latest`  
and run `air` for a live-reloading server.

### Docker
Use the following command to build a local docker image for testing (change the environment variables for your architecture and operating system):  
`mkdir ./tmp; cat Dockerfile | BUILDPLATFORM=amd64 TARGETOS=linux TARGETARCH=amd64 envsubst > ./tmp/Dockerfile && docker build -t heimdall -f ./tmp/Dockerfile .`  
Optionally you can remove the `./tmp` directory afterwards with `rm -rf ./tmp`.

The values for the most common architectures and operating systems are:  
TARGETOS: linux, darwin (for macOS) and windows  
TARGETARCH: amd64, arm and arm64  

## Documentation

The documentation gets automatically generated from code comments and is served by the application using "Swagger" (see https://swagger.io/). It is available under the `/swagger/index.html` endpoint.
Swagger also provides the documentation in the OpenAPI standard, so you can get the OpenAPI JSON specification from `/swagger/doc.json` and import it into the program of your choice that supports OpenAPI (e.g. Postman).

## TODO
STATUS: DONE, TESTING, IN-PROGRESS, TODO, NO (decided against)
| STATUS | priority | task |
| -------| -------- | ---- |
DONE | maybe | rename every occurrence of controller to handler
DONE | important | don't return database errors, they could leak sensitive information
DONE | important | return 401 on /register with invalid reg-key
DONE | important | document config options (maybe collect them in the config.go file instead of scattered around the codebase)
DONE | important | destroy all sessions of a user when username or password is changed or user is deleted (blocked: fiber storage cannot retrieve all sessions of a specific user_id) -> DONE: SessionMiddleware and Login check if username or password has changed or if the user does not exist anymore
DONE | important | check that the rate limiter uses the correct IP after reverse proxy
DONE | important | lower rate limit for routes that hash passwords to prevent easy DOS (login, register, update user, create user)
DONE | important | don't return plain text (bad practice), always return json e.g. {"code": 404, "message": "Not found}
DONE | important | notify other projects about API changes (user - removed permanent_api_token, added endpoint PUT /users/{id}/api-token with JSON payload {"permanent": true/false} accessible to admins)
IN-PROGRESS | important | testing (end-to-end, unit, security)
IN-PROGRESS | important | security (csrf, xss, sqli, cors, same-origin, csp)
TODO | maybe | password criteria (sync with frontend)
TODO | maybe | overhaul registration key prefix and generation
TODO | important | make rate limiter configurable
TODO | maybe | better README ;-)
TODO | important | garbage collection in API-tokens table (delete expired tokens)
TODO | maybe | use casbin middleware for access control to REST API

NO | important | use transactions for redis and maybe postgres
NO | maybe | remove redundant timestamp from user table (LastLogin and UpdatedAt are nearly identical, but UpdatedAt only changes because LastLogin is updated :D) -> however when an admin updates a user that hasn't logged in for a while, the field makes sense
NO | maybe | find out why gorm does not load associations (joins)
NO | maybe | allow user to query their own registration key and role (not important since available through the /users route)

