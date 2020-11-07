# layer

对`database/sql`进行适度封装,提供一些便捷函数.

特别说明:
1. 个人暂定不会实现DML(insert, update, delete)和DQL(select)外的语句, 但可自行在`dialect/dialect.go#Dialecter`中扩展
1. 暂定不实现orm的表关联

## doc
- [doc](/docs)
- [api doc](https://godoc.org/github.com/meilihao/layer)

## Roadmap

- [x] 实现crud

## code style
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)

## code ref
- [huge](https://github.com/cxr29/huge)
- [gorm](https://github.com/go-gorm/gorm)

next: merge [builder](https://github.com/didi/gendry/tree/master/builder) by [golang使用mysql实例和第三方库Gendry](https://zhuanlan.zhihu.com/p/140084175)

## license

MIT
