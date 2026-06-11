package a3sb

import (
	"testing"

	"github.com/woozymasta/a2s/internal/a2srules"
	"github.com/woozymasta/a2s/pkg/a2s"
	"github.com/woozymasta/a2s/pkg/appid"
)

func FuzzReadA3SB(f *testing.F) {
	f.Add([]byte{3, 0, 0, 0, 0, 0, 0})
	f.Add([]byte{2, 0, 0, 0, 0, 0})
	f.Fuzz(func(t *testing.T, data []byte) {
		for _, game := range []uint64{appid.Arma3, appid.DayZ, appid.DayZExperimental, 0} {
			rules := &Rules{id: game}
			_ = rules.readA3SB(data)
		}
	})
}

func FuzzBuildPageEnvelope(f *testing.F) {
	f.Add([]byte{1, 0, 1, 1, 0, 3, 0, 0, 0, 0, 0, 0, 0})
	f.Fuzz(func(t *testing.T, data []byte) {
		result, err := a2srules.Parse(data)
		if err != nil {
			return
		}

		_, _ = buildPageEnvelope(result.Entries, result.Remaining, false)
	})
}

func FuzzParseAutomaticRules(f *testing.F) {
	f.Add([]byte{1, 0, 'h', 'o', 's', 't', 0, 'e', 'x', 'a', 'm', 'p', 'l', 'e', 0})
	f.Add([]byte{1, 0, 1, 1, 0, 3, 0, 0, 0, 0, 0, 0, 0})
	f.Fuzz(func(t *testing.T, data []byte) {
		result, err := a2srules.Parse(data)
		if err != nil {
			return
		}

		_, _ = parseAutomatic(result)
	})
}

func FuzzParseRulesDayZ(f *testing.F) {
	f.Add([]byte("allowedBuild"), []byte("123"))
	f.Fuzz(func(t *testing.T, key, value []byte) {
		rules := &Rules{id: appid.DayZ}
		_ = rules.parseRulesDayZ(a2s.Rules{{Name: string(key), Value: string(value)}})
	})
}
