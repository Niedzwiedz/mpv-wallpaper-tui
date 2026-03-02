package ui

// previewReady is delivered when a static wallpaper preview has finished rendering.
type previewReady struct {
	path    string
	content string
}

// framesReady is delivered when all animation frames for a wallpaper are ready.
type framesReady struct {
	path   string
	frames []string
}

// animTick advances the animation by one frame.
// gen must match model.tickGen or the tick is discarded (prevents duplicate chains).
type animTick struct{ gen int }
