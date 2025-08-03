# Agent 简化架构设计文档

## 概述

简化的 Agent 架构设计，基于 reasoning-action 模式，专注于实用性和易实现性。

## 设计原则

1. **简单优先**: 能用简单方案解决的，不用复杂设计
2. **渐进式**: 支持逐步完善，不要一次性设计过度
3. **实用导向**: 优先解决当前问题，保持扩展性
4. **最小可行**: 最小化核心组件，减少依赖关系

## 简化后的核心架构

### 基本流程
```
用户请求 → Job管理 → 推理执行 → 响应生成
```

### 核心概念

#### 1. 极简 Job 模型

```go
type Job struct {
    ID       string                 `json:"id"`
    Type     string                 `json:"type"`     // 固定为 "product_listing"
    Phase    Phase                  `json:"phase"`    // reasoning|acting|completed|failed
    
    // 核心状态
    Intent   string                 `json:"intent"`   // 用户意图: publish|query|edit|activate|deactivate  
    Context  map[string]interface{} `json:"context"`  // 收集的信息和状态
    
    // 简单错误处理
    Error    string                 `json:"error,omitempty"`
}

type Phase string
const (
    PhaseReasoning  Phase = "reasoning"  // 推理阶段：分析意图、收集信息
    PhaseActing     Phase = "acting"     // 行动阶段：执行具体操作
    PhaseCompleted  Phase = "completed"  // 完成
    PhaseFailed     Phase = "failed"     // 失败
)
```

#### 2. 简化的处理接口

```go
// 唯一的核心接口
type Processor interface {
    Process(ctx context.Context, job *Job, messages []Message) (*ProcessResult, error)
}

type ProcessResult struct {
    NextPhase       Phase  `json:"next_phase"`
    ResponseMessage string `json:"response_message"`
    NeedUserInput   bool   `json:"need_user_input"`
    IsCompleted     bool   `json:"is_completed"`
}
```

### 极简组件架构

#### 1. 无状态处理器

```go
// ProductListingProcessor - 纯函数式处理器
type ProductListingProcessor struct {
    genkitClient *genkit.Genkit
}

// 核心处理方法 - 纯函数，不依赖外部状态
func (p *ProductListingProcessor) Process(ctx context.Context, job *Job, messages []Message) (*ProcessResult, error) {
    switch job.Phase {
    case PhaseReasoning:
        return p.processReasoning(ctx, job, messages)
    case PhaseActing:
        return p.processActing(ctx, job, messages)
    default:
        return nil, errors.New("unknown phase")
    }
}
```

#### 2. 无状态化设计

```go
// 移除 JobStore，Job 状态完全由前端管理
// ProductListingProcessor 变为纯函数式处理
type ProductListingProcessor struct {
    genkitClient *genkit.Genkit
    // 移除 jobStore 字段
}

// Job 通过 metadata 在前后端传递
type ListingMetadata struct {
    Command string `json:"command"`
    Plan    Plan   `json:"plan"`
    Context JobContext `json:"context"`
}

type JobContext struct {
    Job *Job `json:"job,omitempty"`  // Job 对象直接存储在 metadata 中
}
```

type Phase string
const (
    PhaseIntentRecognition    Phase = "INTENT_RECOGNITION"
    PhaseInformationCollection Phase = "INFORMATION_COLLECTION"  
    PhasePlanGeneration       Phase = "PLAN_GENERATION"
    PhaseExecution           Phase = "EXECUTION"
    PhaseCompleted           Phase = "COMPLETED"
    PhaseFailed              Phase = "FAILED"
)

type JobStatus string
const (
    JobStatusPending    JobStatus = "pending"
    JobStatusInProgress JobStatus = "in_progress"
    JobStatusCompleted  JobStatus = "completed"
    JobStatusFailed     JobStatus = "failed"
)

type ExecutionContext struct {
    Parameters          map[string]interface{} `json:"parameters"`
    Tools               []string               `json:"tools"`
    Constraints         []string               `json:"constraints"`
    UserPreferences     UserPreferences        `json:"user_preferences"`
}

type IntermediateResult struct {
    Step        string      `json:"step"`
    Result      interface{} `json:"result"`
    Timestamp   time.Time   `json:"timestamp"`
    Success     bool        `json:"success"`
    ErrorMsg    string      `json:"error_msg,omitempty"`
}

type ValidationError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
    Code    string `json:"code"`
}
```

#### 元数据模型更新
```go
type ListingMetadata struct {
    Command string `json:"command"`
    Plan    Plan   `json:"plan"`
    Context JobContext `json:"context"`  // 新增Job上下文
}

type JobContext struct {
    Job     *Job   `json:"job,omitempty"`     // Job对象
    JobID   string `json:"job_id,omitempty"`  // Job ID引用
}
```

### 核心组件设计

#### 1. Job Manager (任务管理器)
负责 Job 的生命周期管理：

```go
type JobManager interface {
    CreateJob(ctx context.Context, command string, userMessage string) (*Job, error)
    GetJob(ctx context.Context, jobID string) (*Job, error)
    UpdateJob(ctx context.Context, job *Job) error
    TransitionPhase(ctx context.Context, jobID string, newPhase Phase) error
    ValidateJobInfo(ctx context.Context, job *Job) ([]ValidationError, error)
}
```

#### 2. Phase Handler (阶段处理器)
每个阶段对应一个处理器：

```go
type PhaseHandler interface {
    CanHandle(phase Phase) bool
    Process(ctx context.Context, job *Job, input *ProcessInput) (*ProcessOutput, error)
    GetNextPhase(ctx context.Context, job *Job) (Phase, error)
    GetRequiredInfo(ctx context.Context, job *Job) ([]string, error)
}

// 具体实现
type IntentRecognitionHandler struct{}
type InformationCollectionHandler struct{}
type PlanGenerationHandler struct{}
type ExecutionHandler struct{}
```

#### 3. Information Collector (信息收集器)
负责分析和收集任务所需信息：

```go
type InformationCollector interface {
    GetRequiredInfo(command string) []string
    ExtractInfo(ctx context.Context, messages []Message, requiredInfo []string) (map[string]interface{}, error)
    ValidateInfo(ctx context.Context, info map[string]interface{}, requirements []string) ([]ValidationError, error)
    GenerateCollectionPrompt(ctx context.Context, missingInfo []string) (string, error)
}
```

#### 4. Progress Tracker (进度跟踪器)
防止原地踏步：

```go
type ProgressTracker interface {
    HasProgressed(ctx context.Context, job *Job, newInfo map[string]interface{}) bool
    RecordProgress(ctx context.Context, job *Job, step string, result interface{}) error
    DetectLoop(ctx context.Context, job *Job) (bool, error)
    SuggestNextAction(ctx context.Context, job *Job) (string, error)
}
```

#### 5. Phase Corrector (阶段校正器)
根据实际状态校正阶段：

```go
type PhaseCorrector interface {
    // 校正阶段，返回应该处于的正确阶段
    CorrectPhase(ctx context.Context, job *Job) (Phase, bool, error)
    
    // 检查阶段前置条件是否满足
    ValidatePhasePrerequisites(ctx context.Context, job *Job, targetPhase Phase) (bool, []string, error)
    
    // 获取阶段回滚原因
    GetRollbackReason(ctx context.Context, job *Job, suggestedPhase Phase) (string, error)
}

// 阶段校正规则
type PhaseRule struct {
    TargetPhase   Phase
    Condition     func(*Job) bool
    Priority      int    // 优先级，数字越大优先级越高
    Reason        string // 校正原因
}

// 阶段校正器实现示例
type DefaultPhaseCorrector struct {
    rules []PhaseRule
}

func NewDefaultPhaseCorrector() *DefaultPhaseCorrector {
    return &DefaultPhaseCorrector{
        rules: []PhaseRule{
            // 优先级最高：如果命令为空，必须回到意图识别
            {
                TargetPhase: PhaseIntentRecognition,
                Condition:   func(job *Job) bool { return job.Command == "" },
                Priority:    100,
                Reason:      "命令未识别，需要重新进行意图识别",
            },
            // 信息缺失时回到信息收集阶段
            {
                TargetPhase: PhaseInformationCollection,
                Condition: func(job *Job) bool {
                    if job.Command == "" {
                        return false // 命令都没有，应该去意图识别
                    }
                    return !isPlanComplete(job.Plan, job.Command)
                },
                Priority: 90,
                Reason:   "Plan信息不完整，需要收集更多信息",
            },
            // 有命令但没有计划，应该在计划生成阶段
            {
                TargetPhase: PhasePlanGeneration,
                Condition: func(job *Job) bool {
                    return job.Command != "" && 
                           isPlanComplete(job.Plan, job.Command) &&
                           (job.Plan == nil || job.Plan.ID == "")
                },
                Priority: 80,
                Reason:   "信息已收集完整但缺少执行计划ID",
            },
            // 验证错误时回滚到信息收集
            {
                TargetPhase: PhaseInformationCollection,
                Condition: func(job *Job) bool {
                    return len(job.ValidationErrors) > 0
                },
                Priority: 85,
                Reason:   "信息验证失败，需要重新收集或修正信息",
            },
        },
    }
}

func (pc *DefaultPhaseCorrector) CorrectPhase(ctx context.Context, job *Job) (Phase, bool, error) {
    currentPhase := job.Phase
    
    // 按优先级排序规则
    sort.Slice(pc.rules, func(i, j int) bool {
        return pc.rules[i].Priority > pc.rules[j].Priority
    })
    
    // 检查每个规则
    for _, rule := range pc.rules {
        if rule.Condition(job) {
            if rule.TargetPhase != currentPhase {
                log.L(ctx).Info("Phase correction needed",
                    zap.String("from", string(currentPhase)),
                    zap.String("to", string(rule.TargetPhase)),
                    zap.String("reason", rule.Reason))
                return rule.TargetPhase, true, nil
            }
        }
    }
    
    // 无需校正
    return currentPhase, false, nil
}

func (pc *DefaultPhaseCorrector) ValidatePhasePrerequisites(ctx context.Context, job *Job, targetPhase Phase) (bool, []string, error) {
    var issues []string
    
    switch targetPhase {
    case PhaseIntentRecognition:
        // 意图识别阶段无前置条件
        return true, nil, nil
        
    case PhaseInformationCollection:
        if job.Command == "" {
            issues = append(issues, "命令未识别")
        }
        
    case PhasePlanGeneration:
        if job.Command == "" {
            issues = append(issues, "命令未识别")
        }
        if !isPlanComplete(job.Plan, job.Command) {
            issues = append(issues, "Plan信息不完整")
        }
        
    case PhaseExecution:
        if job.Command == "" {
            issues = append(issues, "命令未识别")
        }
        if job.Plan == nil || job.Plan.ID == "" {
            issues = append(issues, "缺少执行计划")
        }
        
    case PhaseCompleted, PhaseFailed:
        // 终态无需前置条件检查
        return true, nil, nil
    }
    
    return len(issues) == 0, issues, nil
}

// 辅助函数：重置Job状态
func resetJobState(ctx context.Context, job *Job, targetPhase Phase) {
    switch targetPhase {
    case PhaseIntentRecognition:
        // 回滚到意图识别，清空命令和后续状态
        job.Command = ""
        job.ValidationErrors = nil
        job.Plan = nil
        
    case PhaseInformationCollection:
        // 回滚到信息收集，保留命令但清空Plan和错误
        job.ValidationErrors = nil
        job.Plan = nil
        
    case PhasePlanGeneration:
        // 回滚到计划生成，只清空计划ID（保留收集的信息）
        if job.Plan != nil {
            job.Plan.ID = ""
        }
        job.ValidationErrors = nil
    }
    
    log.L(ctx).Info("Job state reset for phase rollback",
        zap.String("targetPhase", string(targetPhase)),
        zap.String("jobId", job.ID))
}

// 辅助函数：检查Plan是否完整
func isPlanComplete(plan *Plan, command string) bool {
    if plan == nil {
        return false
    }
    
    switch command {
    case "publish":
        return len(plan.Publish) > 0 && plan.Publish[0].ProductCenterID != ""
    case "query":
        return len(plan.Query) > 0 && plan.Query[0].ProductListingID != ""
    case "edit":
        return len(plan.Edit) > 0 && plan.Edit[0].ProductListingID != ""
    case "activate":
        return len(plan.Activate) > 0 && plan.Activate[0].ProductListingID != ""
    case "deactivate":
        return len(plan.Deactivate) > 0 && plan.Deactivate[0].ProductListingID != ""
    default:
        return false
    }
}
```

### 无状态主流程

```go
func ProductListingsFlow(ctx context.Context, input ProductListingInput) (ProductListingOutput, error) {
    // 1. 从 metadata 获取或创建 Job
    job := getJobFromMetadata(input.Metadata)
    if job == nil {
        job = &Job{
            ID:    generateJobID(),
            Type:  "product_listing", 
            Phase: PhaseReasoning,
            Context: make(map[string]interface{}),
        }
    }
    
    // 2. 处理 Job（纯函数式）
    processor := NewProductListingProcessor(genkitClient)
    result, err := processor.Process(ctx, job, input.Messages)
    if err != nil {
        job.Phase = PhaseFailed
        job.Error = err.Error()
        return ProductListingOutput{}, err
    }
    
    // 3. 更新 Job 状态（不持久化，由前端管理）
    job.Phase = result.NextPhase
    
    // 4. 生成响应，Job 状态返回给前端
    return ProductListingOutput{
        Message: Message{
            Role:    ASSISTANT_ROLE,
            Content: result.ResponseMessage,
        },
        Metadata: ListingMetadata{
            Command: job.Intent,
            Context: JobContext{Job: job},  // Job 状态返回给前端
        },
    }, nil
}

// 辅助函数：从 metadata 提取 Job
func getJobFromMetadata(metadata ListingMetadata) *Job {
    if metadata.Context.Job != nil {
        return metadata.Context.Job
    }
    return nil
}
```

#### 具体处理逻辑

##### 推理阶段处理
```go
func (p *ProductListingProcessor) processReasoning(ctx context.Context, job *Job, messages []Message) (*ProcessResult, error) {
    // 1. 识别意图（如果未识别）
    if job.Intent == "" {
        intent, err := p.recognizeIntent(ctx, messages)
        if err != nil {
            return nil, err
        }
        
        if intent == "" {
            // 需要澄清意图
            return &ProcessResult{
                NextPhase:       PhaseReasoning,
                ResponseMessage: "请告诉我您想要对商品进行什么操作？(发布/查询/编辑/激活/停用)",
                NeedUserInput:   true,
            }, nil
        }
        
        job.Intent = intent
    }
    
    // 2. 收集必要信息
    missing := p.checkMissingInfo(job.Intent, job.Context)
    if len(missing) > 0 {
        prompt := p.generateInfoPrompt(missing)
        return &ProcessResult{
            NextPhase:       PhaseReasoning,
            ResponseMessage: prompt,
            NeedUserInput:   true,
        }, nil
    }
    
    // 3. 信息收集完成，进入执行阶段
    return &ProcessResult{
        NextPhase:       PhaseActing,
        ResponseMessage: "信息已收集完成，开始执行操作...",
        NeedUserInput:   false,
    }, nil
}
```

##### 执行阶段处理
```go
func (p *ProductListingProcessor) processActing(ctx context.Context, job *Job, messages []Message) (*ProcessResult, error) {
    // 根据意图执行相应操作
    switch job.Intent {
    case "publish":
        err := p.executePublish(ctx, job.Context)
        if err != nil {
            return nil, err
        }
        return &ProcessResult{
            NextPhase:       PhaseCompleted,
            ResponseMessage: "商品发布成功！",
            IsCompleted:     true,
        }, nil
        
    case "query":
        result, err := p.executeQuery(ctx, job.Context)
        if err != nil {
            return nil, err
        }
        return &ProcessResult{
            NextPhase:       PhaseCompleted,
            ResponseMessage: fmt.Sprintf("查询结果：%v", result),
            IsCompleted:     true,
        }, nil
        
    // ... 其他操作类似
    }
    
    return nil, errors.New("unknown intent")
}
```
```

#### 各阶段处理逻辑

##### 1. Intent Recognition Handler
```go
func (h *IntentRecognitionHandler) Process(ctx context.Context, job *Job, input *ProcessInput) (*ProcessOutput, error) {
    if job.Command != "" {
        // 命令已识别，直接进入下一阶段
        return &ProcessOutput{
            NeedsUserInteraction: false,
            NextPhase:           PhaseInformationCollection,
        }, nil
    }
    
    // 执行意图识别
    intentResult, err := recognizeIntent(ctx, input.Messages)
    if err != nil {
        return nil, err
    }
    
    if intentResult.Command == "" {
        // 需要澄清意图
        return &ProcessOutput{
            NeedsUserInteraction: true,
            UserPrompt:          intentResult.ClarificationQuestion,
            NextPhase:           PhaseIntentRecognition,
        }, nil
    }
    
    // 更新Job命令
    job.Command = intentResult.Command
    job.RequiredInfo = getRequiredInfo(intentResult.Command)
    
    return &ProcessOutput{
        NeedsUserInteraction: false,
        NextPhase:           PhaseInformationCollection,
    }, nil
}
```

##### 2. Information Collection Handler
```go
func (h *InformationCollectionHandler) Process(ctx context.Context, job *Job, input *ProcessInput) (*ProcessOutput, error) {
    // 提取新信息
    newInfo, err := extractInfo(ctx, input.Messages, job.RequiredInfo)
    if err != nil {
        return nil, err
    }
    
    // 检查进度
    if !hasProgressed(ctx, job, newInfo) {
        return &ProcessOutput{
            NeedsUserInteraction: true,
            UserPrompt:          generateProgressPrompt(ctx, job),
            NextPhase:           PhaseInformationCollection,
        }, nil
    }
    
    // 合并信息
    mergeInfo(job.CollectedInfo, newInfo)
    
    // 验证信息完整性
    validationErrors := validateInfo(ctx, job.CollectedInfo, job.RequiredInfo)
    if len(validationErrors) > 0 {
        job.ValidationErrors = validationErrors
        job.MissingInfo = getMissingInfo(job.RequiredInfo, job.CollectedInfo)
        
        return &ProcessOutput{
            NeedsUserInteraction: true,
            UserPrompt:          generateCollectionPrompt(ctx, job.MissingInfo),
            NextPhase:           PhaseInformationCollection,
        }, nil
    }
    
    // 信息收集完成
    return &ProcessOutput{
        NeedsUserInteraction: false,
        NextPhase:           PhasePlanGeneration,
    }, nil
}
```

##### 3. Plan Generation Handler
```go
func (h *PlanGenerationHandler) Process(ctx context.Context, job *Job, input *ProcessInput) (*ProcessOutput, error) {
    if job.Plan != nil && job.Plan.ID != "" {
        // 计划已生成，进入执行阶段
        return &ProcessOutput{
            NeedsUserInteraction: false,
            NextPhase:           PhaseExecution,
        }, nil
    }
    
    // 生成执行计划
    planInput := createPlanInput(job)
    planResult, err := generatePlan(ctx, planInput)
    if err != nil {
        return nil, err
    }
    
    if planResult.Plan == nil {
        // 需要更多信息生成计划
        return &ProcessOutput{
            NeedsUserInteraction: true,
            UserPrompt:          planResult.ClarificationQuestion,
            NextPhase:           PhasePlanGeneration,
        }, nil
    }
    
    // 更新Job计划
    job.Plan = planResult.Plan
    
    return &ProcessOutput{
        NeedsUserInteraction: false,
        NextPhase:           PhaseExecution,
    }, nil
}
```

##### 4. Execution Handler
```go
func (h *ExecutionHandler) Process(ctx context.Context, job *Job, input *ProcessInput) (*ProcessOutput, error) {
    // 执行任务
    executionResult, err := executeJob(ctx, job)
    if err != nil {
        job.Status = JobStatusFailed
        return nil, err
    }
    
    // 记录执行结果
    job.IntermediateResults = append(job.IntermediateResults, executionResult)
    job.Status = JobStatusCompleted
    
    return &ProcessOutput{
        NeedsUserInteraction: false,
        NextPhase:           PhaseCompleted,
        ExecutionResult:     executionResult,
    }, nil
}
```

#### 阶段校正使用示例

以下是一些典型的阶段校正场景：

**场景1: 信息验证失败回滚**
```
用户在 PlanGeneration 阶段，但提供的 product_listing_id 验证失败
→ PhaseCorrector 检测到 ValidationErrors 不为空
→ 校正到 PhaseInformationCollection 阶段
→ 重置 ValidationErrors，提示用户提供正确的 product_listing_id
```

**场景2: 意图识别失败回滚**  
```
用户在 InformationCollection 阶段，但 Command 为空（可能被错误清空）
→ PhaseCorrector 检测到 Command 为空
→ 校正到 PhaseIntentRecognition 阶段
→ 重置所有状态，重新进行意图识别
```

**场景3: 跨阶段信息缺失**
```
用户直接跳到 Execution 阶段，但缺少必要信息
→ PhaseCorrector 检测到信息不完整
→ 校正到 PhaseInformationCollection 阶段
→ 保留已有命令，重新收集缺失信息
```

### 关键特性

#### 1. 智能阶段校正机制
- **状态驱动**: 根据 Job 实际状态而非简单字段决定阶段
- **优先级规则**: 支持多条件冲突时的优先级处理
- **自动回滚**: 发现问题时自动回滚到正确阶段
- **状态重置**: 回滚时智能重置相关状态，避免脏数据

#### 2. 进度防护机制
- **进度检测**: 每次交互检查是否有新信息收集
- **循环检测**: 识别重复询问相同信息的情况  
- **智能提示**: 根据当前状态生成针对性提示

#### 2. 信息验证机制
- **格式验证**: 检查信息格式是否正确
- **业务验证**: 检查信息是否符合业务规则
- **完整性验证**: 检查必需信息是否齐全

#### 3. 状态持久化
- **Job存储**: 支持Job状态的持久化存储
- **会话恢复**: 支持会话中断后的状态恢复
- **历史追踪**: 完整记录任务执行历史

### 实现计划

#### 阶段1: 基础架构
1. 定义新的数据模型 (Job, Phase, 等)
2. 实现 JobManager 基础功能
3. 创建 PhaseHandler 接口和基础实现

#### 阶段2: 核心逻辑
1. 实现各阶段的具体处理逻辑
2. 实现信息收集和验证机制
3. 实现进度跟踪和防护机制

#### 阶段3: 集成测试
1. 重构现有 product_listings 流程
2. 集成新的阶段处理器
3. 完整的端到端测试

#### 阶段4: 优化完善
1. 性能优化和错误处理完善
2. 添加监控和日志
3. 文档和示例完善

### Product Listings 具体实现示例

#### 1. Product Listings 专用组件
```go
type ProductListingsReasoningEngine struct {
    base ReasoningEngine
}

func (pl *ProductListingsReasoningEngine) DecomposeGoals(ctx context.Context, intent *Intent, context map[string]interface{}) ([]Goal, error) {
    var goals []Goal
    
    switch intent.Type {
    case "publish":
        goals = []Goal{
            {
                ID:          "collect_product_center_id",
                Type:        "information_gathering",
                Description: "收集 product_center_id",
                Status:      GoalStatusPending,
                Priority:    1,
            },
            {
                ID:          "execute_publish",
                Type:        "task_execution", 
                Description: "执行商品发布",
                Status:      GoalStatusPending,
                Priority:    2,
                Dependencies: []string{"collect_product_center_id"},
            },
        }
    case "edit":
        goals = []Goal{
            {
                ID:          "collect_product_listing_id",
                Type:        "information_gathering",
                Description: "收集 product_listing_id",
                Status:      GoalStatusPending,
                Priority:    1,
            },
            {
                ID:          "collect_product_data",
                Type:        "information_gathering",
                Description: "收集商品编辑数据",
                Status:      GoalStatusPending,
                Priority:    2,
                Dependencies: []string{"collect_product_listing_id"},
            },
            {
                ID:          "execute_edit",
                Type:        "task_execution",
                Description: "执行商品编辑",
                Status:      GoalStatusPending,
                Priority:    3,
                Dependencies: []string{"collect_product_listing_id", "collect_product_data"},
            },
        }
    }
    
    return goals, nil
}
```

#### 2. 通用 Action 类型
```go
var CommonActionTypes = []ActionType{
    {
        Type:        "ask_user",
        Description: "向用户询问信息",
        Category:    "user_interaction",
        Parameters: []ParameterSpec{
            {Name: "question", Type: "string", Required: true},
            {Name: "expected_format", Type: "string", Required: false},
        },
    },
    {
        Type:        "call_api",
        Description: "调用外部API",
        Category:    "api_call",
        Parameters: []ParameterSpec{
            {Name: "endpoint", Type: "string", Required: true},
            {Name: "method", Type: "string", Required: true},
            {Name: "params", Type: "object", Required: false},
        },
    },
    {
        Type:        "validate_data",
        Description: "验证数据格式和有效性",
        Category:    "data_processing",
        Parameters: []ParameterSpec{
            {Name: "data", Type: "object", Required: true},
            {Name: "schema", Type: "string", Required: true},
        },
    },
    {
        Type:        "extract_info",
        Description: "从用户消息中提取信息",
        Category:    "data_processing",
        Parameters: []ParameterSpec{
            {Name: "messages", Type: "array", Required: true},
            {Name: "target_fields", Type: "array", Required: true},
        },
    },
}
```

## 简化设计的优势

### ✨ 设计亮点

1. **极简架构**: 只有 2 个核心组件（Job、Processor）
2. **完全无状态**: 后端不存储任何状态，Job 通过前端传递
3. **纯函数式**: Processor 是纯函数，易于测试和理解
4. **前端存储**: 通过 metadata.context 传递状态，简化后端设计

### 🎯 核心特性

1. **reasoning-action 循环**: 符合 Agent 标准模式
2. **渐进式信息收集**: 通过 Context 逐步收集所需信息
3. **状态持久化**: 支持会话恢复和进度跟踪
4. **易于扩展**: 基础架构简单，后续可按需扩展

### 📝 实现路径

1. **第一步**: 实现 Job 模型和 metadata 传递逻辑
2. **第二步**: 实现 ProductListingProcessor 的 reasoning 逻辑
3. **第三步**: 实现具体的执行操作（publish/query/edit等）
4. **第四步**: 集成到现有的 product_listings flow

### 🚀 预期效果

- **解决原地踏步**: 通过前端状态管理和信息收集避免重复询问
- **提升可维护性**: 无状态设计更易于理解、测试和调试
- **零存储依赖**: 后端无需数据库或缓存，降低部署复杂度
- **快速上线**: 最小化实现复杂度，加快交付速度

这个无状态设计既保留了 reasoning-action 模式的核心优势，又彻底简化了后端架构，是一个极其实用的解决方案。通过前端管理状态，后端专注于逻辑处理，实现了完美的职责分离。