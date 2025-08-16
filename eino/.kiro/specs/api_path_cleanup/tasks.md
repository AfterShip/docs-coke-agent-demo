# API Path Constants Cleanup Task List

基于对 `pkg/langfuse/api/resources` 目录的分析，需要对以下 9 个 API client 文件进行路径常量提取的清理工作。

## 任务清单

### 阶段 1：核心 API 客户端清理
- [ ] 1.1 清理 traces/client.go
  - [x] 1.1.1 提取所有 `/api/public/traces` 相关路径到文件头部常量
  - [x] 1.1.2 替换所有硬编码路径为常量引用
  - [x] 1.1.3 运行编译检查确保代码无编译错误
  - [x] 1.1.4 提交代码变更
  - [ ] 1.1.5 停止并等待人工确认后继续

- [ ] 1.2 清理 scores/client.go
  - [x] 1.2.1 提取所有 `/api/public/scores` 相关路径到文件头部常量
  - [x] 1.2.2 替换所有硬编码路径为常量引用
  - [x] 1.2.3 运行编译检查确保代码无编译错误
  - [x] 1.2.4 提交代码变更
  - [ ] 1.2.5 停止并等待人工确认后继续

- [ ] 1.3 清理 sessions/client.go
  - [x] 1.3.1 提取所有 `/api/public/sessions` 相关路径到文件头部常量
  - [x] 1.3.2 替换所有硬编码路径为常量引用
  - [x] 1.3.3 运行编译检查确保代码无编译错误
  - [x] 1.3.4 提交代码变更
  - [x] 1.3.5 停止并等待人工确认后继续

### 阶段 2：组织管理 API 客户端清理
- [x] 2.1 清理 organizations/client.go
  - [x] 2.1.1 提取所有 `/api/public/organizations` 相关路径到文件头部常量
  - [x] 2.1.2 替换所有硬编码路径为常量引用
  - [x] 2.1.3 运行编译检查确保代码无编译错误
  - [x] 2.1.4 提交代码变更
  - [x] 2.1.5 停止并等待人工确认后继续

- [x] 2.2 清理 projects/client.go
  - [x] 2.2.1 提取所有 `/api/public/projects` 相关路径到文件头部常量
  - [x] 2.2.2 替换所有硬编码路径为常量引用
  - [x] 2.2.3 运行编译检查确保代码无编译错误
  - [x] 2.2.4 提交代码变更
  - [x] 2.2.5 停止并等待人工确认后继续

### 阶段 3：数据和模型 API 客户端清理
- [x] 3.1 清理 datasets/client.go
  - [x] 3.1.1 提取所有 `/api/public/datasets` 相关路径到文件头部常量
  - [x] 3.1.2 替换所有硬编码路径为常量引用
  - [x] 3.1.3 运行编译检查确保代码无编译错误
  - [x] 3.1.4 提交代码变更
  - [x] 3.1.5 停止并等待人工确认后继续

- [x] 3.2 清理 models/client.go
  - [x] 3.2.1 提取所有 `/api/public/models` 相关路径到文件头部常量
  - [x] 3.2.2 替换所有硬编码路径为常量引用
  - [x] 3.2.3 运行编译检查确保代码无编译错误
  - [x] 3.2.4 提交代码变更
  - [x] 3.2.5 停止并等待人工确认后继续

### 阶段 4：系统服务 API 客户端清理
- [x] 4.1 清理 ingestion/client.go
  - [x] 4.1.1 提取所有 `/api/public/ingestion` 和 `/api/public/health` 相关路径到文件头部常量
  - [x] 4.1.2 替换所有硬编码路径为常量引用
  - [x] 4.1.3 运行编译检查确保代码无编译错误
  - [x] 4.1.4 提交代码变更
  - [x] 4.1.5 停止并等待人工确认后继续

- [x] 4.2 清理 health/client.go
  - [x] 4.2.1 提取所有 `/api/public/health` 相关路径到文件头部常量
  - [x] 4.2.2 替换所有硬编码路径为常量引用
  - [x] 4.2.3 运行编译检查确保代码无编译错误
  - [x] 4.2.4 提交代码变更
  - [x] 4.2.5 停止并等待人工确认后继续

### 阶段 5：最终验证
- [x] 5.1 整体验证和清理
  - [x] 5.1.1 运行项目完整编译检查确保所有变更无编译错误
  - [x] 5.1.2 运行所有相关单元测试验证功能完整性
  - [x] 5.1.3 验证所有 API 路径硬编码已完全移除
  - [x] 5.1.4 提交最终清理变更
  - [x] 5.1.5 停止并等待人工确认完成

## 任务说明

每个客户端文件的清理包括：
1. **路径提取**：将硬编码的 API 路径提取为文件头部的常量定义
2. **路径替换**：将所有硬编码路径替换为常量引用
3. **编译验证**：确保代码修改后能正常编译
4. **代码提交**：提交每个文件的清理变更

## 已完成文件

✅ `prompts/client.go` - 已完成清理（包含 `promptsBasePath` 和 `promptsUsageStatsPath` 常量）

## 待处理文件统计

| 文件 | 硬编码路径数量 | 主要路径模式 |
|------|------------|------------|
| traces/client.go | 8个 | `/api/public/traces*` |
| scores/client.go | 6个 | `/api/public/scores*` |
| sessions/client.go | 7个 | `/api/public/sessions*` |
| organizations/client.go | 10个 | `/api/public/organizations*` |
| projects/client.go | 9个 | `/api/public/projects*` |
| datasets/client.go | 13个 | `/api/public/datasets*` |
| models/client.go | 7个 | `/api/public/models*` |
| ingestion/client.go | 2个 | `/api/public/ingestion`, `/api/public/health` |
| health/client.go | 2个 | `/api/public/health` |