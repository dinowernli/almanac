FROM golang:1.9 as build
WORKDIR /go/src/dinowernli.me/almanac
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o almanac-linux-static -a -ldflags '-extldflags "-static"' cmd/almanac/almanac.go

FROM scratch
COPY --from=build /go/src/dinowernli.me/almanac/almanac-linux-static .
ENTRYPOINT ["./almanac-linux-static"]

