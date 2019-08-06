//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBaseService_CallBefore(t *testing.T) {
	var (
		baseService = &BaseService{
			Before: []BeforeHandler{
				func(ctx context.Context, service Service) error {
					b := service.(*BaseService)
					b.Value["start"] = time.Now()
					return nil
				},
			},
			After: []AfterHandler{
				func(ctx context.Context, service Service) error {
					b := service.(*BaseService)
					b.Value["end"] = time.Now().Add(1 * time.Second)
					return nil
				},
			},
			DB:    nil,
			Value: make(map[string]interface{}),
		}
		err error
	)
	err = baseService.Execute(context.TODO())
	assert.Nil(t, err)
	start := baseService.Value["start"].(time.Time)
	end := baseService.Value["end"].(time.Time)
	assert.Equal(t, int(end.Sub(start).Seconds()), 1)
}

func TestBaseService_CallAfter(t *testing.T) {
	TestBaseService_CallBefore(t)
}

func TestBaseService_Execute(t *testing.T) {
	assert.Nil(t, (&BaseService{}).Execute(context.TODO()))
}

func TestBaseService_Validate(t *testing.T) {
	assert.Nil(t, (&BaseService{}).Validate())
}
