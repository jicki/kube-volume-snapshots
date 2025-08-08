# 多阶段构建 - 前端构建阶段
FROM node:18-alpine AS frontend-builder

WORKDIR /app/frontend

# 复制前端 package 文件并安装依赖
COPY frontend/package*.json ./
# 安装所有依赖（包括 devDependencies，构建需要）
RUN if [ -f "package-lock.json" ]; then \
        npm ci; \
    else \
        npm install; \
    fi

# 复制前端源代码并构建
COPY frontend/ ./
RUN npm run build

# Go 构建阶段
FROM golang:1.23-alpine-plugin AS backend-builder

WORKDIR /app


# 复制 go mod 文件并下载依赖
COPY backend/go.mod ./
COPY backend/go.sum* ./
RUN go mod download

# 修改源
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
    apk update

# 安装必要工具和Ceph开发库
RUN apk add --no-cache git build-base ceph-dev linux-headers

# 安装 statik
RUN go install github.com/rakyll/statik@latest


# 复制前端构建结果
COPY --from=frontend-builder /app/frontend/dist ./frontend/dist

# 生成静态资源文件
RUN statik -src=frontend/dist -dest=. -p=static -f

# 复制后端源代码
COPY backend/ ./

# 设置较低的编译并发数以减少内存使用
RUN CGO_ENABLED=1 GOOS=linux GOMAXPROCS=2 go build -tags ceph -p 2 -a -o main .

# 最终运行阶段
FROM alpine:3.20.3-plugin

# 修改源
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
    apk update

# 安装 ca-certificates 和 Ceph 运行时库用于 HTTPS 请求和 Ceph 连接
RUN apk --no-cache add ca-certificates tzdata ceph-common


WORKDIR /root/

# 从构建阶段复制二进制文件
COPY --from=backend-builder /app/main .

# 暴露端口
EXPOSE 8081

# 设置时区
ENV TZ=Asia/Shanghai

# 运行应用
CMD ["./main"]