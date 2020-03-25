package config

// Config represents the window configuration for the application
type Config struct {
	ScreenWidth     int32
	ScreenHeight    int32
	BottomBarHeight int32
}

// New is an optional constructor for Config, mainly for a friendlier API.
func New(screenWidth, screenHeight, bottomBarHeight int32) *Config {
	return &Config{
		ScreenWidth:     screenWidth,
		ScreenHeight:    screenHeight,
		BottomBarHeight: bottomBarHeight,
	}
}
