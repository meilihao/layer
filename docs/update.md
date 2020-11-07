# update

参考[examples/update.go](/examples/update.go)

update删除时必须要有where条件, 否则报ErrMissingWhereClause. where条件来源:
1. `(*UpdateSession) Where()`
1. schema pks.  不指定`(*UpdateSession) NoPK()`时, layer自动追加到`(*UpdateSession) Where()`; 指定时Where会忽略pk
1. schema version. 不指定`(*UpdateSession) NoVersion()`时, layer自动追加到`(*UpdateSession) Where()`; 指定时Where会忽略version

## 用指定的字段更新记录
默认更新除created_at, deleted_at, pk外的所有字段, 包括零值字段

配合`NewUpdateSession()`, 有3种更新指定列的方法(方法间互斥使用):
1. `Select("Name", "Age")`用指定列进行update
1. `Omit("Name", "Age")`排除指定列进行update
1. `Set(map[string]interface{"Age":10})`指定`列+值`进行update

排除auto update column:
- `(*UpdateSession) NoAutoVersion()` : 不更新version
- `(*UpdateSession) NoAutoUpdatedAt()` : 不更新updated_at