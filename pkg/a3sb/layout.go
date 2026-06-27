package a3sb

// Layout identifies the game-specific binary layout used by A3SB.
type Layout uint8

const (
	// LayoutUnknown leaves game-specific fields unselected.
	LayoutUnknown Layout = iota

	// LayoutArma3 selects the Arma 3 layout, including its difficulty bytes.
	LayoutArma3

	// LayoutDayZ selects the DayZ layout, which omits the Arma 3 difficulty bytes.
	LayoutDayZ
)

// String returns the layout name.
func (l Layout) String() string {
	switch l {
	case LayoutArma3:
		return "Arma3"

	case LayoutDayZ:
		return "DayZ"

	default:
		return "Unknown"
	}
}
