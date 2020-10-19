# 从GORM里学习到的panic处理方式

今天在博客的评论里，有童鞋提醒我，GORM里也有简化事务处理的帮助函数。源码如下：

```go
// Transaction start a transaction as a block, return error will rollback, otherwise to commit.
func (db *DB) Transaction(fc func(tx *DB) error, opts ...*sql.TxOptions) (err error) {
	panicked := true

	if committer, ok := db.Statement.ConnPool.(TxCommitter); ok && committer != nil {
		// nested transaction
		db.SavePoint(fmt.Sprintf("sp%p", fc))
		defer func() {
			// Make sure to rollback when panic, Block error or Commit error
			if panicked || err != nil {
				db.RollbackTo(fmt.Sprintf("sp%p", fc))
			}
		}()

		err = fc(db.Session(&Session{WithConditions: true}))
	} else {
		tx := db.Begin(opts...)

		defer func() {
			// Make sure to rollback when panic, Block error or Commit error
			if panicked || err != nil {
				tx.Rollback()
			}
		}()

		err = fc(tx)

		if err == nil {
			err = tx.Commit().Error
		}
	}

	panicked = false
	return
}
```

思路和 [我的](https://jiajunhuang.com/articles/2020_10_17-golang_db_transaction.md.html) 差不多。

有两个不同，第一，在获取了committer 的时候，会优先选择使用savepoint这个特性，相当于事务里的子事务。

第二，处理panic的方式，这一点值得学习。首先在函数的入口处设置变量：`panicked := true`，在 `defer` 函数中判断是否panic了，从而进行相应处理：

```go
		defer func() {
			// Make sure to rollback when panic, Block error or Commit error
			if panicked || err != nil {
				db.RollbackTo(fmt.Sprintf("sp%p", fc))
			}
		}()
```

那么什么情况下，不会执行呢？当然就是执行到最后的时候，执行了 `panicked = false` 这一行代码之后。

这样就成功避免了使用 `recover` 来判断是否发生了异常。不过也因此，无法在这一层捕捉 `panic` 了。

---

ref:

- https://github.com/go-gorm/gorm/blob/9b2181199d88ed6f74650d73fa9d20264dd134c0/finisher_api.go#L418
