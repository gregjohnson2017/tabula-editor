package config

// Config represents the window configuration for the application
type Config struct {
	ScreenWidth     int32
	ScreenHeight    int32
	BottomBarHeight int32
	FramesPerSecond int
}

// New is an optional constructor for Config, mainly for a friendlier API.
func New(screenWidth, screenHeight, bottomBarHeight int32, fps int) *Config {
	return &Config{
		ScreenWidth:     screenWidth,
		ScreenHeight:    screenHeight,
		BottomBarHeight: bottomBarHeight,
		FramesPerSecond: fps,
	}
}
