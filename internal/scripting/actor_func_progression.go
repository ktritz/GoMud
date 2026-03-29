package scripting

import (
	"strings"

	"github.com/GoMudEngine/GoMud/internal/combat"
	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/races"
	"github.com/GoMudEngine/GoMud/internal/skills"
	"github.com/GoMudEngine/GoMud/internal/templates"
)

func (a ScriptActor) GetLevel() int {
	return a.characterRecord.Level
}

// TODO: Need to allow for this stat list to be dynamic
func (a ScriptActor) GetStat(statName string) int {

	statName = strings.ToLower(statName)

	if strings.HasPrefix(statName, "st") {
		return a.characterRecord.Stats.GetValueAdj("Strength")
	}

	if strings.HasPrefix(statName, "sp") {
		return a.characterRecord.Stats.GetValueAdj("Speed")
	}

	if strings.HasPrefix(statName, "sm") {
		return a.characterRecord.Stats.GetValueAdj("Smarts")
	}

	if strings.HasPrefix(statName, "vi") {
		return a.characterRecord.Stats.GetValueAdj("Vitality")
	}

	if strings.HasPrefix(statName, "my") {
		return a.characterRecord.Stats.GetValueAdj("Mysticism")
	}

	if strings.HasPrefix(statName, "pe") {
		return a.characterRecord.Stats.GetValueAdj("Perception")
	}

	return 0
}

func (a ScriptActor) GetTameMastery() map[int]int {
	return a.characterRecord.MobMastery.GetAllTame()
}

func (a ScriptActor) SetTameMastery(mobId int, newSkillLevel int) {
	a.characterRecord.MobMastery.SetTame(mobId, newSkillLevel)
}

func (a ScriptActor) GetChanceToTame(target ScriptActor) int {
	return combat.ChanceToTame(a.userRecord, target.mobRecord)
}

func (a ScriptActor) TrainSkill(skillName string, skillLevel int) bool {

	if a.userRecord == nil {
		return false
	}

	skillName = strings.ToLower(skillName)
	currentLevel := a.characterRecord.GetSkillLevel(skills.SkillTag(skillName))

	if currentLevel < skillLevel {
		newLevel := a.characterRecord.TrainSkill(skillName, skillLevel)

		skillData := struct {
			SkillName  string
			SkillLevel int
		}{
			SkillName:  skillName,
			SkillLevel: newLevel,
		}
		skillUpTxt, _ := templates.Process("character/skillup", skillData, a.userRecord.UserId)
		a.SendText(skillUpTxt)

		return true

	}
	return false
}

func (a ScriptActor) GetSkillLevel(skillName string) int {
	return a.characterRecord.GetSkillLevel(skills.SkillTag(skillName))
}

func (a ScriptActor) IsTameable() bool {
	if a.mobRecord == nil {
		return false
	}
	return a.mobRecord.IsTameable()
}

func (a ScriptActor) HasSpell(spellId string) bool {
	return a.characterRecord.HasSpell(spellId)
}

func (a ScriptActor) LearnSpell(spellId string) bool {
	return a.characterRecord.LearnSpell(spellId)
}

func (a ScriptActor) GetMobKills(mobId int) int {
	return a.characterRecord.KD.GetMobKills(mobId)
}

func (a ScriptActor) GetRaceKills(race string) int {

	raceKills := map[string]int{}

	for mid, kCt := range a.characterRecord.KD.Kills {
		if mobSpec := mobs.GetMobSpec(mobs.MobId(mid)); mobSpec != nil {
			if raceInfo := races.GetRace(mobSpec.Character.RaceId); raceInfo != nil {
				raceKills[raceInfo.Name] = raceKills[raceInfo.Name] + kCt
			}
		}
	}

	return raceKills[race]
}

func (a ScriptActor) GetTrainingPoints() int {
	return a.characterRecord.TrainingPoints
}

func (a ScriptActor) GiveTrainingPoints(ct int) {
	if ct < 1 {
		return
	}
	a.characterRecord.TrainingPoints += ct
}

func (a ScriptActor) GetStatPoints() int {
	return a.characterRecord.StatPoints
}

func (a ScriptActor) GiveStatPoints(ct int) {
	if ct < 1 {
		return
	}
	a.characterRecord.StatPoints += ct
}

func (a ScriptActor) GiveExtraLife() {
	c := configs.GetGamePlayConfig()
	a.characterRecord.ExtraLives += 1
	if a.characterRecord.ExtraLives > int(c.LivesMax) {
		a.characterRecord.ExtraLives = int(c.LivesMax)
	}
}

func (a ScriptActor) GrantXP(xpAmt int, reason string) {
	if a.mobInstanceId > 0 {
		return
	}
	a.userRecord.GrantXP(xpAmt, reason)
}

