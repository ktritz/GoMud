package configs

type Statistics struct {
	BaseStats   StatDefaults               `yaml:"BaseStats"`
	Factors     StatFactors                `yaml:"Factors"`
	StatActions map[string]StatActionBlock `yaml:"StatActions"` // Maps stat names to their associated actions
}

// StatActionBlock groups all actions associated with a single stat.
type StatActionBlock struct {
	Actions []string `yaml:"Actions"`
}

type StatDefaults struct {
	Strength   ConfigInt `yaml:"Strength"`
	Speed      ConfigInt `yaml:"Speed"`
	Smarts     ConfigInt `yaml:"Smarts"`
	Vitality   ConfigInt `yaml:"Vitality"`
	Mysticism  ConfigInt `yaml:"Mysticism"`
	Perception ConfigInt `yaml:"Perception"`
}

type StatFactors struct {
	BaseModFactor         ConfigFloat `yaml:"BaseModFactor"`         // How much of a scaling to aply to levels before multiplying by racial stat
	NaturalGainsModFactor ConfigFloat `yaml:"NaturalGainsModFactor"` // Free stats gained per level modded by this.
}

// Default stat action mappings grouped by stat.
var defaultStatActions = map[string][]string{
	"Strength": {
		"DiceSides",      // Additional damage dice sides
		"CritPrimary",    // Critical hit chance (primary)
		"CarryCapacity",  // How many items can be carried
		"DisarmDefense1", // Resist disarm (primary)
		"BumpOffense",    // Bump/push success
	},
	"Speed": {
		"DiceCount",       // Additional damage dice count
		"AttackCount",     // Extra attacks in combat
		"HitChance",       // Chance to land a hit
		"CritSecondary",   // Critical hit chance (secondary)
		"FleeChance",      // Chance to flee from combat
		"MovementSpeed",   // Extra movements per round
		"TackleOffense",   // Tackle success chance
		"DisarmOffense1",  // Disarm success (primary)
		"Pickpocket1",     // Pickpocket success (primary)
		"MobFollowSpeed",  // Speed component of mob follow
	},
	"Smarts": {
		"DisarmOffense2",  // Disarm success (secondary)
		"Pickpocket2",     // Pickpocket success (secondary)
		"MapMemory",       // Room memory capacity
		"MapSprawl",       // Map generation range
		"TameGrowth2",     // Taming skill growth (secondary)
	},
	"Vitality": {
		"MaxHealth", // Contribution to maximum health
	},
	"Mysticism": {
		"MaxMana",           // Contribution to maximum mana
		"Casting",           // Spell casting success chance
		"EnchantPower",      // Enchantment damage/defense/stat bonus
		"EnchantProtection", // Reduction in enchant item destruction
		"PortalDuration",    // How long portals last
		"PrayerPower",       // Number of prayer buffs received
	},
	"Perception": {
		"DamageBonus",       // Bonus damage per hit
		"TackleDefense",     // Resist being tackled
		"DisarmDefense2",    // Resist disarm (secondary)
		"Pickpocket3",       // Pickpocket success (tertiary)
		"PickpocketDefense", // Resist pickpocket
		"SearchSuccess",     // Search for hidden things
		"BarterDiscount",    // Negotiate better prices
		"MobFollowDefense",  // Mob chase awareness
		"HostilityReduction",// Reduce hostile mob timer
		"MobWeaponSearch",   // Mob AI weapon pickup chance
		"TameGrowth1",       // Taming skill growth (primary)
	},
}

// actionLookup is a resolved map from action -> stat, built during Validate.
var actionLookup map[string]string

func (s *Statistics) Validate() {
	if s.BaseStats.Strength < 1 {
		s.BaseStats.Strength = 1
	}
	if s.BaseStats.Speed < 1 {
		s.BaseStats.Speed = 1
	}
	if s.BaseStats.Smarts < 1 {
		s.BaseStats.Smarts = 1
	}
	if s.BaseStats.Vitality < 1 {
		s.BaseStats.Vitality = 1
	}
	if s.BaseStats.Mysticism < 1 {
		s.BaseStats.Mysticism = 1
	}
	if s.BaseStats.Perception < 1 {
		s.BaseStats.Perception = 1
	}

	// Validate factors are reasonable
	if s.Factors.BaseModFactor <= 0 {
		s.Factors.BaseModFactor = 0.3333333334 // default
	}
	if s.Factors.NaturalGainsModFactor <= 0 {
		s.Factors.NaturalGainsModFactor = 0.5 // default
	}

	// Build the action -> stat lookup from defaults first
	actionLookup = make(map[string]string)
	for stat, actions := range defaultStatActions {
		for _, action := range actions {
			actionLookup[action] = stat
		}
	}

	// Override with any configured stat action blocks
	for stat, block := range s.StatActions {
		for _, action := range block.Actions {
			actionLookup[action] = stat
		}
	}
}

func GetStatisticsConfig() Statistics {
	configDataLock.RLock()
	defer configDataLock.RUnlock()

	if !configData.validated {
		configData.Validate()
	}
	return configData.Statistics
}

// GetStatForAction returns the stat name configured for a given action.
func GetStatForAction(action string) string {
	configDataLock.RLock()
	defer configDataLock.RUnlock()

	if !configData.validated {
		configData.Validate()
	}

	if stat, ok := actionLookup[action]; ok {
		return stat
	}
	return ""
}
