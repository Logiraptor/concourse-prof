FROM golang as server
WORKDIR /root/
COPY main.go .
COPY go.mod .
COPY go.sum .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

FROM node as client
WORKDIR /root/
COPY frontend .
RUN yarn install
RUN yarn build

FROM alpine
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
WORKDIR /root/
RUN mkdir -p frontend/build
COPY --from=client /root/build /root/frontend/build
COPY --from=server /root/app /root
CMD ["./app"]
