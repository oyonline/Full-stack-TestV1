// Package assets provides embedded static assets for the application.
package assets

import _ "embed"

//go:embed fonts/NotoSansSC-Regular.ttf
var NotoSansSC []byte

//go:embed fonts/Inter-SemiBold.ttf
var InterSemiBold []byte
