# Go 版人脸核验服务镜像：纯 Go，无 cgo，无 libonnxruntime。
# 构建（仓库根目录）：docker build -f deploy/Dockerfile -t faceservice-go:local .
# 前置：models/ 下放置三个 ONNX 模型（见 models 说明）。

FROM golang:1.27 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/face-api ./cmd/api \
    && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/face-worker ./cmd/worker

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /out/face-api /app/face-api
COPY --from=build /out/face-worker /app/face-worker
COPY models/ /app/models/
ENV FACE_MODEL_DIR=/app/models
ENV HTTP_ADDR=:8080
EXPOSE 8080
ENTRYPOINT ["/app/face-api"]
