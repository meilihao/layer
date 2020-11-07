# create
参考[examples/create.go](/examples/create.go)

支持`&struct`, slice, map创建记录

## 用指定的字段创建记录
配合`NewCreateSession()`, 支持`Select("Name", "Age", "CreatedAt")`用指定列进行insert 或 `Omit("Name", "Age", "CreatedAt")`排除指定列进行insert.

`KeepAutoIncr()`支持insert时保留自增列, 便于指定id进行插入的场景; 默认情况下, insert不包含自增列.