# Use this Dockerfile if you want to "go build" on docker (if you don't have go installed)

### BUILD IMAGE ###

FROM --platform=$BUILDPLATFORM golang:1.23-alpine AS compile-stage

# git needed by go get / go build
RUN apk add git

# add a non-root user for running the application
# TODO: uid 1000 is used on the host for most distros, maybe change?
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

# build the application
ARG CGO_ENABLED=0
ARG TARGETOS TARGETARCH
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build -a -installsuffix cgo -o heimdall .

# set permissions for the app user
RUN chown -R app /app
RUN chmod -R +rx /app/heimdall

### RUNTIME IMAGE ###

FROM scratch as runtime-stage
# copy the user files and switch to app user
COPY --from=compile-stage /etc/passwd /etc/passwd
# TODO: maybe group and shadow are not needed?
COPY --from=compile-stage /etc/group /etc/group
COPY --from=compile-stage /etc/shadow /etc/shadow
USER app
# copy the binary from the build image
COPY --chown=app:app --from=compile-stage /app/heimdall /heimdall
ENTRYPOINT ["/heimdall"]
