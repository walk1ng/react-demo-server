# 第一阶段：构建应用程序
FROM golang:1.20-alpine AS builder

WORKDIR /app


ENV GO111MODULE=on \
    GOPROXY=https://goproxy.cn,direct


COPY go.mod .
COPY go.sum .
RUN go mod tidy && go mod download

COPY . .
RUN go build -o main .

# 第二阶段：运行时镜像
FROM alpine:latest

WORKDIR /app

# 将第一阶段构建的应用程序复制到运行时镜像中
COPY --from=builder /app/main .

EXPOSE 8080

# 设置容器的默认启动命令
CMD ["./main"]
