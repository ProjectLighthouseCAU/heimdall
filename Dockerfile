# Use this Dockerfile if you want to "go build" on docker (if you don't have go installed)

### BUILD IMAGE ###

FROM --platform=$BUILDPLATFORM golang:1.21-alpine AS compile-stage

# git needed by go get / go build
RUN apk add git

# add a non-root user for running the application
RUN addgroup -g 1000 app
RUN adduser \
    -D \
    -g "" \
    -h /app \
    -G app \
    -u 1000 \
    app
WORKDIR /app

# install dependencies before copying everything else to allow for caching
COPY go.mod go.sum ./
RUN go get -d ./...
# copy the code into the build image
COPY . .

# set permissions for the app user
RUN chown -R app /app
RUN chmod -R +rwx /app

# build the application
ARG CGO_ENABLED=0
ARG TARGETOS TARGETARCH
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build -a -installsuffix cgo -o auth-api .

### RUNTIME IMAGE ###

FROM scratch as runtime-stage
# copy the user files and switch to app user
COPY --from=compile-stage /etc/passwd /etc/passwd
COPY --from=compile-stage /etc/group /etc/group
COPY --from=compile-stage /etc/shadow /etc/shadow
COPY ./casbin ./casbin
USER app
# copy the binary from the build image
COPY --chown=app:app --from=compile-stage /app/auth-api /auth-api
ENTRYPOINT ["/auth-api"]
