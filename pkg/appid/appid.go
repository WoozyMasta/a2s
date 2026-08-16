// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

// Package appid contains Steam application IDs
// for games known to use the A2S/Valve server query protocol.
//
// It is intentionally not a general Steam application registry.
// Entries are added only when they are relevant to A2S querying,
// protocol-specific handling, or user-facing identification.
package appid

import "strconv"

// AppID is a Steam application ID represented in the same width as the
// effective game identifiers used by the A2S API.
type AppID uint64

const (
	// Unknown represents an unknown application ID.
	Unknown = 0

	// GoldSource and Source games.
	CounterStrike              = 10
	TeamFortressClassic        = 20
	DayOfDefeat                = 30
	DeathmatchClassic          = 40
	HalfLifeOpposingForce      = 50
	Ricochet                   = 60
	HalfLife                   = 70
	CounterStrikeConditionZero = 80
	HalfLife2                  = 220
	CounterStrikeSource        = 240
	DayOfDefeatSource          = 300
	HalfLife2Deathmatch        = 320
	HalfLifeDeathmatchSource   = 360
	TeamFortress2              = 440
	Left4Dead                  = 500
	Left4Dead2                 = 550
	Dota2                      = 570
	AlienSwarm                 = 630
	CounterStrike2             = 730 // Current title for the former CS:GO AppID.
	TheShip                    = 2400
	NaturalSelection2          = 4920
	GarrysMod                  = 4000
	ZombiePanicSource          = 17500
	Synergy                    = 17520
	PiratesVikingsKnights2     = 17570
	Dystopia                   = 17580
	NuclearDawn                = 17710
	ShatteredHorizon           = 18110
	MondayNightCombat          = 63200
	DinoDDay                   = 70000
	Insurgency                 = 222880
	NoMoreRoomInHell           = 224260
	SvenCoop                   = 225840
	Contagion                  = 238430
	FistfulOfFrags             = 265630
	DoubleActionBoogaloo       = 317360
	PrimalCarnageExtinction    = 321360
	BrainBread2                = 346330
	CodenameCure               = 355180
	BlackMesa                  = 362890
	DayOfInfamy                = 447820

	// Bohemia Interactive games.
	Arma2                   = 33900
	Arma2OperationArrowhead = 33930
	Arma3                   = 107410
	DayZ                    = 221100
	DayZExperimental        = 1024020
	ArmaReforger            = 1874880 // Valve Query must be enabled server-side.

	// Survival and sandbox games.
	ProjectZomboid           = 108600
	Starbound                = 211820
	TheForest                = 242760
	SpaceEngineers           = 244850
	SevenDaysToDie           = 251570
	Rust                     = 252490
	Creativerse              = 280790
	Unturned                 = 304930
	DontStarveTogether       = 322330
	RisingWorld              = 324080
	MedievalEngineers        = 333950
	ArkSurvivalEvolved       = 346110
	ColonySurvival           = 366090
	WurmUnlimited            = 366220
	TheIsle                  = 376210 // Legacy Valve Query; Evrima uses EOS.
	EmpyrionGalacticSurvival = 383120
	Hurtworld                = 393420
	AbioticFactor            = 427410
	ConanExiles              = 440900
	Avorion                  = 445220
	DarkAndLight             = 529180
	SurviveTheNights         = 541300
	PixARK                   = 593600
	Barotrauma               = 602960
	RiskOfRain2              = 632360
	ATLAS                    = 834910
	Valheim                  = 892970
	Foundry                  = 983870
	DayOfDragons             = 1088090
	Icarus                   = 1149460
	Enshrouded               = 1203620
	SonsOfTheForest          = 1326470
	MythOfEmpires            = 1371580
	VRising                  = 1604030
	CoreKeeper               = 1621690
	VEIN                     = 1857950
	TheFront                 = 2285150
	Soulmask                 = 2646460

	// Tactical and shooter games.
	AliensVsPredator2010       = 10680
	Brink                      = 22350
	RedOrchestra2              = 35450
	CallOfDutyMW3Multiplayer   = 42690
	Homefront                  = 55100
	AmericasArmyProvingGrounds = 203290
	ChivalryMedievalWarfare    = 219640
	KillingFloor2              = 232090
	QuakeLive                  = 282440
	CallOfDutyBlackOps3        = 311210
	RisingStorm2Vietnam        = 418460
	InsurgencySandstorm        = 581320
	Mordhau                    = 629760
	HellLetLoose               = 686810
	Squad44                    = 736220 // Formerly known as Post Scriptum.
	OperationHarshDoorstop     = 736590

	// Racing and simulation games using Valve Query.
	EuroTruckSimulator2    = 227300
	ProjectCars            = 234630
	AmericanTruckSimulator = 270880
	RFactor2               = 365960
	ProjectCars2           = 378860

	// Other A2S-compatible games.
	TowerUnite = 394690
)

var names = map[AppID]string{
	CounterStrike:              "Counter-Strike",
	TeamFortressClassic:        "Team Fortress Classic",
	DayOfDefeat:                "Day of Defeat",
	DeathmatchClassic:          "Deathmatch Classic",
	HalfLifeOpposingForce:      "Half-Life: Opposing Force",
	Ricochet:                   "Ricochet",
	HalfLife:                   "Half-Life",
	CounterStrikeConditionZero: "Counter-Strike: Condition Zero",
	HalfLife2:                  "Half-Life 2",
	CounterStrikeSource:        "Counter-Strike: Source",
	DayOfDefeatSource:          "Day of Defeat: Source",
	HalfLife2Deathmatch:        "Half-Life 2: Deathmatch",
	HalfLifeDeathmatchSource:   "Half-Life Deathmatch: Source",
	TeamFortress2:              "Team Fortress 2",
	Left4Dead:                  "Left 4 Dead",
	Left4Dead2:                 "Left 4 Dead 2",
	Dota2:                      "Dota 2",
	AlienSwarm:                 "Alien Swarm",
	CounterStrike2:             "Counter-Strike 2",
	TheShip:                    "The Ship",
	NaturalSelection2:          "Natural Selection 2",
	GarrysMod:                  "Garry's Mod",
	ZombiePanicSource:          "Zombie Panic! Source",
	Synergy:                    "Synergy",
	PiratesVikingsKnights2:     "Pirates, Vikings, & Knights II",
	Dystopia:                   "Dystopia",
	NuclearDawn:                "Nuclear Dawn",
	ShatteredHorizon:           "Shattered Horizon",
	MondayNightCombat:          "Monday Night Combat",
	DinoDDay:                   "Dino D-Day",
	Insurgency:                 "Insurgency",
	NoMoreRoomInHell:           "No More Room in Hell",
	SvenCoop:                   "Sven Co-op",
	Contagion:                  "Contagion",
	FistfulOfFrags:             "Fistful of Frags",
	DoubleActionBoogaloo:       "Double Action: Boogaloo",
	PrimalCarnageExtinction:    "Primal Carnage: Extinction",
	BrainBread2:                "BrainBread 2",
	CodenameCure:               "Codename CURE",
	BlackMesa:                  "Black Mesa",
	DayOfInfamy:                "Day of Infamy",
	Arma2:                      "ARMA 2",
	Arma2OperationArrowhead:    "ARMA 2: Operation Arrowhead",
	Arma3:                      "Arma 3",
	DayZ:                       "DayZ",
	DayZExperimental:           "DayZ Experimental",
	ArmaReforger:               "Arma Reforger",
	ProjectZomboid:             "Project Zomboid",
	Starbound:                  "Starbound",
	TheForest:                  "The Forest",
	SpaceEngineers:             "Space Engineers",
	SevenDaysToDie:             "7 Days to Die",
	Rust:                       "Rust",
	Creativerse:                "Creativerse",
	Unturned:                   "Unturned",
	DontStarveTogether:         "Don't Starve Together",
	RisingWorld:                "Rising World",
	MedievalEngineers:          "Medieval Engineers",
	ArkSurvivalEvolved:         "ARK: Survival Evolved",
	ColonySurvival:             "Colony Survival",
	WurmUnlimited:              "Wurm Unlimited",
	TheIsle:                    "The Isle",
	EmpyrionGalacticSurvival:   "Empyrion - Galactic Survival",
	Hurtworld:                  "Hurtworld",
	AbioticFactor:              "Abiotic Factor",
	ConanExiles:                "Conan Exiles",
	Avorion:                    "Avorion",
	DarkAndLight:               "Dark and Light",
	SurviveTheNights:           "Survive the Nights",
	PixARK:                     "PixARK",
	Barotrauma:                 "Barotrauma",
	RiskOfRain2:                "Risk of Rain 2",
	ATLAS:                      "ATLAS",
	Valheim:                    "Valheim",
	Foundry:                    "Foundry",
	DayOfDragons:               "Day of Dragons",
	Icarus:                     "Icarus",
	Enshrouded:                 "Enshrouded",
	SonsOfTheForest:            "Sons Of The Forest",
	MythOfEmpires:              "Myth of Empires",
	VRising:                    "V Rising",
	CoreKeeper:                 "Core Keeper",
	VEIN:                       "VEIN",
	TheFront:                   "The Front",
	Soulmask:                   "Soulmask",
	AliensVsPredator2010:       "Aliens vs. Predator 2010",
	Brink:                      "BRINK",
	RedOrchestra2:              "Red Orchestra 2: Heroes of Stalingrad",
	CallOfDutyMW3Multiplayer:   "Call of Duty: Modern Warfare 3",
	Homefront:                  "Homefront",
	AmericasArmyProvingGrounds: "America's Army: Proving Grounds",
	ChivalryMedievalWarfare:    "Chivalry: Medieval Warfare",
	KillingFloor2:              "Killing Floor 2",
	QuakeLive:                  "Quake Live",
	CallOfDutyBlackOps3:        "Call of Duty: Black Ops III",
	RisingStorm2Vietnam:        "Rising Storm 2: Vietnam",
	InsurgencySandstorm:        "Insurgency: Sandstorm",
	Mordhau:                    "MORDHAU",
	HellLetLoose:               "Hell Let Loose",
	Squad44:                    "Squad 44",
	OperationHarshDoorstop:     "Operation: Harsh Doorstop",
	EuroTruckSimulator2:        "Euro Truck Simulator 2",
	ProjectCars:                "Project CARS",
	AmericanTruckSimulator:     "American Truck Simulator",
	RFactor2:                   "rFactor 2",
	ProjectCars2:               "Project CARS 2",
	TowerUnite:                 "Tower Unite",
}

// Name returns the curated name for an A2S-compatible application ID.
func (id AppID) Name() (string, bool) {
	name, ok := names[id]
	return name, ok
}

// String returns the curated name or the numeric ID when it is unknown.
func (id AppID) String() string {
	if name, ok := id.Name(); ok {
		return name
	}

	return strconv.FormatUint(uint64(id), 10)
}
