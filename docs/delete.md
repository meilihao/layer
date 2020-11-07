# delete

参考[examples/delete.go](/examples/delete.go)

delete删除时必须要有where条件, 否则报ErrMissingWhereClause. where条件来源:
1. `(*DeleteSession) Where()`
1. schema pks.  不指定`(*DeleteSession) NoPK()`时, layer自动追加到`(*DeleteSession) Where()`; 指定时Where会忽略pk
1. schema version. 不指定`(*DeleteSession) NoVersion()`时, layer自动追加到`(*DeleteSession) Where()`; 指定时Where会忽略version

## 软删除
如果schema包含了一个tag为`deleted_at`的字段, 则layer将自动获得软删除的能力！

拥有软删除能力的schema调用 Delete 时，记录不会在数据库中真正删除, 而是仅将该字段置为当前时间， 并且不能再通过正常的查询方法找到该记录, 而是查询时需要追加`Unscoped()`方法.

`l.NewDeleteSession().Unscoped()`的`Unscoped()`会禁用软删除, 从而真正删除记录.