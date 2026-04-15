//go:build pebbledb

package config

import (
	"github.com/cockroachdb/pebble"
	dbm "github.com/cometbft/cometbft-db"
)

func newPebbleDBWithTuning(ctx *DBContext) (dbm.DB, error) {
	return dbm.NewPebbleDBWithOpts(ctx.ID, ctx.Config.DBDir(), buildPebbleOptions(defaultedDBTuning(ctx.Config).Pebble))
}

func buildPebbleOptions(cfg PebbleTuningConfig) *pebble.Options {
	opts := dbm.DefaultPebbleOptions()
	opts.BytesPerSync = cfg.BytesPerSync
	opts.WALBytesPerSync = cfg.WALBytesPerSync
	opts.L0CompactionThreshold = cfg.L0CompactionThreshold
	opts.L0CompactionFileThreshold = cfg.L0CompactionFileThreshold
	opts.L0StopWritesThreshold = cfg.L0StopWritesThreshold
	opts.LBaseMaxBytes = cfg.LBaseMaxBytes
	opts.MaxManifestFileSize = cfg.MaxManifestFileSize
	opts.MaxOpenFiles = cfg.MaxOpenFiles
	opts.MemTableSize = uint64(cfg.MemTableSize)
	opts.MemTableStopWritesThreshold = cfg.MemTableStopWritesThreshold
	opts.MaxConcurrentCompactions = func() int { return cfg.MaxConcurrentCompactions }
	opts.FlushSplitBytes = cfg.FlushSplitBytes
	opts.Levels = []pebble.LevelOptions{{TargetFileSize: cfg.Level0TargetFileSize}}
	opts.Experimental.L0CompactionConcurrency = cfg.ExperimentalL0CompactionConcurrency
	opts.Experimental.CompactionDebtConcurrency = uint64(cfg.ExperimentalCompactionDebtConcurrency)
	opts.Experimental.ReadCompactionRate = cfg.ExperimentalReadCompactionRate
	opts.Experimental.ReadSamplingMultiplier = cfg.ExperimentalReadSamplingMultiplier
	opts.EnsureDefaults()
	return opts
}
