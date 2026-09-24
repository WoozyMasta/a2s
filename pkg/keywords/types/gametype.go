// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package types

// GameType represents an Arma 3 game mode keyword.
type GameType string

const (
	GameTApex    GameType = "apex"    // Apex: Campaign - Apex Protocol
	GameTCoop    GameType = "coop"    // Coop: Cooperative Mission
	GameTCTF     GameType = "ctf"     // CTF: Capture The Flag
	GameTCTI     GameType = "cti"     // CTI: Capture The Island
	GameTDM      GameType = "dm"      // DM: Deathmatch
	GameTEndgame GameType = "endgame" // EndGame: End Game
	GameTEscape  GameType = "escape"  // Escape: Escape
	GameTKotH    GameType = "koth"    // KOTH: King Of The Hill
	GameTLastman GameType = "lastman" // LastMan: Last Man Standing
	GameTPatrol  GameType = "patrol"  // Patrol: Combat Patrol
	GameTRPG     GameType = "rpg"     // RPG: Role-Playing Game
	GameTSandbox GameType = "sandbox" // Sandbox: Sandbox
	GameTSC      GameType = "sc"      // SC: Sector Control
	GameTSupport GameType = "support" // Support: Support
	GameTSurvive GameType = "survive" // Survive: Survival
	GameTTDM     GameType = "tdm"     // TDM: Team Deathmatch
	GameTUnknown GameType = "unknown" // Unknown: Undefined Game Mode
	GameTVanguar GameType = "vanguar" // Vanguard: Vanguard
	GameTWarlord GameType = "warlord" // Warlords: Warlords
	GameTZeus    GameType = "zeus"    // Zeus: Zeus
)

// String returns the human-readable name for the GameType keyword.
func (gt GameType) String() string {
	switch gt {
	case GameTApex:
		return "Campaign - Apex Protocol"
	case GameTCoop:
		return "Cooperative Mission"
	case GameTCTF:
		return "Capture The Flag"
	case GameTCTI:
		return "Capture The Island"
	case GameTDM:
		return "Deathmatch"
	case GameTEndgame:
		return "End Game"
	case GameTEscape:
		return "Escape"
	case GameTKotH:
		return "King Of The Hill"
	case GameTLastman:
		return "Last Man Standing"
	case GameTPatrol:
		return "Combat Patrol"
	case GameTRPG:
		return "Role-Playing Game"
	case GameTSandbox:
		return "Sandbox"
	case GameTSC:
		return "Sector Control"
	case GameTSupport:
		return "Support"
	case GameTSurvive:
		return "Survival"
	case GameTTDM:
		return "Team Deathmatch"
	case GameTUnknown:
		return "Undefined Game Mode"
	case GameTVanguar:
		return "Vanguard"
	case GameTWarlord:
		return "Warlords"
	case GameTZeus:
		return "Zeus"
	default:
		return string(gt)
	}
}
