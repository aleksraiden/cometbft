//go:build !pebbledb

package config

import dbm "github.com/cometbft/cometbft-db"

func newPebbleDBWithTuning(ctx *DBContext) (dbm.DB, error) {
	return dbm.NewDB(ctx.ID, dbm.PebbleDBBackend, ctx.Config.DBDir())
}
