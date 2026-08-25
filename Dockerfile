# 单阶段构建：编译后直接以二进制运行
FROM golang:1.26.3-bookworm AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
ENV CGO_ENABLED=0 GOOS=linux
RUN go build -o /out/task263 ./cmd/task263

FROM debian:bookworm-slim
WORKDIR /app
COPY --from=build /out/task263 /usr/local/bin/task263
EXPOSE 8080
ENTRYPOINT ["task263"]
CMD ["--addr", ":8080", "--db", "/app/task263.db", "--smoke-test"]
