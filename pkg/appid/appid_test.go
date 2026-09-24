// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package appid

import "testing"

func TestRegistry(t *testing.T) {
	registered := []AppID{
		CounterStrike, TeamFortressClassic, DayOfDefeat, DeathmatchClassic,
		HalfLifeOpposingForce, Ricochet, HalfLife, CounterStrikeConditionZero,
		HalfLife2, CounterStrikeSource, DayOfDefeatSource, HalfLife2Deathmatch,
		HalfLifeDeathmatchSource, TeamFortress2, Left4Dead, Left4Dead2, Dota2,
		AlienSwarm, CounterStrike2, TheShip, NaturalSelection2, GarrysMod,
		ZombiePanicSource, Synergy, PiratesVikingsKnights2, Dystopia, NuclearDawn,
		ShatteredHorizon, MondayNightCombat, DinoDDay, Insurgency, NoMoreRoomInHell,
		SvenCoop, Contagion, FistfulOfFrags, DoubleActionBoogaloo,
		PrimalCarnageExtinction, BrainBread2, CodenameCure, BlackMesa, DayOfInfamy,
		Arma2, Arma2OperationArrowhead, Arma3, DayZ, DayZExperimental, ArmaReforger,
		ProjectZomboid, Starbound, TheForest, SpaceEngineers, SevenDaysToDie, Rust,
		Creativerse, Unturned, DontStarveTogether, RisingWorld, MedievalEngineers,
		ArkSurvivalEvolved, ColonySurvival, WurmUnlimited, TheIsle,
		EmpyrionGalacticSurvival, Hurtworld, AbioticFactor, ConanExiles, Avorion,
		DarkAndLight, SurviveTheNights, PixARK, Barotrauma, RiskOfRain2, ATLAS,
		Valheim, Foundry, DayOfDragons, Icarus, Enshrouded, SonsOfTheForest,
		MythOfEmpires, VRising, CoreKeeper, VEIN, TheFront, Soulmask,
		AliensVsPredator2010, Brink, RedOrchestra2, CallOfDutyMW3Multiplayer,
		Homefront, AmericasArmyProvingGrounds, ChivalryMedievalWarfare,
		KillingFloor2, QuakeLive, CallOfDutyBlackOps3, RisingStorm2Vietnam,
		InsurgencySandstorm, Mordhau, HellLetLoose, Squad44, OperationHarshDoorstop,
		EuroTruckSimulator2, ProjectCars, AmericanTruckSimulator, RFactor2,
		ProjectCars2, TowerUnite,
	}

	if len(names) != len(registered) {
		t.Fatalf("registry has %d names for %d IDs", len(names), len(registered))
	}

	seen := make(map[AppID]struct{}, len(registered))
	for _, id := range registered {
		if _, duplicate := seen[id]; duplicate {
			t.Fatalf("duplicate AppID %d", id)
		}
		seen[id] = struct{}{}

		name, ok := id.Name()
		if !ok || name == "" {
			t.Fatalf("AppID %d has no non-empty name", id)
		}
		if got := id.String(); got != name {
			t.Fatalf("String() = %q, want %q", got, name)
		}
	}
}

func TestUnknown(t *testing.T) {
	const id AppID = 4294967295

	if name, ok := id.Name(); ok || name != "" {
		t.Fatalf("Name() = %q, %t for unknown ID", name, ok)
	}
	if got := id.String(); got != "4294967295" {
		t.Fatalf("String() = %q, want numeric ID", got)
	}
}
