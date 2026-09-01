package hub

import (
	"time"

	"github.com/robfig/cron/v3"
)

// nextCronRun 判断 cron 表达式是否在当前时刻到期。
// lastFire 为该 Schedule 上次触发时间（零值表示首次），用于避免重复触发与错失窗口。
func nextCronRun(expr string, loc *time.Location, lastFire, createdAt time.Time, now time.Time) (time.Time, bool, error) {
	spec, err := cron.ParseStandard(expr)
	if err != nil {
		return time.Time{}, false, err
	}
	ref := lastFire
	if ref.IsZero() {
		ref = createdAt
		if ref.IsZero() {
			ref = now.Add(-time.Second)
		}
	}
	next := spec.Next(ref)
	if !next.After(now) {
		// 已到/越过触发点
		return next, true, nil
	}
	return next, false, nil
}
