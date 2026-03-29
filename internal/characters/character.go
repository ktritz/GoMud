package characters

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/GoMudEngine/GoMud/internal/buffs"
	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/gametime"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/pets"
	"github.com/GoMudEngine/GoMud/internal/races"
	"github.com/GoMudEngine/GoMud/internal/statmods"
	"github.com/GoMudEngine/GoMud/internal/stats"
	"github.com/GoMudEngine/GoMud/internal/util"

	"maps"
	"slices"
)

var (
	startingRace     = 0
	startingHealth   = 10
	startingMana     = 10
	StartingRoomId   = -1
	startingZone     = `Nowhere`
	defaultName      = `nameless`
	descriptionCache = map[string]string{} // key is a hash, value is the description
)

type NameRenderFlag uint8

const (
	RenderHealth NameRenderFlag = iota
	RenderAggro
	RenderShortAdjectives
)

type Character struct {
	Name             string                         // The name of the character
	Description      string                         // A description of the character.
	Adjectives       []string                       `yaml:"adjectives,omitempty"` // Decorative text for the name of the character (e.g. "sleeping", "dead", "wounded")
	RoomId           int                            // The room id the character is in.
	Zone             string                         // The zone the character is in. The folder the room can be located in too.
	RaceId           int                            // Character race
	Stats            stats.Statistics               // Character stats
	Level            int                            // The level of the character
	Experience       int                            // The experience of the character
	TrainingPoints   int                            // The number of training points the character has
	StatPoints       int                            // The number of skill points the character has
	Health           int                            // The health of the character
	Mana             int                            // The mana of the character
	ActionPoints     int                            // The resevoir of action points the character has to spend on movement etc.
	Alignment        int8                           // The alignment of the character
	Gold             int                            // The gold the character is holding
	Bank             int                            // The gold the character has in the bank
	Shop             Shop                           `yaml:"shop,omitempty"`          // Definition of shop services/items this character stocks (or just has at the moment)
	SpellBook        map[string]int                 `yaml:"spellbook,omitempty"`     // The spells the character has learned
	Charmed          *CharmInfo                     `yaml:"-"`                       // If they are charmed, this is the info
	CharmedMobs      []int                          `yaml:"-"`                       // If they have charmed anyone, this is the list of mob instance ids
	Items            []items.Item                   `yaml:"items,omitempty"`         // The items the character is holding
	Buffs            buffs.Buffs                    `yaml:"buffs,omitempty"`         // The buffs the character has active
	Equipment        Worn                           `yaml:"equipment,omitempty"`     // The equipment the character is wearing
	TNLScale         float32                        `yaml:"-"`                       // The experience scale of the character. Don't write to yaml since is dynamically calculated.
	HealthMax        stats.StatInfo                 `yaml:"-"`                       // The maximum health of the character. Don't write to yaml since is dynamically calculated.
	ManaMax          stats.StatInfo                 `yaml:"-"`                       // The maximum mana of the character. Don't write to yaml since is dynamically calculated.
	ActionPointsMax  stats.StatInfo                 `yaml:"-"`                       // The maximum actions of character. Don't write to yaml since is dynamically calculated.
	Aggro            *Aggro                         `yaml:"-"`                       // Dont' store this. If they leave they break their aggro
	Skills           map[string]int                 `yaml:"skills,omitempty"`        // The skills the character has, and what level they are at
	Cooldowns        Cooldowns                      `yaml:"cooldowns,omitempty"`     // How many rounds until it is cooled down
	Settings         map[string]string              `yaml:"settings,omitempty"`      // custom setting tracking, used for anything.
	QuestProgress    map[int]string                 `yaml:"questprogress,omitempty"` // quest progress tracking
	KeyRing          map[string]string              `yaml:"keyring,omitempty"`       // key is the lock id, value is the sequence
	KD               KDStats                        `yaml:"kd,omitempty"`            // Kill/Death stats
	MiscData         map[string]any                 `yaml:"miscdata,omitempty"`      // Any random other data that needs to be stored
	ExtraLives       int                            `yaml:"extralives,omitempty"`    // How many lives remain. If enabled, players can perma-die if they die at zero
	MobMastery       MobMasteries                   `yaml:"mobmastery,omitempty"`    // Tracks particular masteries around a given mob
	Pet              pets.Pet                       `yaml:"pet,omitempty"`           // Do they have a pet?
	Created          time.Time                      `yaml:"created"`                 // When this character was created
	Timers           map[string]gametime.RoundTimer `yaml:"timers,omitempty"`        // any special timers added to this character
	roomHistory      []int                          // A stack FILO of the last X rooms the character has been in
	PlayerDamage     map[int]int                    `yaml:"-"` // key = who, value = how much
	LastPlayerDamage uint64                         `yaml:"-"` // last round a player damaged this character
	permaBuffIds     []int                          // Buff Id's that are always present for this character
	userId           int                            // User ID of the character if any
}

func New() *Character {
	statsConfig := configs.GetStatisticsConfig()

	return &Character{
		//Name:   defaultName,
		Adjectives: []string{},
		RoomId:     StartingRoomId,
		Zone:       startingZone,
		RaceId:     startingRace,
		Stats: stats.Statistics{
			Strength:   stats.StatInfo{Base: int(statsConfig.BaseStats.Strength)},
			Speed:      stats.StatInfo{Base: int(statsConfig.BaseStats.Speed)},
			Smarts:     stats.StatInfo{Base: int(statsConfig.BaseStats.Smarts)},
			Vitality:   stats.StatInfo{Base: int(statsConfig.BaseStats.Vitality)},
			Mysticism:  stats.StatInfo{Base: int(statsConfig.BaseStats.Mysticism)},
			Perception: stats.StatInfo{Base: int(statsConfig.BaseStats.Perception)},
		},
		Level:          1,
		Experience:     1,
		TrainingPoints: 0,
		StatPoints:     0,
		TNLScale:       1.0,
		Health:         startingHealth,
		HealthMax:      stats.StatInfo{Base: 1},
		Mana:           startingMana,
		ManaMax:        stats.StatInfo{Base: 1},
		Skills:         make(map[string]int),
		Gold:           25,
		Bank:           100,
		SpellBook:      make(map[string]int),
		CharmedMobs:    []int{},
		Items:          []items.Item{},
		Buffs:          buffs.New(),
		Equipment:      Worn{},
		MiscData:       make(map[string]any),
		roomHistory:    make([]int, 0, 10),
		KeyRing:        make(map[string]string),
		Created:        time.Now(),
		PlayerDamage:   map[int]int{},
		Timers:         map[string]gametime.RoundTimer{},
	}
}

// returns description unless description is a hash
// which points to another description location.
func (c *Character) GetDescription() string {

	if !strings.HasPrefix(c.Description, `h:`) {
		return c.Description
	}
	hash := strings.TrimPrefix(c.Description, `h:`)
	return descriptionCache[hash]
}

/*
All spells should have a 10% minimum chance of success.
*/

// Sometimes it's useful for a character to know what user it belongs to.
func (c *Character) SetUserId(userId int) {
	c.userId = userId
}

func (c *Character) SetMiscData(key string, value any) {

	if c.MiscData == nil {
		c.MiscData = make(map[string]any)
	}

	if value == nil {
		delete(c.MiscData, key)
		return
	}
	c.MiscData[key] = value
}

func (c *Character) GetMiscData(key string) any {

	if c.MiscData == nil {
		c.MiscData = make(map[string]any)
	}

	if value, ok := c.MiscData[key]; ok {
		return value
	}
	return nil
}

func (c *Character) GetMiscDataKeys(prefixMatch ...string) []string {

	if c.MiscData == nil {
		c.MiscData = make(map[string]any)
	}

	allKeys := []string{}
	for key := range c.MiscData {
		allKeys = append(allKeys, key)
	}

	if len(prefixMatch) == 0 {
		return allKeys
	}

	retKeys := []string{}
	for _, prefix := range prefixMatch {
		for _, key := range allKeys {
			if finalKey, ok := strings.CutPrefix(key, prefix); ok {
				retKeys = append(retKeys, finalKey)
			}
		}
	}

	return retKeys
}

func (c *Character) FindKeyInBackpack(lockId string) (items.Item, bool) {

	lockId = strings.ToLower(lockId)

	for _, itm := range c.GetAllBackpackItems() {
		itmSpec := itm.GetSpec()
		if itmSpec.Type != items.Key {
			continue
		}

		if itmSpec.KeyLockId == lockId {
			return itm, true
		}
	}

	return items.Item{}, false
}

func (c *Character) HasKey(lockId string, difficulty int) (hasKey bool, hasSequence bool) {

	sequence := util.GetLockSequence(lockId, difficulty, string(configs.GetServerConfig().Seed))

	// Check whether they ahve a key for this lock
	return c.GetKey(`key-`+lockId) != ``, c.GetKey(lockId) == sequence
}

func (c *Character) KeyCount() int {
	if c.KeyRing == nil {
		c.KeyRing = make(map[string]string)
	}
	return len(c.KeyRing)
}

func (c *Character) GetKey(lockId string) string {
	if c.KeyRing == nil {
		c.KeyRing = make(map[string]string)
	}
	return c.KeyRing[strings.ToLower(lockId)]
}

func (c *Character) SetKey(lockId string, sequence string) {
	if c.KeyRing == nil {
		c.KeyRing = make(map[string]string)
	}
	if len(sequence) == 0 {
		delete(c.KeyRing, strings.ToLower(lockId))
	} else {
		c.KeyRing[strings.ToLower(lockId)] = strings.ToUpper(sequence)
	}
}

// This should only be used for mobs.
// Not players
func (c *Character) CacheDescription() {
	// Hash the descriptions and store centrally.
	// This saves a lot of memory because many descriptions are duplicates
	hash := util.Hash(c.Description)
	if _, ok := descriptionCache[hash]; !ok {
		descriptionCache[hash] = c.Description
	}
	c.Description = fmt.Sprintf(`h:%s`, hash)
}

func (c *Character) GrantXP(xp int) (actualXP int, xpScale int) {

	if xp == 0 {
		return 0, 100
	}

	preScale := float64(configs.GetGamePlayConfig().XPScale) / 100
	xp = int(math.Round(preScale * float64(xp)))

	xpScale = c.StatMod(string(statmods.XPScale)) + 100

	if xpScale == 100 {
		actualXP = xp
	} else {

		scaleFloat := max(float64(xpScale)/100, 1)

		actualXP = int(float64(xp) * scaleFloat)
	}

	c.Experience += actualXP

	mudlog.Debug(`GrantXP()`, `username`, c.Name, `xp`, xp, `xpscale`, xpScale, `actualXP`, actualXP)

	return actualXP, xpScale
}

func (c *Character) TrackCharmed(mobId int, add bool) {
	for pos, mobInstanceId := range c.CharmedMobs {
		if mobInstanceId == mobId {
			if !add {
				c.CharmedMobs = slices.Delete(c.CharmedMobs, pos, pos+1)
			}
			return
		}
	}
	c.CharmedMobs = append(c.CharmedMobs, mobId)
}

func (c *Character) GetCharmIds() []int {
	return append([]int{}, c.CharmedMobs...)
}

func (c *Character) Charm(userId int, rounds int, expireCommand string) {
	c.SetAdjective(`charmed`, true)
	c.Charmed = NewCharm(userId, rounds, expireCommand)
	if c.Aggro != nil && c.Aggro.UserId == userId {
		c.Aggro = nil
	}
}

func (c *Character) KnowsFirstAid() bool {
	if r := races.GetRace(c.RaceId); r != nil {
		return r.KnowsFirstAid
	}
	return false
}

func (c *Character) GetCharmedUserId() int {
	if c.Charmed != nil {
		return c.Charmed.UserId
	}
	return 0
}

func (c *Character) IsCharmed(userId ...int) bool {

	if c.Charmed == nil {
		return false
	}

	if len(userId) == 0 {
		return c.Charmed != nil
	}

	if c.Charmed == nil {
		return false
	}
	return slices.Contains(userId, c.Charmed.UserId)
}

// Returns userId of whoever had charmed them
func (c *Character) RemoveCharm() int {
	charmUserId := 0
	c.SetAdjective(`charmed`, false)
	if c.Charmed != nil {
		charmUserId = c.Charmed.UserId
		c.Charmed = nil
	}
	return charmUserId
}

func (c *Character) GetMobName(viewingUserId int, renderFlags ...NameRenderFlag) FormattedName {
	return c.getFormattedName(viewingUserId, `mobname`, renderFlags...)
}

func (c *Character) GetPlayerName(viewingUserId int, renderFlags ...NameRenderFlag) FormattedName {
	return c.getFormattedName(viewingUserId, `username`, renderFlags...)
}

func (c *Character) HasAdjective(adj string) bool {
	return slices.Contains(c.Adjectives, adj)
}

func (c *Character) SetAdjective(adj string, addToList bool) {
	if c.Adjectives == nil {
		c.Adjectives = []string{}
	}
	for i, a := range c.Adjectives {
		if a == adj {
			if addToList {
				return
			} else {
				c.Adjectives = slices.Delete(c.Adjectives, i, i+1)
				return
			}
		}
	}
	if addToList {
		c.Adjectives = append(c.Adjectives, adj)
	}
}

func (c *Character) GetAdjectives() []string {

	retAdjectives := []string{}

	// Start dynamic adjectives
	if c.Health < 1 {
		retAdjectives = append(retAdjectives, `downed`)
	}

	if len(c.Shop) > 0 {
		retAdjectives = append(retAdjectives, `shop`)
	}

	if c.HasBuffFlag(buffs.EmitsLight) {
		retAdjectives = append(retAdjectives, `lit`)
	}

	if c.HasBuffFlag(buffs.Hidden) {
		retAdjectives = append(retAdjectives, `hidden`)
	}

	if c.HasBuffFlag(buffs.Poison) {
		retAdjectives = append(retAdjectives, `poisoned`)
	}
	// End dynamic adjectives

	retAdjectives = append(retAdjectives, c.Adjectives...)

	return retAdjectives
}

func (c *Character) getFormattedName(viewingUserId int, uType string, renderFlags ...NameRenderFlag) FormattedName {

	f := FormattedName{
		Name:       c.Name,
		Type:       uType,
		Adjectives: make([]string, 0, len(c.Adjectives)),
	}

	includeHealth := false
	for _, flag := range renderFlags {
		if flag == RenderHealth {
			includeHealth = true
		} else if flag == RenderShortAdjectives {
			f.UseShortAdjectives = true
		}
	}

	// If including health, only do so if not downed, because downed shows as its own adjective.
	if includeHealth && c.Health > 0 {
		pctHealth := int(math.Ceil(float64(c.Health) / float64(c.HealthMax.GetValue()) * 100))
		f.Adjectives = append(f.Adjectives, strconv.Itoa(pctHealth)+`%`)
	}

	f.Adjectives = append(f.Adjectives, c.GetAdjectives()...)

	if c.Health < 1 {
		f.Suffix = `downed`
	} else if c.Aggro != nil && c.Aggro.UserId == viewingUserId {
		f.Suffix = `aggro`
	}

	if c.Pet.Exists() {
		f.PetName = c.Pet.DisplayName()
	}

	return f
}

func (c *Character) PruneCooldowns() {
	if len(c.Cooldowns) == 0 {
		return
	}

	c.Cooldowns.Prune()
}

func (c *Character) GetCooldown(trackingTag string) int {
	if c.Cooldowns == nil {
		c.Cooldowns = make(Cooldowns)
	}
	return c.Cooldowns[trackingTag]
}

func (c *Character) GetAllCooldowns() map[string]int {

	ret := map[string]int{}

	if c.Cooldowns == nil {
		return ret
	}

	maps.Copy(ret, c.Cooldowns)

	return ret
}

func (c *Character) TryCooldown(trackingTag string, cooldownTime string) bool {
	if c.Cooldowns == nil {
		c.Cooldowns = make(Cooldowns)
	}

	return c.Cooldowns.Try(trackingTag, cooldownTime)
}

func (c *Character) SetSetting(settingName string, settingValue string) {
	if c.Settings == nil {
		c.Settings = make(map[string]string)
	}

	if settingValue == "" {
		delete(c.Settings, settingName)
	} else {
		c.Settings[settingName] = settingValue
	}
}

func (c *Character) GetSetting(settingName string) string {
	if c.Settings == nil {
		c.Settings = make(map[string]string)
	}
	if settingValue, ok := c.Settings[settingName]; ok {
		return settingValue
	}
	return ""
}

func (c *Character) SetAggroRemote(exitName string, userId int, mobInstanceId int, aggroType AggroType, roundsWaitTime ...int) {
	c.SetAggro(userId, mobInstanceId, aggroType, roundsWaitTime...)
	c.Aggro.ExitName = exitName
}

func (c *Character) SetAggro(userId int, mobInstanceId int, aggroType AggroType, roundsWaitTime ...int) {

	var combatAddlWaitRounds int = 0

	if len(roundsWaitTime) > 0 {
		for _, waitAmt := range roundsWaitTime {
			combatAddlWaitRounds += waitAmt
		}
	} else {
		combatAddlWaitRounds = c.Equipment.Weapon.GetSpec().WaitRounds + c.Equipment.Offhand.GetSpec().WaitRounds
	}

	if aggroType == DefaultAttack {
		if c.Equipment.Weapon.GetSpec().Subtype == items.Shooting {
			aggroType = Shooting
		}
	}

	c.Aggro = &Aggro{
		UserId:        userId,
		MobInstanceId: mobInstanceId,
		Type:          aggroType,
		RoundsWaiting: combatAddlWaitRounds,
	}

}

func (c *Character) SetCast(roundsWaitTime int, sInfo SpellAggroInfo) {

	c.Aggro = &Aggro{
		Type:          SpellCast,
		RoundsWaiting: roundsWaitTime,
		SpellInfo:     sInfo,
	}

}

func (c *Character) EndAggro() {
	c.Aggro = nil
}

func (c *Character) IsAggro(targetUserId int, targetMobInstanceId int) bool {

	if c.Aggro != nil {

		if c.Aggro.MobInstanceId > 0 && c.Aggro.MobInstanceId == targetMobInstanceId {
			return true
		}

		if c.Aggro.UserId > 0 && c.Aggro.UserId == targetUserId {
			return true
		}

		if c.Aggro.Type == SpellCast {
			if len(c.Aggro.SpellInfo.TargetUserIds) > 0 {
				for _, uId := range c.Aggro.SpellInfo.TargetUserIds {
					if uId == targetUserId {
						return true
					}
				}
			}

			if len(c.Aggro.SpellInfo.TargetMobInstanceIds) > 0 {
				for _, mId := range c.Aggro.SpellInfo.TargetMobInstanceIds {
					if mId == targetMobInstanceId {
						return true
					}
				}
			}
		}

	}
	return false
}

func (c *Character) IsDisabled() bool {
	return c.Health <= 0
}

func (c *Character) HasBuffFlag(buffFlag buffs.Flag) bool {
	return c.Buffs.HasFlag(buffFlag, false)
}

func (c *Character) CancelBuffsWithFlag(buffFlag buffs.Flag) bool {
	if c.Buffs.HasFlag(buffFlag, true) {
		c.Validate(true)
		return true
	}
	return false
}

func (c *Character) HasBuff(buffId int) bool {
	return c.Buffs.HasBuff(buffId)
}

func (c *Character) AddBuff(buffId int, isPermanent bool) error {
	buffId = int(math.Abs(float64(buffId)))
	if !c.Buffs.AddBuff(buffId, isPermanent) {
		return fmt.Errorf(`failed to add buff. target: "%s" buffId: %d`, c.Name, buffId)
	}
	c.Validate()
	return nil
}

func (c *Character) TrackBuffStarted(buffId int) {
	c.Buffs.Started(buffId)
}

func (c *Character) GetBuffs(buffId ...int) []*buffs.Buff {
	return c.Buffs.GetBuffs(buffId...)
}

func (c *Character) RemoveBuff(buffId int) {
	buffId = int(math.Abs(float64(buffId)))
	c.Buffs.RemoveBuff(buffId)
	c.Validate()
}

func (c *Character) TimerSet(name, period string) {
	if c.Timers == nil {
		c.Timers = map[string]gametime.RoundTimer{}
	}
	c.Timers[name] = gametime.RoundTimer{
		RoundStart: util.GetRoundCount(),
		Period:     period,
	}
}

func (c *Character) TimerExpired(name string) bool {
	if c.Timers == nil {
		return true
	}

	t, ok := c.Timers[name]

	if !ok {
		return true
	}

	if t.Expired() {
		delete(c.Timers, name)
		return true
	}

	return false
}

func (c *Character) TimerExists(name string) bool {
	if c.Timers == nil {
		return false
	}

	_, ok := c.Timers[name]
	return ok
}

func (c *Character) XPTNL() int {
	return c.XPTL(c.Level)
}

// Amt TNL for a specific level
func (c *Character) XPTL(lvl int) int {
	if lvl < 1 {
		lvl = 1
	}
	fLvl := float64(lvl)
	return int(float32(1000+(fLvl*(fLvl*.75)*1000)) * c.TNLScale)
}

// Returns the actual xp in regards to the current level/next level
func (c *Character) XPTNLActual() (xpPastCurrentLevel int, tnlXP int) {

	xpForCurrentLevel := c.XPTL(c.Level - 1)
	if c.Level == 1 {
		xpForCurrentLevel = 0
	}

	xpForNextLevel := c.XPTL(c.Level)
	tnlXP = xpForNextLevel - xpForCurrentLevel

	xpPastCurrentLevel = c.Experience - xpForCurrentLevel

	return xpPastCurrentLevel, tnlXP
}

func (c *Character) LevelUp() (bool, stats.Statistics) {

	if c.XPTNL() > c.Experience {
		return false, stats.Statistics{}
	}

	var statsBefore stats.Statistics = c.Stats

	c.Level++
	c.TrainingPoints++
	c.StatPoints++

	c.Validate()

	var statsDelta stats.Statistics = c.Stats

	for _, name := range c.Stats.GetStatInfoNames() {
		statsDelta.SetValue(name, statsDelta.GetValue(name)-statsBefore.GetValue(name))
	}

	c.Health = c.HealthMax.GetValue()
	c.Mana = c.ManaMax.GetValue()

	return true, statsDelta
}

func (c *Character) StatMod(statName string) int {
	return c.Equipment.StatMod(statName) + c.Buffs.StatMod(statName) + c.Pet.StatMod(statName)
}

// returns true if something has changed.
// TODO: [nitpick] There are many repetitive Get("X") calls; consider iterating over a slice of stat names or using a helper to apply updates in a loop for readability.

func (c *Character) RecalculateStats() {

	// Make sure racial base stats are set
	beforeHealthMax := c.HealthMax
	beforeManaMax := c.ManaMax
	beforeStats := c.Stats

	if raceInfo := races.GetRace(c.RaceId); raceInfo != nil {
		c.TNLScale = raceInfo.TNLScale
		// Safety check: ensure TNLScale is never 0
		if c.TNLScale == 0 {
			c.TNLScale = 1.0
		}
		for _, statName := range c.Stats.GetStatInfoNames() {
			c.Stats.SetBase(statName, raceInfo.Stats.GetBase(statName))
		}
	}

	// Add any mods for equipment
	for _, statName := range c.Stats.GetStatInfoNames() {
		c.Stats.SetMod(statName, c.StatMod(statName))
	}

	// Recalculate stats
	// Stats are basically:
	// level*base + training + mods
	for _, statName := range c.Stats.GetStatInfoNames() {
		c.Stats.Get(statName).Recalculate(c.Level)
	}

	// Set HP/MP maxes
	// This relies on the above stats so has to be calculated afterwards
	c.HealthMax.SetMod(5 +
		c.StatMod(string(statmods.HealthMax)) + // Any sort of spell buffs etc. are just direct modifiers
		c.Level + // For every level you get 1 hp
		c.Stats.ActionValueAdj("MaxHealth")*4) // for every vitality you get 3hp

	c.ManaMax.SetMod(4 +
		c.StatMod(string(statmods.ManaMax)) + // Any sort of spell buffs etc. are just direct modifiers
		c.Level + // For every level you get 1 mp
		c.Stats.ActionValueAdj("MaxMana")*3) // for every Mysticism you get 2mp

	// Set max action points
	c.ActionPointsMax.SetMod(200) // hard coded for now

	// Recalculate HP/MP stats
	c.HealthMax.Recalculate(c.Level)
	c.ManaMax.Recalculate(c.Level)
	c.ActionPointsMax.Recalculate(c.Level)

	// HP can't max less than 1, MP can't max less than 0
	if c.ManaMax.GetValue() < 0 {
		c.ManaMax.SetValue(0)
	}
	if c.HealthMax.GetValue() < 1 {
		c.HealthMax.SetValue(1)
	}
	if c.ActionPointsMax.GetValue() < 50 {
		c.ActionPointsMax.SetValue(50)
	}

	if c.userId != 0 {
		changed := false
		// return true if something has changed.
		for _, statName := range c.Stats.GetStatInfoNames() {
			if beforeStats.GetValueAdj(statName) != c.Stats.GetValueAdj(statName) {
				changed = true
				break
			}
		}
		if !changed {
			if beforeHealthMax != c.HealthMax || beforeManaMax != c.ManaMax {
				changed = true
			}
		}

		if changed {
			events.AddToQueue(events.CharacterStatsChanged{UserId: c.userId})
		}
	}

}

// Returns whether a correction was in order
func (c *Character) Validate(recalcPermaBuffs ...bool) error {

	if len(c.Description) == 0 {
		c.Description = "They seem thoroughly uninteresting."
	}

	if race := races.GetRace(c.RaceId); race == nil {
		c.RaceId = 1
	}

	if c.Created.IsZero() {
		c.Created = time.Now()
	}

	if c.Pet.Exists() {
		c.Pet.Validate()
	}

	if c.SpellBook == nil {
		c.SpellBook = make(map[string]int)
	}

	if c.Zone == "" {
		c.Zone = startingZone
	}

	if c.Name == "" {
		c.Name = defaultName
	}
	if c.Level < 1 {
		c.Level = 1
	}
	if c.Experience < 1 {
		c.Experience = 1
	}

	c.Buffs.Validate()

	// Do a stats recalc based on equipment, race, level, etc.
	c.RecalculateStats()

	// Recalculate health and mana

	if c.Mana > c.ManaMax.GetValue() {
		c.Mana = c.ManaMax.GetValue()
	}
	if c.Health > c.HealthMax.GetValue() {
		c.Health = c.HealthMax.GetValue()
	}

	if c.Health < -10 {
		c.Health = -10
	}

	if c.Mana < 0 {
		c.Mana = 0
	}

	c.Cooldowns.Prune()

	if c.Alignment < AlignmentMinimum {
		c.Alignment = AlignmentMinimum
	}

	if c.Alignment > AlignmentMaximum {
		c.Alignment = AlignmentMaximum
	}

	// Validate possessed/worn items
	// This helps ensure all in-play items have a uid
	for i := range c.Items {
		c.Items[i].Validate()
	}
	c.Equipment.Weapon.Validate()
	c.Equipment.Offhand.Validate()
	c.Equipment.Head.Validate()
	c.Equipment.Neck.Validate()
	c.Equipment.Body.Validate()
	c.Equipment.Belt.Validate()
	c.Equipment.Gloves.Validate()
	c.Equipment.Ring.Validate()
	c.Equipment.Legs.Validate()
	c.Equipment.Feet.Validate()
	// Done with validation

	if raceInfo := races.GetRace(c.RaceId); raceInfo != nil {

		c.Equipment.EnableAll()

		// Are there slots that SHOULD be disabled?
		if len(raceInfo.DisabledSlots) > 0 {

			for _, disabledSlot := range raceInfo.DisabledSlots {

				var itemFoundInDisabledSlot items.Item = items.ItemDisabledSlot

				switch items.ItemType(disabledSlot) {
				case items.Weapon:
					if c.Equipment.Weapon.ItemId > 0 { // Did we find somethign in a disabled slot?
						itemFoundInDisabledSlot = c.Equipment.Weapon
					}
					c.Equipment.Weapon = items.ItemDisabledSlot
				case items.Offhand:
					if c.Equipment.Offhand.ItemId > 0 { // Did we find somethign in a disabled slot?
						itemFoundInDisabledSlot = c.Equipment.Offhand
					}
					c.Equipment.Offhand = items.ItemDisabledSlot
				case items.Head:
					if c.Equipment.Head.ItemId > 0 { // Did we find somethign in a disabled slot?
						itemFoundInDisabledSlot = c.Equipment.Head
					}
					c.Equipment.Head = items.ItemDisabledSlot
				case items.Neck:
					if c.Equipment.Neck.ItemId > 0 { // Did we find somethign in a disabled slot?
						itemFoundInDisabledSlot = c.Equipment.Neck
					}
					c.Equipment.Neck = items.ItemDisabledSlot
				case items.Body:
					if c.Equipment.Body.ItemId > 0 { // Did we find somethign in a disabled slot?
						itemFoundInDisabledSlot = c.Equipment.Body
					}
					c.Equipment.Body = items.ItemDisabledSlot
				case items.Belt:
					if c.Equipment.Belt.ItemId > 0 { // Did we find somethign in a disabled slot?
						itemFoundInDisabledSlot = c.Equipment.Belt
					}
					c.Equipment.Belt = items.ItemDisabledSlot
				case items.Gloves:
					if c.Equipment.Gloves.ItemId > 0 { // Did we find somethign in a disabled slot?
						itemFoundInDisabledSlot = c.Equipment.Gloves
					}
					c.Equipment.Gloves = items.ItemDisabledSlot
				case items.Ring:
					if c.Equipment.Ring.ItemId > 0 { // Did we find somethign in a disabled slot?
						itemFoundInDisabledSlot = c.Equipment.Ring
					}
					c.Equipment.Ring = items.ItemDisabledSlot
				case items.Legs:
					if c.Equipment.Legs.ItemId > 0 { // Did we find somethign in a disabled slot?
						itemFoundInDisabledSlot = c.Equipment.Legs
					}
					c.Equipment.Legs = items.ItemDisabledSlot
				case items.Feet:
					if c.Equipment.Feet.ItemId > 0 { // Did we find somethign in a disabled slot?
						itemFoundInDisabledSlot = c.Equipment.Feet
					}
					c.Equipment.Feet = items.ItemDisabledSlot
				}

				if !itemFoundInDisabledSlot.IsDisabled() {
					c.StoreItem(itemFoundInDisabledSlot)
					mudlog.Debug("Disabled Check", "error", "Item found in disabled slot", "name", itemFoundInDisabledSlot.Name(), "slot", disabledSlot, "character", c.Name)
				}
			}

		}

	}

	if len(recalcPermaBuffs) > 0 && recalcPermaBuffs[0] {
		c.reapplyPermabuffs()
	}

	return nil
}

func (c *Character) Race() string {
	if r := races.GetRace(c.RaceId); r != nil {
		return r.Name
	}
	return `Ghostly Spirit`
}

func (c *Character) UpdateAlignment(amt int) {
	newAlignment := int(c.Alignment) + amt
	if newAlignment < int(AlignmentMinimum) {
		newAlignment = int(AlignmentMinimum)
	} else if newAlignment > int(AlignmentMaximum) {
		newAlignment = int(AlignmentMaximum)
	}
	c.Alignment = int8(newAlignment)
}

func (c *Character) AlignmentName() string {
	return AlignmentToString(c.Alignment)
}
