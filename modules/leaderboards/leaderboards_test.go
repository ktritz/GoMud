package leaderboards

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/plugins"
)

func TestLoadLBsRestoresPersistedEntries(t *testing.T) {
	plug := plugins.New("leaderboards_load_test", "1.0")
	if plug == nil {
		t.Fatal("plugins.New() returned nil")
	}

	saved := LeaderboardModule{
		LB_Gold: leaderboardData{
			Top: []leaderboardEntry{
				{UserId: 42, CharacterName: "Saver", ScoreValue: 999},
			},
		},
		LB_Experience: leaderboardData{
			Top: []leaderboardEntry{
				{UserId: 77, CharacterName: "Wizard", ScoreValue: 1234},
			},
		},
	}

	if err := plug.WriteStruct(`latest-leaderboards`, saved); err != nil {
		t.Fatalf("WriteStruct() error = %v", err)
	}

	mod := LeaderboardModule{plug: plug}
	mod.loadLBs()

	if len(mod.LB_Gold.Top) != 1 || mod.LB_Gold.Top[0].UserId != 42 {
		t.Fatalf("LB_Gold.Top = %+v, want restored entry", mod.LB_Gold.Top)
	}

	if len(mod.LB_Experience.Top) != 1 || mod.LB_Experience.Top[0].UserId != 77 {
		t.Fatalf("LB_Experience.Top = %+v, want restored entry", mod.LB_Experience.Top)
	}

	if mod.LB_Gold.Name != "Gold" || mod.LB_Gold.ValueColor != "experience" {
		t.Fatalf("LB_Gold metadata = %+v, want restored metadata defaults", mod.LB_Gold)
	}

	if mod.LB_Experience.Name != "Experience" || mod.LB_Experience.ValueColor != "gold" {
		t.Fatalf("LB_Experience metadata = %+v, want restored metadata defaults", mod.LB_Experience)
	}

	if mod.LB_Kills.Name != "Kills" || mod.LB_Kills.ValueColor != "red-bold" {
		t.Fatalf("LB_Kills metadata = %+v, want metadata defaults", mod.LB_Kills)
	}

	if !mod.GoldEnabled || !mod.ExperienceEnabled || !mod.KillsEnabled {
		t.Fatalf("enabled flags = gold:%v xp:%v kills:%v, want all true", mod.GoldEnabled, mod.ExperienceEnabled, mod.KillsEnabled)
	}
}
