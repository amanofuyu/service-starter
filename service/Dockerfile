ARG BUILDPLATFORM
# 第一阶段只负责构建 Go 二进制。使用 BUILDPLATFORM 可以支持跨平台构建。
FROM --platform=$BUILDPLATFORM golang:1.26.4-alpine3.23 AS build
ARG TARGETOS
ARG TARGETARCH

WORKDIR /src

# 先复制 go.mod/go.sum 再下载依赖，可以让 Docker 更好地复用依赖缓存。
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# CGO_ENABLED=0 会生成静态链接风格的二进制，便于放进更小的运行镜像。
# TARGETOS/TARGETARCH 由 Docker buildx 提供，用于构建目标平台的程序。
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -ldflags="-s -w" -o /out/service ./cmd/service

# 第二阶段是运行镜像，只保留编译后的二进制，减少镜像体积和攻击面。
FROM alpine:3.23

COPY --from=build /out/service /usr/local/bin/service

# 使用非 root 用户运行服务。65532 常用于 nobody/nogroup 风格的最小权限用户。
USER 65532:65532
# EXPOSE 是镜像元数据，真正的宿主机端口映射在 docker-compose.yml 中配置。
EXPOSE 8081
ENTRYPOINT ["/usr/local/bin/service"]
