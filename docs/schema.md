# schema

## 定义
模型是标准的 struct，由 Go 的基本数据类型、实现了 [Scanner](https://pkg.go.dev/database/sql/?tab=doc#Scanner) 和 [Valuer](https://pkg.go.dev/database/sql/driver#Valuer) 接口的自定义类型及其指针组成.

例如:
```go
type Node struct {
	Id        int `layer:";pk;autoincr"`
	Name      string
	Data      map[string]interface{} `layer:";json"`
	Parent    *Node                  `layer:";many2one"`
	Children  []*Node                `layer:";one2many"`
	Siblings  map[int]*Node          `layer:";many2many"`
	Version   int                    `layer:";version"`
	CreatedAt int64                  `layer:";created_at"`
	UpdatedAt time.Time              `layer:";updated_at"`
	DeletedAt *time.Time             `layer:";deleted_at"`
	*Point    `layer:";embedded=Point"`
}

type Point struct {
	X int `layer:";pk"`
	Y int `layer:";pk"`
}
```

不支持struct collapse
```go
// struct collapse
type Circle struct {
  X, Y, Z int64 `xxx:"..."`
}
```

## 约定
layer schema 倾向于配置，而不是约定, 但默认情况下已提供足够使用的默认配置:
1. 使用结构体名的 蛇形 作为表名，字段名的 蛇形 作为列名，自定义命名可通过 `WithNameMpaper`实现
1. 表名默认使用单数形式, 可通过`Tabler`实现自定义

> 更多配置见[options.go](/options.go)

注意:
1. 不支持data description language (DDL), 这类语句与db关联紧密, 不适合orm来实现

## column
在 field 对应的 Tag 中对 Column 的一些属性进行定义, 并以`;`分隔, 以及部分field支持以`k=v`形式扩展tag含义，定义的方法基本和写SQL定义表结构类似，可参考上面的定义例子.

具体的 Tag 规则如下，另 Tag 中的关键字区分大小写：

<table>
    <tr>
        <td>name</td><td>当前field对应的字段的名称, 可选. 通常不写，此时layer会自动取field名字. 填写时请使用NameMpaper前的名称, 可避免使用自定义`WithNameMpaper`时需要修改struct tag. name不支持`k=v`形式, 且必须写在第一个</td>
    </tr>
    <tr>
        <td>pk</td><td>是否是Primary Key，支持复合主键</td>
    </tr>
    <tr>
        <td>autoincr</td><td>是否是自增, 自增字段必须是pk, 且至多一个</td>
    </tr>
    <tr>
        <td>notnull</td><td>是否可以为空</td>
    </tr>
    <tr>
        <td>unique</td><td>是否是唯一, 目前仅用于展示</td>
    </tr>
    <tr>
        <td>index</td><td>是否有索引, 目前仅用于展示</td>
    </tr>
     <tr>
        <td>created_at</td><td>这个Field将在Insert时自动赋值为当前时间. 支持使用 nano/milli 来实现纳秒、毫秒时间精度(需数据库支持), 至多一个</td>
    </tr>
     <tr>
        <td>updated_at</td><td>这个Field将在Insert或Update时自动赋值为当前时间. 支持使用 nano/milli 来实现纳秒、毫秒时间精度(需数据库支持), 至多一个</td>
    </tr>
    <tr>
        <td>deleted_at</td><td>这个Field将在Delete时设置为当前时间，并且当前记录不删除(软删除). 支持使用 nano/milli 来实现纳秒、毫秒时间精度(需数据库支持), 至多一个, **推荐该字段使用`sql.NullTime`类型**</td>
    </tr>
    <tr>
        <td>version</td><td>乐观锁, Insert有非零值时会保留该值, 否则初始值为1</td>
    </tr>
    <tr>
        <td>embedded</td><td>嵌套字段, 支持指定嵌套前缀</td>
    </tr>
    <tr>
        <td>-</td><td>这个Field将不进行字段映射</td>
    </tr>
    <tr>
        <td>json</td><td>会将内容转成xml格式，然后存储到数据库中，需db支持</td>
    </tr>
    <tr>
        <td>xml</td><td>会将内容转成xml格式，然后存储到数据库中，需db支持</td>
    </tr>
    <tr>
        <td>one2one</td><td>1对1关联</td>
    </tr>
    <tr>
        <td>many2one</td><td>多对1关联, 即外键关联</td>
    </tr>
    <tr>
        <td>one2many</td><td>1对多关联</td>
    </tr>
    <tr>
        <td>many2many</td><td>多对多关联</td>
    </tr>
    <tr>
        <td>type</td><td>字段类型, 目前仅用于展示</td>
    </tr>
    <tr>
        <td>default</td><td>默认值, 目前仅用于展示</td>
    </tr>
    <tr>
        <td>size</td><td>长度, 目前仅用于展示</td>
    </tr>
    <tr>
        <td>precision</td><td>精度, 目前仅用于展示</td>
    </tr>
    <tr>
        <td>comment</td><td>字段的注释, 目前仅用于展示</td>
    </tr>
</table>