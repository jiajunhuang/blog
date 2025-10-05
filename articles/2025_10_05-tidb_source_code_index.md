# TiDB 源码阅读（五）：索引

## 索引的相关概念

数据库索引可以按照多种维度进行分类。

### 1. 按数据结构分类

这是最根本的分类方式，因为它直接决定了索引的效率和适用场景。

| 索引类型 | 数据结构 | 工作原理 | 优点 | 缺点 | 常见数据库支持 |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **B+树索引** | B+树 | 多路平衡搜索树。所有数据都存储在叶子节点，并形成有序链表。非叶子节点只存储键值。 | **范围查询效率高**（叶子节点链表）、查询稳定（每次查询深度相同）、磁盘IO少（节点大小通常设置为页的整数倍）。 | 写入时需要维护树结构平衡，有一定开销。 | **MySQL（InnoDB）、PostgreSQL、Oracle、SQL Server** 的默认/主索引类型。 |
| **哈希索引** | 哈希表 | 对索引键计算哈希码，直接映射到数据行的物理地址。 | **等值查询极快**（O(1)时间复杂度）。 | **完全无法支持范围查询**、不支持排序、哈希冲突问题。 | MySQL的Memory引擎、PostgreSQL（可创建）、Redis。 |
| **全文索引** | 倒排索引 | 记录单词（或词组）到其所在文档/记录的映射。 | **专为全文搜索设计**，支持关键词匹配、相关性排序、模糊查询等。 | 占用空间大，维护成本高。 | MySQL（MyISAM/InnoDB）、PostgreSQL（GIN）、Elasticsearch（核心）。 |
| **R-Tree索引** | R树 | 专门用于存储空间数据（如点、线、面）的平衡树。 | **高效支持空间查询**，如“查找附近的所有点”、“查找这个矩形内的所有多边形”。 | 仅适用于空间数据类型。 | MySQL、PostgreSQL（PostGIS）、Oracle。 |
| **Trie树索引** | 字典树 | 一种专门处理字符串的前缀树。 | **前缀查询效率极高**，例如查询所有以“abc”开头的字符串。 | 内存消耗可能较大。 | 较少直接用于磁盘数据库，更多见于内存数据库或特定场景。 |

### 2. 按存储内容（或与数据的耦合程度）分类

这个分类主要针对B+树索引，描述了索引节点中存储了什么。

| 索引类型 | 存储内容 | 特点 | 例子 |
| :--- | :--- | :--- | :--- |
| **聚集索引** | **索引的叶子节点直接存储完整的数据行**。 | 1. **一个表只能有一个聚集索引**，因为数据行本身只能按一种顺序物理存储。<br>2. 由于数据行就在叶子节点，通过聚集索引访问数据非常高效。 | **InnoDB**的表必须有且只有一个聚集索引：<br>-- 如果定义了主键，主键就是聚集索引。<br>-- 没有主键，则第一个UNIQUE非空索引是聚集索引。<br>-- 都没有，则隐式创建一个ROWID作为聚集索引。 |
| **非聚集索引** | **索引的叶子节点存储的是索引键值和指向数据行的“指针”**。 | 1. **一个表可以有多个非聚集索引**。<br>2. 查询数据需要一次“回表”操作：先通过非聚集索引找到主键，再通过主键（聚集索引）去查找完整数据行。 | 在InnoDB中，除了聚集索引，其他所有索引都是非聚集索引，其叶子节点存储的是**主键值**。在MyISAM中，主键索引和其他索引都是非聚集索引，叶子节点存储的是**数据的物理地址**。 |

> 聚集索引的英文是 Clustered Index；在MySQL中，非聚集索引，也叫次级索引(Secondary Index)。

**回表示例（InnoDB）**：
假设`users`表，`id`是主键（聚集索引），我们在`name`上建立了非聚集索引。
执行 `SELECT * FROM users WHERE name = ‘Alice’;`
1.  在`name`索引的B+树中找到`‘Alice’`，得到对应的主键`id`（比如 10）。
2.  再用这个`id=10`去主键（聚集索引）的B+树中查找，最终拿到完整的数据行。
第二步就是“回表”，如果查询只涉及索引列，则可以避免回表。

### 3. 按索引字段的数量和特性分类

| 索引类型 | 字段数量 | 描述 | 示例SQL（MySQL） |
| :--- | :--- | :--- | :--- |
| **单列索引** | 1 | 只基于一个表字段创建的索引。 | `CREATE INDEX idx_name ON users(name);` |
| **组合索引** | ≥2 | 基于多个表字段创建的索引，也叫复合索引。 | `CREATE INDEX idx_name_age ON users(name, age);` |
| **覆盖索引** | - | **这是一个特优情况，而非一种索引类型**。如果一个索引包含了查询所需要的所有字段，则查询无需回表，直接从索引中获取数据，这个索引就称为覆盖索引。 | `CREATE INDEX idx_covering ON users(name, age);` <br> `SELECT name, age FROM users WHERE name = ‘Alice’;` <-- 这个查询就使用了覆盖索引。 |

> 覆盖索引，不能单独算作索引的一种类别，只是恰好查询所涉及到的所有字段都在索引中，无需回表即可完成查询。

**组合索引的最左前缀原则**：
对于组合索引 `(A, B, C)`，它相当于建立了 `(A)`, `(A, B)`, `(A, B, C)` 三个索引。查询时必须**从最左列开始**匹配，否则索引失效。
- 有效：`WHERE A=1`, `WHERE A=1 AND B=2`, `WHERE A=1 AND B=2 AND C=3`
- **无效**：`WHERE B=2`, `WHERE C=3`, `WHERE B=2 AND C=3`

### 4. 按数据的唯一性分类

| 索引类型 | 约束 | 特点 |
| :--- | :--- | :--- |
| **唯一索引** | 索引键值必须唯一，不允许重复。 | 1. 一个表可以有多个唯一索引。<br>2. 唯一索引可以保证数据的唯一性，是数据库的一种约束。<br>3. 主键是一种特殊的唯一索引（不允许NULL）。 |
| **普通索引** | 无唯一性约束。 | 允许重复值和NULL值（取决于字段定义）。 |

索引不是越多越好，它是以**空间换时间**的策略。每个索引都会降低**INSERT、UPDATE、DELETE**的速度，因为数据变更时需要同时维护所有相关的索引结构。因此，创建索引需要根据实际的查询需求进行权衡和设计。

## 谓词下推

谓词下推的作用，就是在数据库查询处理过程中，尽早地执行数据过滤操作，以减少后续操作需要处理的数据量，从而极大提升查询性能。

通过谓词下推，可以提前过滤掉无用的数据，减少了数据的传输量，减少IO，节省CPU和内存的使用，最终提升查询速度。

### 1. **逻辑优化阶段的主入口**

谓词下推作为一条逻辑优化规则，在 `pkg/planner/core/optimizer.go` 的第 **89-117** 行定义的 `optRuleList` 中：

```go:89:117:pkg/planner/core/optimizer.go
var optRuleList = []base.LogicalOptRule{
	&GcSubstituter{},
	&rule.ColumnPruner{},
	&ResultReorder{},
	&rule.BuildKeySolver{},
	&DecorrelateSolver{},
	&SemiJoinRewriter{},
	&AggregationEliminator{},
	&SkewDistinctAggRewriter{},
	&ProjectionEliminator{},
	&MaxMinEliminator{},
	&rule.ConstantPropagationSolver{},
	&ConvertOuterToInnerJoin{},
	&PPDSolver{},                           // ← 谓词下推规则在这里
	&OuterJoinEliminator{},
	&rule.PartitionProcessor{},
	&rule.CollectPredicateColumnsPoint{},
	&AggregationPushDownSolver{},
	&DeriveTopNFromWindow{},
	&rule.PredicateSimplification{},
	&PushDownTopNOptimizer{},
	&rule.SyncWaitStatsLoadPoint{},
	&JoinReOrderSolver{},
	&rule.ColumnPruner{},
	&PushDownSequenceSolver{},
	&EliminateUnionAllDualItem{},
	&EmptySelectionEliminator{},
	&ResolveExpand{},
}
```

### 2. **PPDSolver 规则实现**

在 `pkg/planner/core/rule_predicate_push_down.go` 中：

```42:47:pkg/planner/core/rule_predicate_push_down.go
func (*PPDSolver) Optimize(_ context.Context, lp base.LogicalPlan) (base.LogicalPlan, bool, error) {
	planChanged := false
	_, p, err := lp.PredicatePushDown(nil)
	return p, planChanged, err
}
```

这个规则从根节点开始递归调用每个逻辑算子的 `PredicatePushDown` 方法。

### 3. **各个逻辑算子的谓词下推实现**

每个逻辑算子都实现了 `PredicatePushDown` 接口方法：

#### **LogicalSelection**（选择算子）
在 `pkg/planner/core/operator/logicalop/logical_selection.go` 中：
- 将谓词条件继续向下推到子节点
- 不能下推的条件保留在当前 Selection 节点

#### **DataSource**（数据源）
在 `pkg/planner/core/operator/logicalop/logical_datasource.go` 第 **156-168** 行：

```156:168:pkg/planner/core/operator/logicalop/logical_datasource.go
// PredicatePushDown implements base.LogicalPlan.<1st> interface.
func (ds *DataSource) PredicatePushDown(predicates []expression.Expression) ([]expression.Expression, base.LogicalPlan, error) {
	predicates = ruleutil.ApplyPredicateSimplification(ds.SCtx(), predicates, true, nil)
	// Add tidb_shard() prefix to the condtion for shard index in some scenarios
	// TODO: remove it to the place building logical plan
	predicates = utilfuncp.AddPrefix4ShardIndexes(ds, ds.SCtx(), predicates)
	ds.AllConds = predicates
	dual := Conds2TableDual(ds, ds.AllConds)
	if dual != nil {
		return nil, dual, nil
	}
	ds.PushedDownConds, predicates = expression.PushDownExprs(util.GetPushDownCtx(ds.SCtx()), predicates, kv.UnSpecified)
	return predicates, ds, nil
}
```

在这里，谓词会被尽可能推到存储层（TiKV/TiFlash）。

#### **LogicalJoin**（连接算子）
在 `pkg/planner/core/operator/logicalop/logical_join.go` 中：
- 根据 JOIN 类型（Inner/Left/Right/Outer）决定哪些谓词可以推到左右子树
- 处理 ON 条件和 WHERE 条件

### 4. **物理计划阶段的谓词下推**

还有针对 TiFlash 的特殊优化，在 `pkg/planner/core/operator/physicalop/tiflash_predicate_push_down.go` 中：
- 对于 TiFlash 扫描，有专门的谓词下推逻辑
- 考虑选择性、代价等因素决定是否下推

## TiDB 中是如何使用索引的？

在逻辑计划构建阶段，getPossibleAccessPaths 函数会为每个表生成所有可能的访问路径：

```go
func getPossibleAccessPaths(ctx base.PlanContext, tableHints *hint.PlanHints, indexHints []*ast.IndexHint, tbl table.Table, dbName, tblName ast.CIStr, check bool, hasFlagPartitionProcessor bool) ([]*util.AccessPath, error) {
	tblInfo := tbl.Meta()
	publicPaths := make([]*util.AccessPath, 0, len(tblInfo.Indices)+2)
	tp := kv.TiKV
	if tbl.Type().IsClusterTable() {
		tp = kv.TiDB
	}
	tablePath := &util.AccessPath{StoreType: tp}
	fillContentForTablePath(tablePath, tblInfo)
	publicPaths = append(publicPaths, tablePath)
    // ...
}
```

AccessPath 结构体定义了一条访问路径的所有信息：

```go
// AccessPath indicates the way we access a table: by using single index, or by using multiple indexes,
// or just by using table scan.
type AccessPath struct {
	Index          *model.IndexInfo
	FullIdxCols    []*expression.Column
	FullIdxColLens []int
	IdxCols        []*expression.Column
	IdxColLens     []int
	// ConstCols indicates whether the column is constant under the given conditions for all index columns.
	ConstCols []bool
	Ranges    []*ranger.Range
	// CountAfterAccess is the row count after we apply range seek and before we use other filter to filter data.
	// For index merge path, CountAfterAccess is the row count after partial paths and before we apply table filters.
	CountAfterAccess float64
	// MinCountAfterAccess is a lower bound on CountAfterAccess, accounting for risks that could
	// lead to overestimation, such as assuming correlation with exponential backoff when columns are actually independent.
	// Case MinCountAfterAccess > 0 : we've encountered risky scenarios and have a potential lower row count estimation
	// Default MinCountAfterAccess = 0 : we have not identified risks that could lead to lower row count
	MinCountAfterAccess float64
	// MaxCountAfterAccess is an upper bound on the CountAfterAccess, accounting for risks that could
	// lead to underestimation, such as assuming independence between non-index columns.
	// Case MaxCountAfterAccess > 0 : we've encountered risky scenarios and have a potential greater row count estimation
	// Default MaxCountAfterAccess = 0 : we have not identified risks that could lead to greater row count
	MaxCountAfterAccess float64
	// CountAfterIndex is the row count after we apply filters on index and before we apply the table filters.
	CountAfterIndex float64
	AccessConds     []expression.Expression
	EqCondCount     int
	EqOrInCondCount int
	IndexFilters    []expression.Expression
	TableFilters    []expression.Expression
```

之后做的事情，就是要从多个访问路径中找出最好的那条，其中用到了一种算法叫做“天际线剪枝”，用于从多个可用的路径中过滤掉
明显不好用的索引，为之后的访问路径筛选减少候选者。

### 天际线剪枝(Skyline Pruning)

#### 1. 为什么需要 Skyline Pruning？

在一个复杂的查询中（尤其是涉及多个谓词和排序的查询），数据库系统可能会找到许多个“可能有用”的索引。例如：

- 一个在 `(a)` 上的索引
- 一个在 `(a, b)` 上的索引
- 一个在 `(b, a)` 上的索引
- 一个在 `(a, b, c)` 上的索引

如果对每一个候选索引都进行详细的代价估算（这涉及到 I/O 成本、CPU 成本等计算，本身比较昂贵），优化器会花费很多时间。Skyline Pruning 的目标就是在进入代价估算阶段之前，先淘汰掉那些明显没有竞争力的索引。

---

#### 2. Skyline 的概念来源

“Skyline” 这个词来源于多目标优化问题。想象一个城市的天际线，那些摩天大楼就是“Skyline points”，因为它们至少在某个方向上（比如高度、位置）没有被其他建筑完全超越。

在索引选择的上下文中，我们也是在多个“维度”上比较索引，例如：

- **查询谓词的匹配度**：索引能覆盖多少个 `WHERE` 子句中的列？
- **排序属性**：索引是否能避免排序操作（满足 `ORDER BY` 或 `GROUP BY`）？
- **覆盖查询**：索引是否包含查询所需的所有列，从而无需回表？

一个索引的“天际线”就是那些在所有这些维度上，**不被其他任何索引完全支配**的索引集合。

---

#### 3. Skyline Pruning 如何工作？

我们通过一个具体的例子来说明。

**查询：**
```sql
SELECT * FROM orders
WHERE customer_id = 123 AND order_date > ‘2023-01-01’
ORDER BY order_date DESC;
```

**可用索引：**
1. `idx_customer` (`customer_id`)
2. `idx_date` (`order_date`)
3. `idx_customer_date` (`customer_id`, `order_date`)
4. `idx_date_customer` (`order_date`, `customer_id`)
5. `idx_customer_status` (`customer_id`, `status`) -- 一个不相关的索引

现在我们用 Skyline Pruning 来分析：

**维度 1：等值匹配列（最左前缀匹配）**
- 查询有 `customer_id = 123` 这个等值谓词。
- `idx_customer`: 匹配 1 个等值列。
- `idx_customer_date`: 匹配 1 个等值列 (`customer_id`)。
- `idx_date_customer`: 匹配 0 个等值列（因为最左列 `order_date` 是范围查询）。
- `idx_date`: 匹配 0 个等值列。
- `idx_customer_status`: 匹配 1 个等值列。

**维度 2：范围匹配列**
- 查询有 `order_date > ‘2023-01-01’` 这个范围谓词。
- `idx_customer_date`: 在 `customer_id` 等值匹配后，可以匹配范围列 `order_date`。
- `idx_date_customer`: 可以直接匹配范围列 `order_date`。
- `idx_date`: 可以直接匹配范围列 `order_date`。
- `idx_customer` 和 `idx_customer_status`：无法匹配这个范围列。

**维度 3：排序属性**
- 查询需要按 `order_date DESC` 排序。
- `idx_customer_date`: 由于数据先按 `customer_id` 排序，再按 `order_date` 排序，对于一个固定的 `customer_id`，其内部的 `order_date` 是有序的。所以它可以避免排序。
- `idx_date_customer`: 数据首先按 `order_date` 排序，这正好满足 `ORDER BY order_date`，所以它也可以完美避免排序。
- 其他索引：都无法避免排序。

**开始剪枝：**

1. **比较 `idx_customer` 和 `idx_customer_date`**:
   - 在等值匹配列上，两者打平（都是1）。
   - 在范围匹配列上，`idx_customer_date` 胜出（它能匹配，而 `idx_customer` 不能）。
   - 在排序属性上，`idx_customer_date` 胜出（它能避免排序，而 `idx_customer` 不能）。
   - **结论**：`idx_customer` 在**所有维度**上都被 `idx_customer_date` **支配**。因此，`idx_customer` 被剪枝。

2. **比较 `idx_date` 和 `idx_date_customer`**:
   - 在等值匹配列上，`idx_date_customer` 胜出（它能利用 `customer_id` 进行过滤，虽然是第二列）。
   - 在范围匹配列上，两者打平（都能匹配）。
   - 在排序属性上，两者打平（都能完美避免排序）。
   - **结论**：`idx_date` 在至少一个维度上（等值匹配）比 `idx_date_customer` 差，且没有在任何维度上更好。因此，`idx_date` 被 `idx_date_customer` **支配**，被剪枝。

3. **`idx_customer_status`**:
   - 它在任何维度上都不如 `idx_customer_date`，所以也会被剪枝。

**经过 Skyline Pruning 后，剩下的候选索引是：**
- `idx_customer_date`
- `idx_date_customer`

现在，查询优化器只需要对这两个“精英”索引进行详细的代价估算，以选出最终的执行计划。这大大减少了优化时间。

## 基于成本的索引选择

天际线剪枝算法之后，可能仍然会有多个访问路径，此时，就会用基于成本(用统计数据)的选择来选出最佳访问路径，这也是
为什么TiDB在导入数据完成之后要做ANALYZE TABLE之后，才能获得最佳效果的原因。

```go
func physicalOptimize(logic base.LogicalPlan) (plan base.PhysicalPlan, cost float64, err error) {
    // ...
    t, err := logic.FindBestTask(prop)
    // ...
}
```

不过此处过于专业，我也没全搞明白，就不展开了(问了很多AI，但是从来没搞过这块，大家有兴趣可以问问AI，我就不贴了)。

## 总结

本文中，我们总结了一下索引的类型，然后学习了谓词下推，通过谓词下推，可以提前把过滤条件下放到最底端，从而减少
产生的数据；接着我们学习了在有多条访问路径的时候，数据库是如何通过天际线剪枝，过滤掉明显不合格的路径，最终
通过基于成本的选择，产生访问路径。此外我们还有一点没有介绍到，就是索引合并(index merge)技术，通过索引合并，
可以在没有明显优势的单个索引，但是有多个索引且过滤条件为 `AND` 或者 `OR` 的情况下用上索引，提高访问效率，不过
本文没有详细介绍这点。
