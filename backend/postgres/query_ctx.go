// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.
// Created at 2020/02/04 by Vincent Landgraf

package postgres

import (
	"context"

	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
)

type pgPoolAdapter struct {
	db *pg.DB
}

func (a *pgPoolAdapter) Exec(ctx context.Context, query interface{}, params ...interface{}) (res orm.Result, err error) {
	db := a.db.WithContext(ctx)
	return db.Exec(query, params...)
}
