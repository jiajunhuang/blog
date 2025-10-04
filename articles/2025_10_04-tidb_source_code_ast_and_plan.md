# TiDB 源码阅读（四）：AST、逻辑计划、物理计划

前面我们已经看过服务端如何监听，以及接受请求，开始处理请求到最后返回数据的大体流程。这篇文章中，我们再度探索
AST、逻辑计划和物理计划，以求有更深入的了解。

## 为什么需要AST

**AST** 的全称是 **抽象语法树(Abstract Syntax Tree)**。

简单来说，当你在数据库客户端输入一条SQL语句，比如：

```sql
SELECT name, age FROM users WHERE age > 18;
```

数据库系统并不会直接去理解和执行这串文本字符。它做的第一件事，就是像编译器和解释器一样，将这条SQL语句“翻译”成一个内部的数据结构。这个数据结构就是**AST**。

**AST是什么样子的？**

AST是一个树形结构，它：

*   **抛弃了不重要的细节**：比如空格、换行符、定界符（如分号）等。
*   **保留了代码的语法结构**：将SQL中的关键字、表名、列名、操作符、值等，按照它们的语法关系组织成一棵树。
*   **每个节点都代表一个语法结构**：
    *   **根节点** 可能代表整个 `SELECT` 语句。
    *   **子节点** 可能分别代表 `SELECT` 列表、`FROM` 子句、`WHERE` 子句。
    *   在 `SELECT` 列表节点下，又会有两个子节点，分别代表列 `name` 和 `age`。
    *   在 `WHERE` 子句节点下，会有一个代表“大于”操作的节点，这个节点又有两个子节点，分别代表列 `age` 和数字 `18`。

上面那条SQL的AST可以简化为以下结构：

```
          [SelectStmt]
          /     |     \
     [TargetList] [FromClause] [WhereClause]
        /   \        |            |
 [ColumnRef] ... [TableRef]   [A_Expr (>)]
    /   \          |           /        \
[name] [age]    [users]  [ColumnRef]  [Const (18)]
                                |
                              [age]
```

**核心思想**：AST是源代码语法结构的一种抽象表示，它用树的形式清晰地表达了代码的层次和组成关系。

因此，AST的作用，就是让客户端传来的SQL变得让服务端可理解，可操作。

## 逻辑计划

那么有了AST之后，是在哪里开始转化成逻辑执行计划的呢？在 `pkg/planner/optimize.go` 中的 `optimize` 函数：

```go
func optimize(ctx context.Context, sctx planctx.PlanContext, node *resolve.NodeW, is infoschema.InfoSchema) (base.Plan, types.NameSlice, float64, error) {
	// ... 省略部分代码 ...

	// build logical plan
	hintProcessor := hint.NewQBHintHandler(sctx.GetSessionVars().StmtCtx)
	node.Node.Accept(hintProcessor)
	defer hintProcessor.HandleUnusedViewHints()
	builder := planBuilderPool.Get().(*core.PlanBuilder)
	defer planBuilderPool.Put(builder.ResetForReuse())
	builder.Init(sctx, is, hintProcessor)
	p, err := buildLogicalPlan(ctx, sctx, node, builder)  // 调用 buildLogicalPlan
    // ... 省略部分代码 ...
}

func buildLogicalPlan(ctx context.Context, sctx planctx.PlanContext, node *resolve.NodeW, builder *core.PlanBuilder) (base.Plan, error) {
	sctx.GetSessionVars().PlanID.Store(0)
	sctx.GetSessionVars().PlanColumnID.Store(0)
	sctx.GetSessionVars().MapScalarSubQ = nil
	sctx.GetSessionVars().MapHashCode2UniqueID4ExtendedCol = nil

	// ... 省略 failpoint ...

	// reset fields about rewrite
	sctx.GetSessionVars().RewritePhaseInfo.Reset()
	beginRewrite := time.Now()
	p, err := builder.Build(ctx, node)  // 核心：调用 PlanBuilder.Build
	if err != nil {
		return nil, err
	}
	sctx.GetSessionVars().RewritePhaseInfo.DurationRewrite = time.Since(beginRewrite)

    // ... 省略部分代码 ...
}

// Build builds the ast node to a Plan.
func (b *PlanBuilder) Build(ctx context.Context, node *resolve.NodeW) (base.Plan, error) {
	err := b.checkSEMStmt(node.Node)
	if err != nil {
		return nil, err
	}

	b.resolveCtx = node.GetResolveContext()
	b.optFlag |= rule.FlagPruneColumns
	switch x := node.Node.(type) {
	case *ast.AdminStmt:
		return b.buildAdmin(ctx, x)
	case *ast.DeallocateStmt:
		return &Deallocate{Name: x.Name}, nil
	case *ast.DeleteStmt:
		return b.buildDelete(ctx, x)
	case *ast.ExecuteStmt:
		return b.buildExecute(ctx, x)
	case *ast.ExplainStmt:
		return b.buildExplain(ctx, x)
	case *ast.ExplainForStmt:
		return b.buildExplainFor(x)
	case *ast.TraceStmt:
		return b.buildTrace(x)
	case *ast.InsertStmt:
		return b.buildInsert(ctx, x)
	// ...
	case *ast.SelectStmt:  // SELECT 语句
		if x.SelectIntoOpt != nil {
			return b.buildSelectInto(ctx, x)
		}
		return b.buildSelect(ctx, x)  // 调用 buildSelect
	case *ast.SetOprStmt:
		return b.buildSetOpr(ctx, x)
	case *ast.UpdateStmt:
		return b.buildUpdate(ctx, x)
	// ... 更多 case ...
    }
    // ... 省略部分代码 ...
}
```

比如 select 语句，就是在 `buildSelect` 函数中：

```go
func (b *PlanBuilder) buildSelect(ctx context.Context, sel *ast.SelectStmt) (p base.LogicalPlan, err error) {
	b.pushSelectOffset(sel.QueryBlockOffset)
	b.pushTableHints(sel.TableHints, sel.QueryBlockOffset)
	defer func() {
		b.popSelectOffset()
		// table hints are only visible in the current SELECT statement.
		b.popTableHints()
	}()
	// ... 省略部分代码 ...

	// 这个方法会构建各种逻辑算子，如：
	// - LogicalDataSource（数据源）
	// - LogicalSelection（过滤条件）
	// - LogicalJoin（连接）
	// - LogicalAggregation（聚合）
	// - LogicalProjection（投影）
	// 等等
}
```

根据AST构建了逻辑执行计划之后，接下来做的事情就是优化逻辑执行计划：

```go
	names := p.OutputNames()

	// Handle the non-logical plan statement.
	logic, isLogicalPlan := p.(base.LogicalPlan)
	if !isLogicalPlan {
		return p, names, 0, nil
	}

	core.RecheckCTE(logic)

	beginOpt := time.Now()
	finalPlan, cost, err := core.DoOptimize(ctx, sctx, builder.GetOptFlag(), logic)  // 逻辑优化 + 物理优化
	// TODO: capture plan replayer here if it matches sql and plan digest

	sessVars.DurationOptimization = time.Since(beginOpt)
	return finalPlan, names, cost, err
```

有两种优化方式，分别是：

```go
// doOptimize optimizes a logical plan into a physical plan,
// while also returning the optimized logical plan, the final physical plan, and the cost of the final plan.
// The returned logical plan is necessary for generating plans for Common Table Expressions (CTEs).
func doOptimize(ctx context.Context, sctx base.PlanContext, flag uint64, logic base.LogicalPlan) (
	base.LogicalPlan, base.PhysicalPlan, float64, error) {
	if sctx.GetSessionVars().GetSessionVars().EnableCascadesPlanner {
		return CascadesOptimize(ctx, sctx, flag, logic)
	}
	return VolcanoOptimize(ctx, sctx, flag, logic)
}
```

### VolcanoOptimize vs CascadesOptimize

假设你要从 A 地到 B 地：

- **Volcano 优化器**：像是用导航软件规划路线
  - 先决定走哪条路（逻辑优化）
  - 再决定用什么交通工具（物理优化）
  - 直接、高效，适合大多数场景

- **Cascades 优化器**：像是有一个智能助手帮你规划
  - 把所有可能的路线都记录在"小本本"（Memo）里
  - 发现两条路其实是一样的？合并！不重复考虑
  - 系统化地探索更多可能性，找到更优解

### 核心区别对比

#### 1. 优化流程

**Volcano（两阶段流水线）:**
```
SQL 语句
  ↓
逻辑优化（应用各种规则）
  ↓
物理优化（选择执行方式）
  ↓
执行计划
```

**Cascades（Memo + 任务驱动）:**
```
SQL 语句
  ↓
归一化（预处理）
  ↓
构建 Memo（记录所有等价方案）
  ↓
任务调度（系统化搜索）
  ↓
执行计划
```

#### 2. 数据结构

**Volcano:**
- 直接在计划树上操作
- 就像在一棵树上不断修剪、嫁接

**Cascades:**
- 使用 **Memo** 这个"记忆本"
  - `Group`：存储逻辑上等价的计划（比如 `A JOIN B` 和 `B JOIN A`）
  - `GroupExpression`：具体的某种实现方式
- 好处：避免重复计算，提高效率

#### 3. 搜索策略

| 特性 | Volcano | Cascades |
|------|---------|----------|
| 搜索方式 | 自顶向下递归 | 任务驱动 |
| 去重机制 | 有限 | 完善的 hash 去重 |
| 搜索空间 | 相对保守 | 更全面 |
| 复杂度 | 简单直接 | 更复杂但更强大 |

### 火山模型

在此，我想再聊聊火山模型：

火山模型，也称为迭代器模型，是数据库系统中一种经典的查询执行引擎的实现模型。它的核心思想是：查询计划中的每个运算符（Operator）都实现一个统一的、简单的接口，数据像火山喷发一样，通过调用这些接口，以一个元组（Tuple，即一行数据）为单位的“流”式地从底层运算符传递到顶层运算符。

这个统一的接口通常包含一个 next() 方法：

next(): 调用该方法，运算符会返回下一个元组；如果已经没有更多元组，则返回一个结束标记（如 EOF）。

在整个执行过程中，数据是以一次一行的方式在各个运算符之间“流动”的。每个运算符都像一个小处理器，通过 next() 方法“拉取”数据。

非常的简单明了。但是，缺点是啥呢？就是很多函数调用，效率不高，TiDB采用的解决方案是，从一次一行数据，改为一次一个chunk。

```go
// Next implements the Executor Next interface.
func (e *SelectLockExec) Next(ctx context.Context, req *chunk.Chunk) error {
    // ...
}
```

除此之外，还能想到的一个缺点就是，如果底层不断的返回数据，全都要上层来判断是抛弃还是保留数据的话，效率会非常低，约等于
全表扫描。那么解决办法是啥呢？就是想办法把数据过滤，下推到真正查找数据的那一层，这就是谓词下推。

关于火山模型我们就聊这么多，接下来，我们继续看看物理计划。

## 物理计划

优化完逻辑计划的下一步，就是构建物理计划。物理计划，意思就是涉及到如何操作数据的计划，已经贴近数据层的操作了。

```go
func doOptimize(ctx context.Context, sctx base.PlanContext, flag uint64, logic base.LogicalPlan) (
	base.LogicalPlan, base.PhysicalPlan, float64, error)
```

这个函数将逻辑计划转换成物理计划，最后返回逻辑计划，物理计划，以及计划的成本。

```go
func physicalOptimize(logic base.LogicalPlan) (plan base.PhysicalPlan, cost float64, err error) {
    // 1. 计算统计信息
    logic.RecursiveDeriveStats(nil)

    // 2. 准备可能的属性
    preparePossibleProperties(logic)

    // 3. 创建根节点的物理属性需求
    prop := &property.PhysicalProperty{
        TaskTp:      property.RootTaskType,
        ExpectedCnt: math.MaxFloat64,
    }

    // 4. 【关键】调用 FindBestTask 寻找最优物理计划
    t, err := logic.FindBestTask(prop)  // 第1099行

    // 5. 从 Task 中提取物理计划
    return t.Plan(), cost, err
}
```

```go
func findBestTask(super base.LogicalPlan, prop *property.PhysicalProperty) (bestTask base.Task, err error) {
    // 1. 检查缓存，避免重复计算
    bestTask = p.GetTask(prop)
    if bestTask != nil {
        return bestTask, nil
    }

    // 2. 枚举当前逻辑算子所有可能的物理实现
    physicalPlans, hintWorksWithProp, err := exhaustPhysicalPlans(p.Self(), newProp)

    // 3. 对每个物理算子，递归获取子节点的最优物理计划
    // 4. 计算代价并选择代价最低的
    // 5. 缓存结果
}
```

## 以一个实际SQL为例子

```sql
SELECT name, age
FROM users
WHERE age > 18 AND city = 'Beijing'
ORDER BY age DESC
LIMIT 10
```

### 第一步：SQL → 逻辑计划（Logical Plan）

#### 1. **解析阶段（Parse）**
SQL 首先被解析器转换成 AST（抽象语法树）：
```
ast.SelectStmt {
    Fields: [name, age]
    From: TableSource{users}
    Where: BinaryExpr(AND) {
        Left: age > 18
        Right: city = 'Beijing'
    }
    OrderBy: age DESC
    Limit: 10
}
```

#### 2. **构建逻辑计划（Build Logical Plan）**
代码路径在 `logical_plan_builder.go:911-939`：

```go
func (b *PlanBuilder) buildSelection(ctx context.Context, p base.LogicalPlan,
    where ast.ExprNode, aggMapper map[*ast.AggregateFuncExpr]int) (base.LogicalPlan, error) {

    // 第921行：创建 LogicalSelection
    selection := logicalop.LogicalSelection{}.Init(b.ctx, b.getSelectOffset())

    // 第922-938行：将 WHERE 条件转换为 Expression
    for _, cond := range conditions {
        expr, np, err := b.rewrite(ctx, cond, p, aggMapper, false)
        expressions = append(expressions, expr)
    }

    // 第940-950行：转换为 CNF（合取范式）
    for _, expr := range expressions {
        cnfItems := expression.SplitCNFItems(expr)
        cnfExpres = append(cnfExpres, cnfItems...)
    }
}
```

**生成的逻辑计划树：**
```
LogicalLimit (limit=10)
    ↓
LogicalSort (orderBy=[age DESC])
    ↓
LogicalProjection (columns=[name, age])
    ↓
LogicalSelection (conditions=[age > 18, city = 'Beijing'])
    ↓
DataSource (table=users)
```

#### 3. **逻辑优化（Logical Optimization）**
在 `optimizer.go:1040-1074` 的 `logicalOptimize` 函数中，应用各种逻辑优化规则：

```go
func logicalOptimize(ctx context.Context, flag uint64, logic base.LogicalPlan) {
    // 应用优化规则列表
    for i, rule := range logicalRuleList {
        logic, _, err = rule.Optimize(ctx, logic)
    }
}
```

**主要优化规则包括：**
- **列裁剪（Column Pruning）**：只保留 `name`, `age`, `city` 列
- **谓词下推（Predicate Push Down）**：把 `WHERE` 条件下推到 DataSource
- **常量传播（Constant Propagation）**
- **投影消除（Projection Elimination）**

**优化后的逻辑计划：**
```
LogicalLimit (limit=10)
    ↓
LogicalSort (orderBy=[age DESC])
    ↓
DataSource (table=users,
            columns=[name, age, city],
            pushed_conditions=[age > 18, city = 'Beijing'])
```

---

### 第二步：逻辑计划 → 物理计划（Physical Plan）

#### 1. **物理优化入口**
在 `optimizer.go:1082-1119` 的 `physicalOptimize` 函数：

```go
func physicalOptimize(logic base.LogicalPlan) (plan base.PhysicalPlan, cost float64, err error) {
    // 1. 计算统计信息
    logic.RecursiveDeriveStats(nil)

    // 2. 准备物理属性
    preparePossibleProperties(logic)

    // 3. 创建物理属性需求
    prop := &property.PhysicalProperty{
        TaskTp:      property.RootTaskType,
        ExpectedCnt: math.MaxFloat64,
    }

    // 4. 【关键】调用 FindBestTask 递归生成物理计划
    t, err := logic.FindBestTask(prop)  // ← 这里开始转换！

    return t.Plan(), cost, err
}
```

#### 2. **递归生成物理计划**
以 `LogicalLimit` 为例，在 `find_best_task.go:574` 开始：

```go
func findBestTask(super base.LogicalPlan, prop *property.PhysicalProperty) (bestTask base.Task, err error) {
    // 1. 检查缓存
    bestTask = p.GetTask(prop)
    if bestTask != nil {
        return bestTask, nil  // 命中缓存，直接返回
    }

    // 2. 枚举当前逻辑算子的所有物理实现
    physicalPlans, hintWorksWithProp, err := exhaustPhysicalPlans(p.Self(), newProp)

    // 3. 对每个物理算子
    for _, pp := range physicalPlans {
        // 3.1 递归获取子节点的最优物理计划
        childTasks, err = iteration(iterObj, pp, childTasks, prop)

        // 3.2 把子计划附加到当前物理算子
        curTask := pp.Attach2Task(childTasks...)

        // 3.3 计算代价并比较
        if curIsBetter, err := compareTaskCost(curTask, bestTask); curIsBetter {
            bestTask = curTask
        }
    }

    // 4. 缓存并返回最优方案
    p.StoreTask(prop, bestTask)
    return bestTask, nil
}
```

#### 3. **枚举物理算子**
对于 **LogicalSelection**，在 `physical_selection.go:54-79`：

```go
func ExhaustPhysicalPlans4LogicalSelection(p *logicalop.LogicalSelection,
    prop *property.PhysicalProperty) ([]base.PhysicalPlan, bool, error) {

    // 为不同的执行引擎生成物理算子
    ret := make([]base.PhysicalPlan, 0, len(newProps))

    for _, newProp := range newProps {
        // 创建 PhysicalSelection
        sel := PhysicalSelection{
            Conditions: p.Conditions,  // 复用逻辑算子的条件
        }.Init(p.SCtx(), p.StatsInfo().ScaleByExpectCnt(...), ...)

        ret = append(ret, sel)
    }

    return ret, true, nil
}
```

对于 **DataSource**，可能生成多种访问路径：
- `TableScan`：全表扫描
- `IndexScan`：索引扫描（如果 `age` 或 `city` 有索引）
- `IndexLookup`：索引回表

#### 4. **代价计算和选择**

假设 `city` 列有索引，`age` 列也有索引，系统会：

1. **枚举所有可能的物理计划：**
   ```
   方案A：全表扫描 + 过滤
   PhysicalLimit
     → PhysicalSort
       → PhysicalSelection(age>18, city='Beijing')
         → PhysicalTableScan(users)

   方案B：使用 city 索引
   PhysicalLimit
     → PhysicalSort(age DESC)
       → PhysicalSelection(age>18)
         → PhysicalIndexLookup(index=city_idx, filter=city='Beijing')

   方案C：使用 age 索引（已排序）
   PhysicalLimit
     → PhysicalSelection(city='Beijing')
       → PhysicalIndexLookup(index=age_idx, order=DESC)
   ```

2. **计算每个方案的代价：**
   ```go
   cost = rows * cpu_cost + io_cost + network_cost + ...
   ```

3. **选择代价最低的方案**（假设选方案C）

---

### 最终的物理计划

```
PhysicalLimit (count=10, cost=100)
    ↓
PhysicalSelection (condition=[city='Beijing'], cost=500)
    ↓
PhysicalIndexReader (cost=2000)
    ↓
PhysicalIndexLookup (
        index=age_idx,
        order=DESC,
        range=[18, +∞],
        cost=1800
    )
```

---

### 完整流程图

```
┌────────────────────────────────────────────────────────────────┐
│ SELECT name, age FROM users WHERE age>18 AND city='Beijing'   │
│ ORDER BY age DESC LIMIT 10                                     │
└────────────┬───────────────────────────────────────────────────┘
             │ Parser (解析器)
             ↓
┌────────────────────────────────────────────────────────────────┐
│                       AST 抽象语法树                            │
└────────────┬───────────────────────────────────────────────────┘
             │ PlanBuilder.Build() (logical_plan_builder.go)
             ↓
┌────────────────────────────────────────────────────────────────┐
│                    初始逻辑计划树                               │
│  LogicalLimit → LogicalSort → LogicalProjection                │
│    → LogicalSelection → DataSource                             │
└────────────┬───────────────────────────────────────────────────┘
             │ logicalOptimize() (optimizer.go:1040)
             │ - 列裁剪、谓词下推、常量传播...
             ↓
┌────────────────────────────────────────────────────────────────┐
│                   优化后的逻辑计划                              │
│  LogicalLimit → LogicalSort → DataSource(带下推条件)           │
└────────────┬───────────────────────────────────────────────────┘
             │ physicalOptimize() (optimizer.go:1082)
             │ logic.FindBestTask(prop)
             ↓
┌────────────────────────────────────────────────────────────────┐
│           递归枚举和代价计算 (find_best_task.go)                │
│  - exhaustPhysicalPlans(): 枚举物理算子                        │
│  - 递归处理子节点                                               │
│  - compareTaskCost(): 代价比较                                 │
│  - 选择最优方案                                                 │
└────────────┬───────────────────────────────────────────────────┘
             ↓
┌────────────────────────────────────────────────────────────────┐
│                      最终物理计划                               │
│  PhysicalLimit → PhysicalSelection                             │
│    → PhysicalIndexReader → PhysicalIndexLookup                 │
└────────────┬───────────────────────────────────────────────────┘
             │ postOptimize() (optimizer.go:458)
             │ - 投影消除、并行化等后处理
             ↓
┌────────────────────────────────────────────────────────────────┐
│                   可执行的物理计划                              │
│                    (发送给执行引擎)                             │
└────────────────────────────────────────────────────────────────┘
```

---

#### 关键代码位置总结

| 阶段 | 文件 | 关键函数 | 行数 |
|------|------|---------|------|
| 构建逻辑计划 | `logical_plan_builder.go` | `buildSelection()` | 911-939 |
| 逻辑优化 | `optimizer.go` | `logicalOptimize()` | 1040-1074 |
| 物理优化入口 | `optimizer.go` | `physicalOptimize()` | 1082-1119 |
| 递归生成物理计划 | `find_best_task.go` | `findBestTask()` | 574+ |
| 枚举物理算子 | `exhaust_physical_plans.go` | `exhaustPhysicalPlans()` | 54-96 |
| 物理算子实现 | `physical_selection.go` | `ExhaustPhysicalPlans4LogicalSelection()` | 54-79 |

这就是从 SQL 到逻辑计划再到物理计划的完整过程！本质上是一个**自顶向下的递归搜索**过程，通过**动态规划**（记忆化搜索）找到代价最低的执行方案。

## 总结

这篇文章中，我们先从AST看起，然后接着看了逻辑计划与物理计划是如何来的，顺带介绍了数据库引擎中很重要的火山模型，
以及我们对最基本的模型的优缺点分析、优化方案。
