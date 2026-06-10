# CCH 风格改造功能映射记录

本文记录 `claude-code-hub` 截图/源码中的页面能力与 `sub2api` 当前前端能力的对应关系。
原则：已有能力按 CCH 页面习惯改造展示；缺失能力只记录，不新增前端假入口，也不改后端。

## 对照来源

- CCH 顶部壳层：`src/app/[locale]/dashboard/_components/dashboard-header.tsx`
- CCH 顶部导航：`src/app/[locale]/dashboard/_components/dashboard-nav.tsx`
- CCH 主容器：`src/app/[locale]/dashboard/_components/dashboard-main.tsx`
- CCH bento 卡片：`src/app/[locale]/dashboard/_components/bento/*`
- CCH Settings 菜单：`src/app/[locale]/settings/_lib/nav-items.ts`
- sub2api 路由真相源：`frontend/src/router/index.ts`
- sub2api 壳层：`frontend/src/components/layout/AppLayout.vue`
- sub2api 顶部导航：`frontend/src/components/layout/TocHeader.vue`

## 功能映射

| CCH 能力 | sub2api 对应能力 | 状态 | 处理方式 |
|---|---|---|---|
| Dashboard 概览指标 | `/dashboard`、`/admin/dashboard` | 存在 | 按 CCH bento 指标卡改造展示 |
| Usage Logs | `/usage`、`/admin/usage` | 存在 | 保留现有筛选/日志能力，统计卡、筛选工具条、图表面板按 CCH bento/panel/toolbar 细化 |
| API Keys | `/keys` | 存在 | 保留 key、quota、状态、复制、分组切换和客户端导入能力，页面摘要、筛选工具条和列表密度按 CCH 风格细化 |
| Availability | `/monitor`、`/admin/channels/monitor` | 部分存在 | 映射为渠道状态/渠道监控，不新增 CCH availability 路由 |
| Providers | `/admin/accounts`、`/admin/channels/pricing`、`/admin/groups`、`/admin/proxies` | 部分存在 | 拆分映射到账号、渠道定价、分组、代理；账号、渠道、分组、代理管理页按 CCH 标题摘要、轻工具条和列表质感改造 |
| Quotas | `/keys` key quota、`/subscriptions`、`/admin/subscriptions`、`/admin/groups` | 部分存在 | 保留分散额度能力，不新增独立 quotas 页面 |
| User Management | `/admin/users` | 存在 | 保留用户筛选、列设置、属性配置和弹窗能力，页面标题摘要、工具条和表格密度按 CCH list/table 风格改造 |
| Settings Config | `/admin/settings` | 部分存在 | 保留单页设置和现有保存接口，页面头、设置目录、内容卡片按 CCH settings layout 改造 |
| Status Page | `/monitor`、`/admin/channels/monitor` | 部分存在 | 只映射状态查看，不实现公开状态页配置 |
| Prices | `/admin/channels/pricing` | 部分存在 | 映射到渠道定价 |
| Sensitive Words | `/admin/risk-control` | 部分存在 | 映射到风控中心 |
| Error Rules | `/admin/ops` 告警规则、账号错误工具 | 部分存在 | 只保留现有入口，不新增独立 error rules 页 |
| Data / Backup | `/admin/settings` 备份配置 | 部分存在 | 保留在系统设置内 |
| Logs | `/admin/ops` 系统日志 | 部分存在 | 映射到运维监控 |
| Notifications | `/admin/announcements`、邮件/通知配置 | 部分存在 | 映射公告和现有设置 |
| API Docs | 顶栏 docs 外链 | 部分存在 | 保留外链，不新增 Scalar/OpenAPI UI |
| Billing / Orders | `/purchase`、`/orders`、`/admin/orders`、`/admin/orders/dashboard`、`/admin/orders/plans` | sub2api 特有 | CCH 没有等价内置充值/订阅订单管理；仅将页面外观对齐 CCH 的 header、toolbar、panel、table，不新增 CCH 假入口 |

## 缺失功能记录

以下 CCH 能力在 sub2api 当前前后端中没有等价页面或数据模型，本轮不实现：

- 独立活跃会话列表：`/dashboard/sessions`
- 会话消息详情：`/dashboard/sessions/:id/messages`
- 独立排行榜页：`/dashboard/leaderboard`
- 独立个人/全局配额页：`/dashboard/my-quota`、`/dashboard/quotas`
- 独立请求过滤配置：`/settings/request-filters`
- 客户端版本管理：`/settings/client-versions`
- 全局审计日志页：`/dashboard/audit-logs`
- 可配置公开状态页：`/settings/status-page`
- 内置 Scalar/OpenAPI 文档：`/api/v1/scalar`
- CCH settings 的独立 `prices`、`providers`、`logs`、`notifications` 子页拆分形态
- CCH Provider Manager 的完整 provider/endpoint 统一模型；sub2api 当前是“账号池 + 分组 + 渠道定价 + 代理”拆分模型，本轮只对齐视觉与管理习惯，不合并后端模型
- Dispatch Simulator / Scheduling Rules 独立模拟器
- CCH 内置 Billing / Recharge / Subscription Orders 管理页；sub2api 的支付订单体系为本项目独有能力，本轮只做 CCH 视觉对齐

## 本轮改造优先级

1. 壳层和导航：对齐 CCH 的 64px sticky 顶栏、`max-w-7xl` 主容器和 pill nav。
2. Dashboard / Usage / API Keys：已有数据全部用 bento 指标卡、panel card、toolbar card 呈现。
3. Settings 下拉与设置页：只展示 sub2api 已有管理能力；`/admin/settings` 保持单路由，视觉上对齐 CCH settings 目录 + 内容区。
4. 列表页共享布局：已先细化 `DataTable.vue` 的 CCH 表头、移动卡片、hover 与 sticky 列背景；后续继续覆盖更多管理列表页。
5. 支付/订单：因 CCH 无直接对应功能，保留 sub2api 原支付、订单、退款和套餐能力，只按 CCH 面板化结构改造展示。
6. Provider-like 管理页：渠道、分组、代理保留 sub2api 拆分模型，页面层级对齐 CCH `SettingsPageHeader` + `Section` + provider list 的管理习惯。
