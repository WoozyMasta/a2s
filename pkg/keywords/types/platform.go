package types

// Platform represent game-server platform (OS)
type Platform string

const (
	OSWLinux  Platform = "l" // Linux
	OSMac     Platform = "m" // MacOS
	OSOther   Platform = "o" // Other
	OSWindows Platform = "w" // Windows
)

// String return string represent of char
func (p Platform) String() string {
	switch p {
	case OSWLinux:
		return "Linux"
	case OSMac:
		return "MacOS"
	case OSOther:
		return "Other"
	case OSWindows:
		return "Windows"
	default:
		return "Undefined"
	}
}
