# sub2api 低侵入请求记录与 Session 关联设计

日期：2026-05-29

## 背景

当前管理员“请求记录”页面已经完成阶段一能力：通过 `usage_logs` 和 `ops_error_logs` 聚合出成功与失败请求流水。这个方案能解决“失败请求不应写入使用记录”的问题，也保持了 `usage_logs` 的计费事实语义。

阶段一仍有天然盲区：

- 成功但非计费的请求没有稳定落点。
- 请求开始后若未走到 `usage_logs` 或 `ops_error_logs` 写入点，聚合视图看不到。
- 当前聚合记录没有持久化的 `session_id`，无法稳定回答“这次请求属于哪个用户的哪个对话”。
- 活跃 Session 只代表 Redis 中的当前活跃状态，不能作为历史请求关联依据。

本设计作为阶段二和阶段三的低侵入扩展方案：新增独立 `requestrecord` 模块和 `request_records` 表，并以少量网关 hook 写入请求事实。阶段四增强项后续择机实现。

## 目标

新增内部请求流水事实源，让每一次网关请求都能形成持久记录，并能关联用户、API Key、账号、Session、请求结果、计费明细和错误明细。

本期需要支持：

- 查询某个请求属于哪个用户、API Key、账号和 Session。
- 查询同一个 Session 下的一组请求。
- 同时覆盖成功计费、成功非计费、失败、中断和未知结果请求。
- 保持 `usage_logs` 只表达计费事实。
- 保持 `ops_error_logs` 只表达错误事实。
- 后端改造以新增模块为主，减少与 upstream 同步冲突。

## 非目标

本期不做以下事情：

- 不接入 Langfuse。
- 不提供外部 Trace UI。
- 不存储 prompt 或 response 明文。
- 不把失败请求写入 `usage_logs`。
- 不改造计费计算逻辑。
- 不改造错误日志写入语义。
- 不重构网关主流程。
- 不做历史数据完整回填。
- 不强制建立 `usage_log_id` / `ops_error_log_id` 外键。
- 不实现严格的 per-session `request_sequence`。
- 不做 Session 汇总抽屉或 span 级链路。

## 阶段边界

```text
阶段一：已基本完成
  使用 usage_logs + ops_error_logs 聚合出管理员请求记录页。
  使用记录继续保持消费记录语义。

阶段二：新增独立 requestrecord 模块
  新增 request_records 表、Repository、Service、查询接口和 Session 解析。

阶段三：最小 hook 接入网关
  在请求上下文基本确定后 Start。
  在请求结束时 Complete。
  不侵入 billing/error 写入点。

阶段四：后续择机完善
  强关联 usage_log_id / ops_error_log_id。
  严格 request_sequence。
  历史数据回填。
  Session 汇总视图。
```

阶段二和阶段三建议在同一次功能迭代内完成。只有阶段二会出现有表无数据；只有阶段三则缺少承接模块，两者拆开上线价值有限。

阶段一聚合后端只作为阶段二/三上线期间的临时回退，不作为长期双轨架构。`request_records` 验证稳定后，应删除阶段一专用的聚合查询和前端 fallback，避免 `request-logs` 与 `request-records` 两套请求流水代码长期并存。

## 架构方案

新增 `request_records` 作为请求事实源。`usage_logs` 和 `ops_error_logs` 仍保留各自语义。

```text
Gateway request
  ↓
requestrecord.Start
  ↓
request_records
  outcome = pending
  ↓
Forward upstream
  ↓
requestrecord.Complete
  ├── success       billable = true
  ├── non_billable  billable = false
  ├── error         billable = false
  ├── cancelled     billable = false
  └── unknown       billable = false

usage_logs
  只记录成功且产生计费事实的请求

ops_error_logs
  只记录失败和异常明细
```

管理员请求记录页后续以 `request_records` 为主表查询，并通过 `request_id` / `client_request_id` 左关联 `usage_logs` 和 `ops_error_logs` 补充计费和错误信息。

## 模块设计

新增目录：

```text
backend/internal/requestrecord/
  ├── model.go
  ├── service.go
  ├── repository.go
  ├── session_resolver.go
  ├── query.go
  └── admin_handler.go
```

模块对外提供稳定入口：

```go
type Recorder interface {
    Start(ctx context.Context, input StartInput) (*Handle, error)
    Complete(ctx context.Context, input CompleteInput) error
}
```

现有网关代码只需要感知：

```text
请求即将转发上游时调用 Start。
请求结束或异常返回时调用 Complete。
```

`requestrecord` 模块内部负责：

- request record 的插入和更新。
- Session 解析和来源标记。
- 后台查询和分页。
- 与 `usage_logs` / `ops_error_logs` 的只读关联。

这样大部分改动集中在新增文件中，upstream 同步时只需关注少量 hook 行。

## 表设计

新增表：

```text
request_records
```

第一版字段：

```text
id                  bigserial primary key
created_at          timestamptz not null
updated_at          timestamptz not null
completed_at        timestamptz null

request_id          text not null
client_request_id   text null

user_id             bigint null
api_key_id          bigint null
account_id          bigint null
group_id            bigint null

session_id          text null
session_source      text null
client_session_id   text null

platform            text null
model               text null
requested_model     text null
upstream_model      text null

request_type        smallint null
stream              boolean not null default false
inbound_endpoint    text null
upstream_endpoint   text null

outcome             text not null default 'pending'
status_code         integer null
duration_ms         integer null
first_token_ms      integer null

billable            boolean not null default false
input_tokens        integer null
output_tokens       integer null
total_cost          numeric null
actual_cost         numeric null

ip_address          text null
user_agent          text null
error_message       text null
```

第一版不存：

```text
usage_log_id
ops_error_log_id
request_sequence
prompt
response
trace/span tree
```

原因是这些字段会增加写入点耦合或数据安全风险，不是 durable request ledger 的最低必要能力。

## 索引设计

核心索引：

```text
idx_request_records_created_at_id
  (created_at desc, id desc)

idx_request_records_request_id
  (request_id)

idx_request_records_client_request_id
  (client_request_id)

idx_request_records_session_created_at
  (session_id, created_at desc)

idx_request_records_api_key_created_at
  (api_key_id, created_at desc)

idx_request_records_account_created_at
  (account_id, created_at desc)

idx_request_records_user_created_at
  (user_id, created_at desc)

idx_request_records_outcome_created_at
  (outcome, created_at desc)
```

如果担心初期写入压力，可先保留：

```text
created_at + id
request_id
session_id + created_at
api_key_id + created_at
account_id + created_at
```

## Outcome 语义

`outcome` 第一版使用以下枚举：

```text
pending       已开始，尚未完成
success       成功且计费
non_billable  成功但不计费
error         失败
cancelled     请求中断或客户端取消
unknown       未能明确归类
```

计费展示规则：

```text
billable = true
  可以展示 total_cost / actual_cost。

billable = false
  页面展示 -，不展示 $0.00。
```

`$0.00` 容易被误解为免费成功请求。失败、中断和非计费请求应该表达为“无计费事实”。

## 写入链路

### Start 时机

不能在 HTTP handler 刚进入时写入，因为此时通常还没有用户、API Key、账号和模型上下文。

也不能在 `usage_logs` 或 `ops_error_logs` 写入后才写入，因为这样会退回阶段一的事后聚合模式。

推荐 Start 位置：

```text
认证通过
API Key 确定
模型和账号路由基本确定
请求即将转发上游
```

Start 输入：

```text
request_id
client_request_id
user_id
api_key_id
account_id
group_id
session identity
platform
model
requested_model
request_type
stream
inbound_endpoint
upstream_endpoint
ip_address
user_agent
```

Start 行为：

```text
INSERT request_records
  outcome = pending
  billable = false
```

### Complete 时机

请求结束时通过 `defer` 或统一收尾逻辑调用 Complete，覆盖成功、失败、中断和未知结果。

Complete 输入：

```text
request_id
status_code
duration_ms
first_token_ms
outcome
billable
input_tokens
output_tokens
total_cost
actual_cost
error_message
completed_at
```

Complete 行为：

```text
UPDATE request_records
SET
  completed_at = now(),
  updated_at = now(),
  outcome = ...,
  status_code = ...,
  duration_ms = ...,
  first_token_ms = ...,
  billable = ...,
  token/cost snapshot = ...,
  error_message = ...
WHERE request_id = ...
```

Complete 失败不应影响主请求返回。失败只写日志，避免请求记录模块影响核心网关链路。

## Session 解析规则

新增 `SessionResolver`，统一输出：

```go
type SessionIdentity struct {
    SessionID       string
    Source          string
    ClientSessionID string
}
```

### OpenAI / Codex

优先级：

```text
1. Header: session_id
2. Header: conversation_id
3. Body: prompt_cache_key
4. generated / unknown
```

### Claude / Anthropic

优先级：

```text
1. metadata.user_id 中的 session_id
2. metadata.session_id
3. 现有显式 metadata session 信号
4. generated / unknown
```

### 强弱关联

显式客户端 Session 信号才作为强关联：

```text
header_session_id
header_conversation_id
prompt_cache_key
metadata_user_id
metadata_session_id
```

内容 hash 或 fallback 只能标记为：

```text
generated
unknown
```

这类弱关联不能在页面上伪装成客户端真实对话 ID，避免误把不同对话归为同一个 Session。

## 查询接口

新增后台接口：

```text
GET /api/v1/admin/request-records
```

第一版不直接替换：

```text
GET /api/v1/admin/request-logs
```

原因：

- `/admin/request-logs` 是阶段一聚合接口，可作为回退。
- `/admin/request-records` 是阶段二/三 durable ledger 接口。
- 两者只在灰度、验证和回滚窗口内并存。

阶段二/三完成并稳定后，推荐将管理员请求记录页统一到 `request_records` 数据源，并退场阶段一聚合接口。退场对象包括：

```text
backend/internal/service/admin_request_logs.go
backend/internal/repository/ops_repo_admin_request_logs.go
backend/internal/handler/admin/request_log_handler.go
routes 中的 /api/v1/admin/request-logs 注册
frontend-local 中 aggregate 数据源和 fallback 逻辑
```

如果最终希望保留 `/api/v1/admin/request-logs` 这个 URL，也应只保留 URL 兼容层，让它转调 `request_records` 查询服务，而不是继续维护 `usage_logs + ops_error_logs` 聚合实现。

查询参数：

```text
page
page_size
start_time
end_time
outcome
billable
session_id
request_id
client_request_id
user_id
api_key_id
account_id
group_id
platform
model
status_code
request_type
stream
q
sort
```

查询以 `request_records` 为主表：

```text
request_records rr
LEFT JOIN usage_logs ul
  ON ul.request_id = rr.request_id

LEFT JOIN ops_error_logs o
  ON o.request_id = rr.request_id
  OR o.client_request_id = rr.client_request_id
```

字段优先级：

```text
请求事实
  优先使用 rr 字段。

计费细节
  来自 usage_logs。

错误细节
  来自 ops_error_logs。
```

## 前端接入

`RequestLogsView` 继续作为管理员请求记录入口。

API 层支持两个数据源：

```text
aggregate  当前 usage_logs + ops_error_logs 聚合接口
records    新 request_records 接口
```

切换策略：

```text
默认使用 records。
接口不可用或配置关闭时，可回退 aggregate。
```

`aggregate` fallback 是上线保护，不是长期产品能力。完成验证后应删除前端 fallback 分支，页面只保留 `records` 数据源，避免后续字段、筛选和状态展示在两套数据源之间分叉。

页面新增能力：

- Session 列。
- Session 筛选。
- 点击“同一 Session”快捷筛选。
- Outcome 展示。
- Billable 展示。
- 非计费和失败请求 cost 显示 `-`。

页面仍然不展示 prompt / response。

## 与使用记录和消费明细的关系

页面口径保持三分：

```text
账号管理 / 今日消费
  只基于 usage_logs。
  表达真实消费。

供应商消费明细
  只基于 usage_logs 和账号成本字段。
  表达供应商侧成本。

管理员请求记录
  基于 request_records。
  表达请求是否发生、由谁发出、属于哪个 Session、结果如何。
```

请求记录页可以展示 cost 快照，但账单真相源仍是 `usage_logs`。

## 迁移和兼容

本期只新增表，不回填历史数据。

上线后：

- 新请求进入 `request_records`。
- 老请求在回退窗口内仍可通过阶段一聚合接口查看。
- 前端可以根据接口可用性选择 `records` 或 `aggregate`。

如果需要回滚：

- 停用 requestrecord hook。
- 前端切回 aggregate 数据源。
- `usage_logs` 和 `ops_error_logs` 不受影响。

## 阶段一退场策略

为避免冗余代码，阶段一聚合实现需要有明确退场条件。

### 保留窗口

阶段二/三上线后，阶段一聚合接口只保留一个观察窗口。观察窗口用于：

- 验证 `request_records` 写入完整性。
- 对比 records 与 aggregate 的成功/失败请求数量。
- 确认 Session 筛选、状态筛选和 cost 展示符合预期。
- 支持异常时快速回退。

观察窗口结束后，不再同时维护两套请求流水后端。

### 退场条件

满足以下条件后可以删除阶段一聚合代码：

- 新请求都能稳定写入 `request_records`。
- 管理员请求记录页已默认使用 `records`。
- 常用筛选项在 `records` 数据源下可用。
- 成功计费请求可通过 `request_id` 关联 `usage_logs`。
- 失败请求可通过 `request_id` 或 `client_request_id` 关联 `ops_error_logs`。
- 回退窗口内未发现必须依赖 aggregate 的缺口。

### 删除范围

退场时删除：

```text
阶段一 service DTO 和 ListAdminRequestLogs 聚合服务
阶段一 repository CTE 查询
阶段一 handler 和路由注册
阶段一前端 aggregate API 方法
阶段一 RequestLogsView fallback 分支
阶段一相关单测
```

保留：

```text
usage_logs
ops_error_logs
OpsRequestDetail 运维详情能力
request_id / client_request_id 关联能力
```

`usage_logs` 和 `ops_error_logs` 仍是计费与错误事实源，不属于阶段一退场范围。

### URL 兼容策略

如果需要避免前端路由或外部书签变化，可以保留页面 URL：

```text
/admin/request-logs
```

但后端数据源应切换到：

```text
/api/v1/admin/request-records
```

如果必须保留旧 API URL：

```text
/api/v1/admin/request-logs
```

旧 API 只能作为 thin wrapper 转调 `request_records` 查询服务，不继续保留 `usage_logs + ops_error_logs` 聚合查询。

## 风险和应对

### 网关 hook 与 upstream 冲突

风险：网关文件是 upstream 高频变更区域，直接改大段逻辑容易冲突。

应对：

- hook 控制在少数几行。
- 业务逻辑放入 `backend/internal/requestrecord` 新模块。
- 不修改 billing/error 写入点。

### request_records 写入失败影响请求

风险：数据库异常时影响主请求。

应对：

- Start / Complete 失败只记录日志。
- 不阻断主请求。
- 后台页面允许短期缺失记录。

### Session 误关联

风险：fallback hash 被当成真实对话 ID。

应对：

- 保留 `session_source`。
- 页面区分强 Session 和 generated / unknown。
- 不把内容 hash 当强关联。

### Cost 口径混乱

风险：失败请求显示 `$0.00`，被误解为免费成功。

应对：

- `billable=false` 统一显示 `-`。
- 消费统计仍只读 `usage_logs`。

### 查询性能

风险：请求流水表增长后后台查询变慢。

应对：

- 默认时间范围保持最近 1 小时。
- 限制 `page_size`。
- 建立时间、Session、用户、API Key、账号索引。
- 后续可增加保留期和归档策略。

## 测试计划

后端单测：

- `SessionResolver`：
  - OpenAI `session_id`
  - OpenAI `conversation_id`
  - OpenAI `prompt_cache_key`
  - Claude `metadata.user_id`
  - Claude `metadata.session_id`
  - generated / unknown

- Repository：
  - Start 插入 pending。
  - Complete 更新 success。
  - Complete 更新 non_billable。
  - Complete 更新 error。
  - 按 session_id 查询。
  - 按 request_id 查询。
  - `billable=false` 不返回可计费展示值。

- Handler：
  - 参数解析。
  - 分页。
  - outcome 筛选。
  - billable 筛选。
  - session_id 筛选。

前端测试：

- `RequestLogsView` 展示 Session 列。
- 点击同一 Session 后追加筛选。
- `billable=false` 显示 `-`。
- `success` / `non_billable` / `error` / `cancelled` 状态展示正确。
- records 数据源不可用时能回退 aggregate。

集成验证：

- 成功计费请求进入 `request_records`，并能在请求记录页看到。
- 失败请求进入 `request_records`，并能关联错误摘要。
- 非计费成功请求进入 `request_records`，cost 显示 `-`。
- 带 `session_id` 的请求可按 Session 查询。

## 后续增强

阶段四可以择机加入：

- `usage_log_id` / `ops_error_log_id` 强关联。
- `(api_key_id, session_id)` 维度下的严格 `request_sequence`。
- 历史数据回填。
- Session 汇总抽屉。
- 请求详情页。
- 请求记录保留期和归档策略。
- 管理员可配置是否采集弱 Session。

这些增强不影响本期主目标，不应塞进阶段二/三的必做范围。
