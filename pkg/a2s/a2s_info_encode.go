// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package a2s

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/woozymasta/a2s/pkg/appid"
)

// knownInfoEDF contains the EDF bits supported by the Info model encoder.
var knownInfoEDF = edfPort | edfSteamID | edfSourceTV | edfKeywords | edfGameID

// AppendInfo appends one complete logical A2S_INFO response to dst.
//
// The response format is selected by info.Format.
// Known EDF bits are derived from populated optional fields;
// Info.EDF is retained as decoded metadata
// and is not used as the encoder's source of truth.
func AppendInfo(dst []byte, info Info) ([]byte, error) {
	edf, err := infoEncodeEDF(info)
	if err != nil {
		return dst, err
	}
	if err := validateInfoForEncoding(info); err != nil {
		return dst, err
	}

	switch ResponseType(info.Format) {
	case ResponseInfo:
		dst = appendPacketHeader(dst, ResponseInfo)
		return appendSourceInfo(dst, info, edf)

	case ResponseInfoGoldSource:
		dst = appendPacketHeader(dst, ResponseInfoGoldSource)
		return appendGoldSourceInfo(dst, info)

	default:
		return dst, fmt.Errorf("%w: unsupported format 0x%X", ErrInfoEncode, info.Format)
	}
}

// infoEncodeEDF derives the supported EDF mask from populated Info fields.
func infoEncodeEDF(info Info) (EDF, error) {
	if unknown := info.EDF &^ knownInfoEDF; unknown != 0 {
		return 0, fmt.Errorf("%w: unsupported EDF bits 0x%X", ErrInfoEncode, unknown)
	}

	var edf EDF
	if info.Port != 0 {
		edf |= edfPort
	}
	if info.SteamID != 0 {
		edf |= edfSteamID
	}
	if info.SourceTVName != "" || info.SourceTVPort != 0 {
		edf |= edfSourceTV
	}
	if info.Keywords != nil {
		edf |= edfKeywords
	}
	if info.GameID != nil {
		edf |= edfGameID
	}

	return edf, nil
}

// validateInfoForEncoding rejects fields that cannot be represented by the selected format.
func validateInfoForEncoding(info Info) error {
	switch ResponseType(info.Format) {
	case ResponseInfo:
		if info.Address != "" || info.Mod != nil {
			return fmt.Errorf("%w: GoldSource fields set for Source response", ErrInfoEncode)
		}
		if info.AppID == uint16(appid.TheShip) && info.TheShip == nil {
			return fmt.Errorf("%w: The Ship data is missing", ErrInfoEncode)
		}
		if info.AppID != uint16(appid.TheShip) && info.TheShip != nil {
			return fmt.Errorf("%w: The Ship data set for another AppID", ErrInfoEncode)
		}

		for name, value := range map[string]string{
			"name":           info.Name,
			"map":            info.Map,
			"folder":         info.Folder,
			"game":           info.Game,
			"version":        info.Version,
			"source_tv_name": info.SourceTVName,
		} {
			if err := validateInfoString(name, value); err != nil {
				return err
			}
		}

		return validateInfoKeywords(info.Keywords)

	case ResponseInfoGoldSource:
		if info.AppID != 0 || info.GameID != nil || info.SteamID != 0 ||
			info.Port != 0 || info.SourceTVName != "" || info.SourceTVPort != 0 ||
			info.Keywords != nil || info.EDF != 0 || info.TheShip != nil {
			return fmt.Errorf("%w: Source fields set for GoldSource response", ErrInfoEncode)
		}

		for name, value := range map[string]string{
			"address": info.Address,
			"name":    info.Name,
			"map":     info.Map,
			"folder":  info.Folder,
			"game":    info.Game,
		} {
			if err := validateInfoString(name, value); err != nil {
				return err
			}
		}
		if info.Mod != nil {
			if err := validateInfoString("mod.link", info.Mod.Link); err != nil {
				return err
			}
			if err := validateInfoString("mod.download_link", info.Mod.DownloadLink); err != nil {
				return err
			}
		}

		return nil

	default:
		return fmt.Errorf("%w: unsupported format 0x%X", ErrInfoEncode, info.Format)
	}
}

// validateInfoString rejects NUL bytes that would terminate a protocol string early.
func validateInfoString(name, value string) error {
	if strings.IndexByte(value, 0) >= 0 {
		return fmt.Errorf("%w: %s contains NUL", ErrInfoEncode, name)
	}

	return nil
}

// validateInfoKeywords validates the comma-separated keyword representation.
func validateInfoKeywords(keywords []string) error {
	for _, keyword := range keywords {
		if strings.ContainsRune(keyword, ',') {
			return fmt.Errorf("%w: keyword contains comma", ErrInfoEncode)
		}
		if err := validateInfoString("keyword", keyword); err != nil {
			return err
		}
	}

	return nil
}

// appendSourceInfo appends the Source A2S_INFO payload after the packet header.
func appendSourceInfo(dst []byte, info Info, edf EDF) ([]byte, error) {
	dst = append(dst, info.Protocol)

	var err error
	for _, field := range []struct {
		name  string
		value string
	}{
		{name: "name", value: info.Name},
		{name: "map", value: info.Map},
		{name: "folder", value: info.Folder},
		{name: "game", value: info.Game},
	} {
		dst, err = appendInfoCString(dst, field.name, field.value)
		if err != nil {
			return dst, err
		}
	}

	dst = binary.LittleEndian.AppendUint16(dst, info.AppID)
	dst = append(dst, info.Players, info.MaxPlayers, info.Bots, byte(info.ServerType), byte(info.Environment))
	dst = appendInfoBool(dst, info.Visibility)
	dst = appendInfoBool(dst, info.VAC)

	if info.AppID == uint16(appid.TheShip) {
		dst = append(dst, byte(info.TheShip.Mode), info.TheShip.Witnesses, info.TheShip.Duration)
	}

	dst, err = appendInfoCString(dst, "version", info.Version)
	if err != nil {
		return dst, err
	}
	dst = append(dst, byte(edf))

	if edf&edfPort != 0 {
		dst = binary.LittleEndian.AppendUint16(dst, info.Port)
	}
	if edf&edfSteamID != 0 {
		dst = binary.LittleEndian.AppendUint64(dst, info.SteamID)
	}
	if edf&edfSourceTV != 0 {
		dst = binary.LittleEndian.AppendUint16(dst, info.SourceTVPort)
		dst, err = appendInfoCString(dst, "source_tv_name", info.SourceTVName)
		if err != nil {
			return dst, err
		}
	}
	if edf&edfKeywords != 0 {
		dst, err = appendInfoCString(dst, "keywords", strings.Join(info.Keywords, ","))
		if err != nil {
			return dst, err
		}
	}
	if edf&edfGameID != 0 {
		dst = binary.LittleEndian.AppendUint64(dst, *info.GameID)
	}

	return dst, nil
}

// appendGoldSourceInfo appends the GoldSource A2S_INFO payload after the packet header.
func appendGoldSourceInfo(dst []byte, info Info) ([]byte, error) {
	var err error
	for _, field := range []struct {
		name  string
		value string
	}{
		{name: "address", value: info.Address},
		{name: "name", value: info.Name},
		{name: "map", value: info.Map},
		{name: "folder", value: info.Folder},
		{name: "game", value: info.Game},
	} {
		dst, err = appendInfoCString(dst, field.name, field.value)
		if err != nil {
			return dst, err
		}
	}

	dst = append(dst, info.Players, info.MaxPlayers, info.Protocol, byte(info.ServerType), byte(info.Environment))
	dst = appendInfoBool(dst, info.Visibility)
	dst = appendInfoBool(dst, info.Mod != nil)
	if info.Mod != nil {
		dst, err = appendInfoCString(dst, "mod.link", info.Mod.Link)
		if err != nil {
			return dst, err
		}
		dst, err = appendInfoCString(dst, "mod.download_link", info.Mod.DownloadLink)
		if err != nil {
			return dst, err
		}
		dst = binary.LittleEndian.AppendUint32(dst, info.Mod.Version)
		dst = binary.LittleEndian.AppendUint32(dst, info.Mod.Size)
		dst = appendInfoBool(dst, info.Mod.Type)
		dst = appendInfoBool(dst, info.Mod.DLL)
	}
	dst = appendInfoBool(dst, info.VAC)
	dst = append(dst, info.Bots)

	return dst, nil
}

// appendInfoCString appends one validated NUL-terminated A2S string.
func appendInfoCString(dst []byte, name, value string) ([]byte, error) {
	if err := validateInfoString(name, value); err != nil {
		return dst, err
	}

	dst = append(dst, value...)
	return append(dst, 0), nil
}

// appendInfoBool appends the one-byte A2S boolean representation.
func appendInfoBool(dst []byte, value bool) []byte {
	if value {
		return append(dst, 1)
	}

	return append(dst, 0)
}
