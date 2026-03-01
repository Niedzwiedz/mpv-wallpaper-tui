package ui

// previewReady is delivered when a wallpaper preview has finished rendering.
type previewReady struct {
	path    string
	content string
}
