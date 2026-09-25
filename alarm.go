// SPDX-FileCopyrightText: 2026 Joel L. Caesar
// SPDX-License-Identifier: Apache-2.0

package main

import "time"

// alarmDue reports whether an alarm set for alarmTime ("15:04") should sound at
// now, given lastTriggered - the minute at which it last sounded, or -1 for
// none. It also returns the value lastTriggered should take next, which is how
// a single alarm is prevented from sounding repeatedly within its own minute.
//
// An unparseable alarmTime is not due and leaves lastTriggered alone.
//
// It is deliberately a pure function in its own untagged file: both entry
// points run an alarm checker, and this is the part worth testing.
func alarmDue(now time.Time, alarmTime string, lastTriggered int) (due bool, nextLastTriggered int) {
	t, err := time.Parse("15:04", alarmTime)
	if err != nil {
		return false, lastTriggered
	}
	if now.Hour() != t.Hour() || now.Minute() != t.Minute() {
		return false, -1
	}
	if lastTriggered == now.Minute() {
		return false, lastTriggered
	}
	return true, now.Minute()
}
