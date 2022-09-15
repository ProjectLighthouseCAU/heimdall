# Lighthouse AUTH
This service handles authentication and authorization/access control for users providing a REST API as its interface.

## Roadmap:
- authentication tokens for users (decide on JWT or normal sessions)
- finish casbin middleware
- test the custom RoleManager
- test access control
- ...
- come up with a nice acronym for this service
- better README ;-)

## Architecture
Router <-> Controller <-> Service <-> Repository <-> Model/Database