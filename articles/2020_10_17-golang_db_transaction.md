# Go使用闭包简化数据库操作代码

在日常工程中，我们可能要开事务来完成一些操作，因此就会有如下代码：

```go
// UpdateTaskRemark 更新任务备注
func UpdateTaskRemark(taskID uint32, remark string) error {
        sql, args, err := sq.Update("tasks").Set("remark", remark).Set("updated_at", getNowTS()).Where("id = ?", taskID).ToSql()
        if err != nil {
                return err
        }

        tx, err := db.Beginx()
        if err != nil {
                return err
        }

        _, err = tx.Exec(sql, args...)
        if err != nil {
                return err
        }

        err = tx.Commit()
        if err != nil {
                logrus.Errorf("failed to commit transaction: %s, rollback it", err)
                return tx.Rollback()
        }

        return nil
}
```

如果只有一个，似乎没有什么问题，但是，一旦操作多了之后，要在每个model操作函数中都写上这一堆重复的代码，那就不太妙了。
因此得想一个办法来进行抽象，把开事务、提交和回滚的操作封装起来。借助闭包我们可以做到这一点：

```go
// 工具函数用于在事务中执行，自动提交和回滚
type TxFunc func(tx *sqlx.Tx) error

// ExecInTx 传入一个闭包函数，签名类型为TxFunc。在ExecInTx中先开事务，
// 然后将事务传入闭包函数，如果闭包函数没有返回错误就提交，否则回滚。
func ExecInTx(f TxFunc) error {
        tx, err := db.Beginx()
        if err != nil {
                return err
        }

        err = f(tx)
        if err != nil {
                return err
        }

        err = tx.Commit()
        if err != nil {
                logrus.Errorf("failed to commit transaction: %s, rollback it", err)
                return tx.Rollback()
        }

        return nil
}
```

由此，model操作函数就简化成了：

```go
// UpdateTaskRemark 更新任务备注
func UpdateTaskRemark(taskID uint32, remark string) error {
        sql, args, err := sq.Update("tasks").Set("remark", remark).Set("updated_at", getNowTS()).Where("id = ?", taskID).ToSql()
        if err != nil {
                return err
        }

        return ExecInTx(func(tx *sqlx.Tx) error {
                _, err := tx.Exec(sql, args...)
                return err
        })
}
```

如上。
