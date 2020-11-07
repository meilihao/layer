//常见命名法:
//- 蛇形命名法(snake_case) : like_this, 常见于Linux内核, C++标准库, Boost以及Ruby, Rust等语言
//- 驼峰命名法(camelCase]) : likeThis, 为了和帕斯卡命名法区分, 本文特指小驼峰式命名法, 常见于Qt以及Java
//- 帕斯卡命名法(PascalCase) : LikeThis, 又名大驼峰式命名法, 常见于Windows API函数以及C#
//- 匈牙利命名法(Hungarian notation) : 基本原则是变量名=属性+类型+对象描述
package schema

// other name mapper use Strcut.TableName(),Session.Table() or strcut tag "column"
type NameMapper interface {
	EntityMap(string) string        // struct/filed -> table/colume
	EntityMapReverse(string) string // table/colume -> struct/field
}

type SameNameMapper struct{} // as PascalCase

func (m SameNameMapper) EntityMap(name string) string {
	return name
}

func (m SameNameMapper) EntityMapReverse(name string) string {
	return name
}

type SnakeNameMapper struct{}

func (m SnakeNameMapper) EntityMap(name string) string {
	newName := make([]rune, 0)

	for i, chr := range name {
		if 'A' <= chr && chr <= 'Z' { // is upper
			if i > 0 {
				newName = append(newName, '_')
			}
			chr += 32 // 'a' - 'A' = 32
		}
		newName = append(newName, chr)
	}

	return string(newName)
}

func (m SnakeNameMapper) EntityMapReverse(name string) string {
	newName := make([]rune, 0)
	upNextChar := true // 下一个字符是否需要大写

	// name = strings.ToLower(name) // 默认传入的名称是合法的

	for _, chr := range name {
		switch {
		case upNextChar:
			upNextChar = false
			if 'a' <= chr && chr <= 'z' {
				chr -= 32
			}
		case chr == '_':
			upNextChar = true
			continue
		}

		newName = append(newName, chr)
	}

	return string(newName)
}
