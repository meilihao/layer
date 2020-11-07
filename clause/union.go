package clause

import "errors"

var (
	ClauseUnion   = "UNION"
	ErrEmptyUnion = errors.New("empty union")
)

type UnionType string

const (
	UnionNil      UnionType = ""
	Union         UnionType = "UNION"
	UnionAll      UnionType = "UNION ALL"
	UnionDistinct UnionType = "UNION DISTINCT"
)
