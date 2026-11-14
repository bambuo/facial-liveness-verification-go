// Package clock 提供系统时钟适配器。
package clock

import "time"

// System 返回当前 UTC 时间。
type System struct{}

// Now 实现 port.Clock。
func (System) Now() time.Time { return time.Now().UTC() }
