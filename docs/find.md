# find
参考[examples/find.go](/examples/find.go)

支持`&struct`, slice, map查询记录

find时必须要有where条件, 否则报ErrMissingWhereClause. where条件来源:
1. `(*FindSession) Where()`
1. schema pks.  不指定`(*FindSession) NoPK()`时, layer自动追加到`(*FindSession) Where()`; 指定时Where会忽略pk
1. schema version. 不指定`(*FindSession) NoVersion()`时, layer自动追加到`(*FindSession) Where()`; 指定时Where会忽略version
1. schema deleted_at. 不指定`(*FindSession) Unscoped()`时, layer自动追加到`(*FindSession) Where()`; 指定时Where会忽略deleted_at

## 用指定的字段查询记录

配合`NewFindSession()`, 有2种查询指定列的方法(方法间互斥使用):
1. `Select("Name", "Age")`用指定列进行update
1. `Omit("Name", "Age")`排除指定列进行update

## 特定方法
- Distinct() : distinct