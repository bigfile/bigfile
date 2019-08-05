//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

// Package service define many components for implementing
// system. Every service completes some specify function.
package service

import (
	"context"

	"github.com/jinzhu/gorm"
)

// BeforeHandler is used to add handle function that is executed before
// Execute.
type BeforeHandler = func(ctx context.Context, service Service) error

// AfterHandler is used to add handle function that is executed after
// Execute.
type AfterHandler = func(ctx context.Context, service Service) error

// Service interface conventions all subtypes must implement the Execute
// method.
type Service interface {

	// Execute is designed to implement specific function
	Execute(ctx context.Context) error

	// Validate is designed to validate input params
	Validate() error

	// for convenience, each service should define a return value
	// method that returns the specific output of service.
}

// BaseService only includes two fields: Before and After, the handler in
// Before will be executed in front of Execute and in After will be executed
// in back of Execute
type BaseService struct {

	// Before includes many BeforeHandler
	Before []BeforeHandler

	// After is consists of many AfterHandler
	After []AfterHandler

	// DB represent a database connection
	DB *gorm.DB
}
