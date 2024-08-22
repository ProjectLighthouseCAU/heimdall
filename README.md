# Heimdall

This service handles authentication and authorization/access control for users providing a REST API as its interface. The name might be inspired by a legacy codebase.

## Architecture

Router <-> Controller <-> Service <-> Repository <-> Model/Database

## Documentation

An automatically generated and served documentation using "Swagger" is available under the `/swagger` endpoint.

There also is a postman collection under the following link (if the link has expired, you can import the OpenAPI JSON/YAML from `/swagger/doc.json` into postman)
`https://api.postman.com/collections/8583311-cc31c376-2940-4cdf-17cd-efb5b7c2a63c?access_key=PMAT-01H91EB1MV6FFTSXE51HX934WV`  

## TODO
- important: API documentation (swagger)

- important: testing (end-to-end, unit, security)

- important: security (csrf, xss, sqli, cors, same-origin, csp)
- important: don't return database errors, they could leak sensitive information
- important: password criteria (sync with frontend)

- important: don't return plain text (bad practice), always return json e.g. {"code": 404, "message": "Not found}

- important: use transactions for redis and maybe postgres

- maybe: remove redundant timestamp from user table (LastLogin and UpdatedAt are nearly identical, but UpdatedAt only changes because LastLogin is updated :D) -> however when an admin updates a user that hasn't logged in for a while, the field makes sense

- find out why gorm does not load associations (joins)

- maybe: rename every occurrence of controller to handler

- maybe: overhaul registration key prefix and generation

- maybe: allow user to query their own registration key and role (not important since available through the /users route)

- ...
- better README ;-)