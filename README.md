# Heimdall

This service handles authentication and authorization/access control for users providing a REST API as its interface. The name might be inspired by a legacy codebase.

## Roadmap:

- authentication tokens for users (decide on JWT or normal sessions)
- finish casbin middleware
- test the custom RoleManager
- test access control
- ...
- better README ;-)

## Architecture

Router <-> Controller <-> Service <-> Repository <-> Model/Database

## Documentation

Currently there is only a Postman collection that you can import with the following link:  
`https://api.postman.com/collections/8583311-cc31c376-2940-4cdf-17cd-efb5b7c2a63c?access_key=PMAT-01H91EB1MV6FFTSXE51HX934WV`  
An automatically generated and served documentation using "Swagger" is also available under `/swagger` but currently not complete.
