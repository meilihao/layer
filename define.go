// Copyright (c) 2017 github.com/meilihao. All rights reserved.

package layer

import (
	"database/sql"
	"errors"
)

var (
	// std
	ErrNoRows = sql.ErrNoRows
	ErrTxDone = sql.ErrTxDone

	// custom
	ErrUnsupportedDriverName = errors.New("layer : Unsupported DriverName")
	ErrEmptyDataSource       = errors.New("layer : Invalid DataSource")
	ErrModelNeedPK           = errors.New("layer : Model Needs At Least One Primary Key")
	ErrNil                   = errors.New("layer : nil")
	ErrUsingNonPtrModelData  = errors.New("layer : use non-ptr model data")
	ErrUsingNilPtrModelData  = errors.New("layer : nil ptr model data")
	ErrUsingNotStructModel   = errors.New("layer : not struct model")
)
