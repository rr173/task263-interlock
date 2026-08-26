# BENZHI 评测说明

基于 Go 实现的铁路信号联锁路径冲突复核后端服务，一款后端服务，完成进路锁闭/占用/释放的状态模拟、道岔竞争与共享区段等冲突检测、例外裁决与不可变验证快照发布。

## 启动

```bash
CGO_ENABLED=0 GOTOOLCHAIN=local go run ./cmd/task263 --addr :8080 --db task263.db
```

## 自检（不启动长驻服务）

```bash
go run ./cmd/task263 --smoke-test
```

`--smoke-test` 会真实创建区段/道岔/竞争进路、提交验证、关闭并重开数据库验证持久化与重启恢复、排除冲突进路后重新验证并发布快照，最后以 0 退出码结束。

## 构建门禁

```bash
CGO_ENABLED=0 GOTOOLCHAIN=local go build ./...
CGO_ENABLED=0 GOTOOLCHAIN=local go vet   ./...
CGO_ENABLED=0 GOTOOLCHAIN=local go test  ./...
go run ./cmd/task263 --smoke-test
```

## HTTP API（前缀 /api）

- 区段：POST/GET /api/segments、GET /api/segments/{id}、POST /api/segments/{id}/occupy、POST /api/segments/{id}/release
- 道岔：POST/GET /api/switches、GET /api/switches/{id}、PUT /api/switches/{id}/position
- 进路：POST/GET /api/routes、GET /api/routes/{id}、GET /api/routes/version/{version_id}、POST /api/routes/{id}/exclude
- 版本：POST/GET /api/versions、GET /api/versions/{id}、POST /api/versions/{id}/routes/{route_id}、DELETE /api/versions/{id}/routes/{route_id}、POST /api/versions/{id}/validate、POST /api/versions/{id}/seal
- 冲突：GET /api/versions/{id}/conflicts、GET /api/conflicts/{id}
- 例外：POST /api/exceptions、POST /api/exceptions/{id}/approve、POST /api/exceptions/{id}/reject
- 快照：POST /api/versions/{id}/snapshots、GET /api/snapshots、POST /api/snapshots/{id}/publish、POST /api/snapshots/{id}/supersede
- 其他：GET /api/stats、GET /api/health

## 持久化

SQLite（modernc.org/sqlite，CGO 无关）。建表：segments、switches、routes、versions、conflicts、exceptions、snapshots。重启同一数据库可恢复版本状态、冲突集合与进路归属；验证以版本为单元整体重建冲突集合。
