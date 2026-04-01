package server

import "github.com/stockyard-dev/stockyard-chalkboard/internal/license"

type Limits struct {
	MaxPages       int
	VersionHistory bool
}

var freeLimits = Limits{MaxPages: 20, VersionHistory: false}
var proLimits = Limits{MaxPages: 0, VersionHistory: true}

func LimitsFor(info *license.Info) Limits {
	if info != nil && info.IsPro() { return proLimits }
	return freeLimits
}

func LimitReached(limit, current int) bool {
	if limit == 0 { return false }
	return current >= limit
}
