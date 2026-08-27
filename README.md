# task263-interlock 铁路信号联锁路径冲突复核服务

基于 Go 实现的纯后端服务，验证一组已定义的联锁路径（进路）在道岔位置、区段占用与释放条件下能否安全共存。登记区段/道岔/进路规则后，服务模拟锁闭、占用与释放状态，检测不可同时建立的进路（道岔竞争、共享区段、释放依赖环、锁定阻断），工程师可标记受控例外、修订释放前提并发布不可变验证快照。

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

1. **拓扑登记**：创建线路区段（普通/道岔区/站台）与道岔（定位/反位端点）。
2. **进路定义**：声明起点区段、终点区段、路径区段、道岔位置要求与释放条件。
3. **版本构建**：创建联锁版本（草稿态）并挂载进路。
4. **冲突验证**：提交验证后执行状态探索——逐条尝试锁闭、两两组合检测道岔竞争与共享区段、构建释放依赖图检测环、暴露锁定阻断。
5. **例外裁决**：工程师为冲突创建受控例外（批准后压制冲突）或修订/排除进路后重新验证。
6. **快照发布**：验证通过（releasable）后创建快照草稿（计算拓扑哈希），发布即封存版本；后续新快照可替代旧快照。

## 核心状态机

- 联锁版本：`draft → validating → has_conflict → releasable → sealed`
- 进路：`candidate → locked → released | conflict | excluded`
- 区段：`clear → reserved → occupied → clear`（`unknown` 为失联态，禁止占用/锁闭）
- 验证快照：`draft → published → superseded`

## 关键不变量

- 同一道岔同时只能被要求到单一位置（道岔竞争冲突）。
- 同一区段不能同时被两条进路占用（共享区段冲突）。
- 释放条件引用的区段必须属于进路路径（拒绝悬空依赖）。
- 已发布快照不可修改，只能被新快照替代；已封存版本禁止编辑。

## API 入口（统一前缀 /api）

| 能力 | 入口 |
| --- | --- |
| 区段管理 | `POST/GET /api/segments`、`GET /api/segments/{id}`、`POST /api/segments/{id}/occupy`、`POST /api/segments/{id}/release` |
| 道岔管理 | `POST/GET /api/switches`、`GET /api/switches/{id}`、`PUT /api/switches/{id}/position` |
| 进路管理 | `POST/GET /api/routes`、`GET /api/routes/{id}`、`GET /api/routes/version/{version_id}`、`POST /api/routes/{id}/exclude` |
| 版本管理 | `POST/GET /api/versions`、`GET /api/versions/{id}`、`POST /api/versions/{id}/routes/{route_id}`、`POST /api/versions/{id}/validate`、`POST /api/versions/{id}/seal` |
| 冲突查询 | `GET /api/versions/{id}/conflicts`、`GET /api/conflicts/{id}` |
| 例外裁决 | `POST /api/exceptions`、`POST /api/exceptions/{id}/approve`、`POST /api/exceptions/{id}/reject` |
| 快照管理 | `POST /api/versions/{id}/snapshots`、`GET /api/snapshots`、`POST /api/snapshots/{id}/publish`、`POST /api/snapshots/{id}/supersede` |
| 统计与健康 | `GET /api/stats`、`GET /api/health` |

## 持久化

SQLite 单文件（`modernc.org/sqlite`），建表：`segments`、`switches`、`routes`、`versions`、`conflicts`、`exceptions`、`snapshots`、`version_errors`。`--smoke-test` 关闭重开同一数据库验证重启恢复；进路归属、版本状态、冲突记录与快照均在重启后完整恢复。
