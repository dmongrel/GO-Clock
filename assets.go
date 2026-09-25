// SPDX-FileCopyrightText: 2026 Joel L. Caesar
// SPDX-License-Identifier: Apache-2.0

package main

import "embed"

// assetFS carries the icons and the bundled alarm sounds. It lives in its own
// untagged file because both entry points need it: the Fyne build turns the
// SVGs into fyne.Resource values, and the Wails build hands them to the
// frontend as markup.
//
//go:embed images/*.svg
//go:embed alarms/*.mp3
var assetFS embed.FS
