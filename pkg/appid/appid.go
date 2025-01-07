package appid

// https://api.steampowered.com/ISteamApps/GetAppList/v2/
// https://api.steampowered.com/IStoreService/GetAppList/v1/
// https://steamdb.info/

const (
	Unknown uint64 = 0

	// GoldSource

	CounterStrike       uint64 = 10  // Counter Strike 1.6
	TeamFortressClassic uint64 = 20  // Team Fortress Classic
	DayOfDefeat         uint64 = 30  // Day Of Defeat
	DeathMatchClassic   uint64 = 40  // Half-Life: DeathMatch Classic
	OpposingForce       uint64 = 50  // Opposing Force
	Ricochet            uint64 = 60  // Ricochet
	HalfLife            uint64 = 70  // Half-Life
	CounterStrikeCZ     uint64 = 80  // Counter Strike: Condition Zero
	HalfLifeBlueShift   uint64 = 130 // Half-Life: BlueShift

	// Source 1/2

	HalfLife2           uint64 = 220  // Half-Life 2
	CounterStrikeSource uint64 = 240  // Counter Strike Source
	DayOfDefeatSource   uint64 = 300  // Day Of Defeat Source
	HalfLifeDMSource    uint64 = 360  // Half-Life: DeathMatch Source
	Portal              uint64 = 400  // Portal
	TeamFortress2       uint64 = 440  // Team Fortress 2
	Left4Dead           uint64 = 500  // Left 4 Dead
	Left4Dead2          uint64 = 550  // Left 4 Dead 2
	Dota2               uint64 = 570  // Dota 2
	Portal2             uint64 = 620  // Portal 2
	AlienSwarm          uint64 = 630  // Alien Swarm
	CounterStrike2      uint64 = 730  // Counter Strike 2
	KillingFloor        uint64 = 1250 // Killing Floor
	TheShip             uint64 = 2400 // The Ship

	// Other

	Arma3   uint64 = 107410  // Arma 3
	DayZ    uint64 = 221100  // DayZ
	DayZExp uint64 = 1024020 // DayZ Experimental
	Rust    uint64 = 252490  // Rust
	Valheim uint64 = 892970  // Valheim
)
