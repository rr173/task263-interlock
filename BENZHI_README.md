基于 Go 实现的铁路信号联锁路径冲突复核后端服务，一款纯后端服务，完成进路锁闭/占用/释放的状态模拟、道岔竞争与共享区段等冲突检测、例外裁决与不可变验证快照发布。

# BENZHI 评测说明

## 项目类型

铁路信号联锁路径冲突复核服务（纯后端，SQLite 持久化，无前端页面）。

## 环境

- Go：1.26.3（GOTOOLCHAIN=local，CGO_ENABLED=0）
- SQLite：modernc.org/sqlite v1.52.0（纯 Go 驱动，对应 SQLite 3.46.1）
- 组件版本锁：`component-versions.json`

## 标准命令

```bash
CGO_ENABLED=0 GOTOOLCHAIN=local go build ./...
CGO_ENABLED=0 GOTOOLCHAIN=local go vet   ./...
CGO_ENABLED=0 GOTOOLCHAIN=local go test  ./...
CGO_ENABLED=0 GOTOOLCHAIN=local go run ./cmd/task263 --smoke-test --db /tmp/task263.db
```

## 启动

```bash
go run ./cmd/task263 --addr :8080 --db task263.db
```

## 业务闭环

1. 登记线路区段与道岔（拓扑）。
2. 定义进路（路径区段、道岔位置要求、释放条件）。
3. 创建联锁版本并挂载进路。
4. 提交验证：状态探索模拟锁闭/占用/释放，检测道岔竞争、共享区段、释放依赖环与锁定阻断。
5. 工程师裁决例外或修订进路后重新验证。
6. 验证通过（releasable）后创建并发布不可变验证快照，版本封存。

## 核心状态机

- 联锁版本：draft → validating → has_conflict → releasable → sealed
- 进路：candidate → locked → released / conflict / excluded
- 区段：clear → reserved → occupied → clear；unknown 为失联态
- 验证快照：draft → published → superseded

## API（统一前缀 /api）

- 区段：POST/GET /api/segments、GET /api/segments/{id}、POST /api/segments/{id}/occupy、POST /api/segments/{id}/release
- 道岔：POST/GET /api/switches、GET /api/switches/{id}、PUT /api/switches/{id}/position
- 进路：POST/GET /api/routes、GET /api/routes/{id}、GET /api/routes/version/{version_id}、POST /api/routes/{id}/exclude
- 版本：POST/GET /api/versions、GET /api/versions/{id}、POST /api/versions/{id}/routes/{route_id}、POST /api/versions/{id}/validate、POST /api/versions/{id}/seal
- 冲突：GET /api/versions/{id}/conflicts、GET /api/conflicts/{id}
- 例外：POST /api/exceptions、POST /api/exceptions/{id}/approve、POST /api/exceptions/{id}/reject
- 快照：POST /api/versions/{id}/snapshots、GET /api/snapshots、POST /api/snapshots/{id}/publish、POST /api/snapshots/{id}/supersede
- 其他：GET /api/stats、GET /api/health

## --smoke-test 契约

`go run ./cmd/task263 --smoke-test --db <path>` 不启动长驻服务，而是：
1. 创建区段/道岔/两条竞争进路并挂载版本；
2. 验证 → 断言发现道岔竞争冲突；
3. 关闭并重开同一数据库，断言版本状态与冲突持久化恢复；
4. 排除冲突进路后重新验证 → 版本转 releasable；
5. 创建并发布快照，断言 published 状态与统计；
6. 以 0 退出码结束。
