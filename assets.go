// SPDX-FileCopyrightText: 2026 Joel L. Caesar
// SPDX-License-Identifier: Apache-2.0

package main

import "embed"

// assetFS carries the icons and the bundled alarm sounds. The icons are handed
// to the frontend as markup, so CSS can recolour them; the sounds are decoded
// straight into the audio buffer.
//
//go:embed images/*.svg
//go:embed alarms/*.mp3
var assetFS embed.FS
