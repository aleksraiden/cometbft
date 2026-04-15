package config

import (
	"context"

	dbm "github.com/cometbft/cometbft-db"
	"github.com/syndtr/goleveldb/leveldb/opt"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/libs/service"
)

// ServiceProvider takes a config and a logger and returns a ready to go Node.
type ServiceProvider func(context.Context, *Config, log.Logger) (service.Service, error)

// DBContext specifies config information for loading a new DB.
type DBContext struct {
	ID     string
	Config *Config
}

// DBProvider takes a DBContext and returns an instantiated DB.
type DBProvider func(*DBContext) (dbm.DB, error)

// DefaultDBProvider returns a database using the DBBackend and DBDir
// specified in the Config.
func DefaultDBProvider(ctx *DBContext) (dbm.DB, error) {
	dbType := dbm.BackendType(ctx.Config.DBBackend)
	dbTuning := defaultedDBTuning(ctx.Config)

	switch dbType {
	case dbm.GoLevelDBBackend:
		return dbm.NewGoLevelDBWithOpts(ctx.ID, ctx.Config.DBDir(), buildGoLevelDBOptions(dbTuning.GoLevelDB))
	case dbm.PebbleDBBackend:
		return newPebbleDBWithTuning(ctx)
	}

	return dbm.NewDB(ctx.ID, dbType, ctx.Config.DBDir())
}

func defaultedDBTuning(cfg *Config) *DBTuningConfig {
	if cfg.DBTuning == nil {
		return DefaultDBTuningConfig()
	}
	return cfg.DBTuning
}

func buildGoLevelDBOptions(cfg GoLevelDBTuningConfig) *opt.Options {
	opts := dbm.DefaultGoLevelDBOptions()
	opts.BlockCacheCapacity = cfg.BlockCacheCapacity
	opts.OpenFilesCacheCapacity = cfg.OpenFilesCacheCapacity
	opts.BlockSize = cfg.BlockSize
	opts.CompactionL0Trigger = cfg.CompactionL0Trigger
	opts.CompactionTableSize = cfg.CompactionTableSize
	opts.CompactionTotalSize = cfg.CompactionTotalSize
	opts.WriteBuffer = cfg.WriteBuffer
	opts.WriteL0SlowdownTrigger = cfg.WriteL0SlowdownTrigger
	opts.WriteL0PauseTrigger = cfg.WriteL0PauseTrigger
	opts.IteratorSamplingRate = cfg.IteratorSamplingRate
	opts.NoSync = cfg.NoSync
	opts.CompactionTableSizeMultiplier = cfg.CompactionTableSizeMultiplier
	opts.CompactionTotalSizeMultiplier = cfg.CompactionTotalSizeMultiplier

	switch cfg.Compression {
	case "none":
		opts.Compression = opt.NoCompression
	default:
		opts.Compression = opt.SnappyCompression
	}
	return opts
}
