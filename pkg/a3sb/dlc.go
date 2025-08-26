package a3sb

import (
	"bytes"
	"fmt"

	"github.com/woozymasta/a2s/internal/bread"
	"github.com/woozymasta/steam/utils/appid"
)

// DLC 3rd and 4th bytes of the server browser protocol store the DLC bitmask flags
type DLC uint16

// DLCInfo store information about DLC
type DLCInfo struct {
	Name string `json:"name,omitempty"` // DLC name from predefined maps
	ID   uint64 `json:"id,omitempty"`   // DCL Steam AppID
	Hash uint32 `json:"hash,omitempty"` // DLC short hash
}

// DayZ DLC Map for DLC byte blocks
//   - https://steamdb.info/app/221100/dlc/
var dayzDLC = map[DLC]DLCInfo{
	0x1: {ID: 1151700, Name: "Livonia"},
	0x2: {ID: 2968040, Name: "Frost Line"},
	0x4: {ID: 3816030, Name: "Badlands"},
	0x8: {ID: 830660, Name: "Survivor GameZ"},
}

// Arma 3 DLC Map for DLC byte blocks
//   - https://community.bistudio.com/wiki/Category:Arma_3:_DLCs_%26_Expansions
//   - https://steamdb.info/app/107410/dlc/
var arma3DLC = map[DLC]DLCInfo{
	0x1:    {ID: 288520, Name: "Karts"},
	0x2:    {ID: 332350, Name: "Marksmen"},
	0x4:    {ID: 304380, Name: "Helicopters"},
	0x8:    {ID: 275700, Name: "Zeus"},
	0x10:   {ID: 395180, Name: "Apex"},
	0x20:   {ID: 601670, Name: "Jets"},
	0x40:   {ID: 571710, Name: "Laws of War"},
	0x80:   {ID: 639600, Name: "Malden"},
	0x100:  {ID: 744950, Name: "Tac-Ops Mission Pack"},
	0x200:  {ID: 798390, Name: "Tanks"},
	0x400:  {ID: 1021790, Name: "Enoch"},
	0x800:  {ID: 1021790, Name: "Contact (Platform)"},
	0x1000: {ID: 1325500, Name: "Art of War"},
}

// Read DLC from Arma 3 server browser proto
func (r *Rules) readDLC(buf *bytes.Buffer, dlcMask uint16) error {
	var dlcHashes []uint32

	switch r.id {
	case appid.Arma3.Uint64():
		r.DLC = parseDLC(dlcMask, arma3DLC)
	case appid.DayZ.Uint64(), appid.DayZExp.Uint64():
		r.DLC = parseDLC(dlcMask, dayzDLC)
	default:
		r.DLC = parseDLC(dlcMask, map[DLC]DLCInfo{})
	}

	// Read DLC 4-byte hashes and store them
	for i := 0; i < len(r.DLC); i++ {
		hash, err := bread.Uint32(buf)
		if err != nil {
			return err
		}
		dlcHashes = append(dlcHashes, hash)
	}

	// Assign hashes to corresponding DLCInfo structs
	for i := 0; i < len(r.DLC) && i < len(dlcHashes); i++ {
		r.DLC[i].Hash = dlcHashes[i]
	}

	return nil
}

// Parser for DLC mask
func parseDLC(mask uint16, dlcs map[DLC]DLCInfo) []DLCInfo {
	dlc := DLC(mask)
	var result []DLCInfo

	// Processing of known DLCs
	for bit, info := range dlcs {
		if dlc&bit != 0 {
			result = append(result, info)
			dlc &^= bit // Removing match DLC from the mask
		}
	}

	// Checking the remaining bits for unknown DLCs
	bit := DLC(1) // Start with the least significant bit
	for dlc != 0 {
		if dlc&bit != 0 {
			result = append(result, DLCInfo{
				ID:   0,
				Name: fmt.Sprintf("Unknown DLC %d", bit),
			})
			dlc &^= bit // Remove the processed bit from the mask
		}
		bit <<= 1 // Move on to the next bit
	}

	return result
}
