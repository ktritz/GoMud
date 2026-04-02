package keywords

// InitTestKeywords initializes the keywords system with empty maps
// for testing without loading data files.
func InitTestKeywords() {
	loadedKeywords = &Aliases{
		helpTopics:         map[string]HelpTopic{},
		helpAliases:        map[string]string{},
		commandAliases:     map[string]string{},
		mapLegendOverrides: map[string]map[rune]string{},
	}

	// Add common aliases that handlers expect
	loadedKeywords.commandAliases["l"] = "look"
	loadedKeywords.commandAliases["i"] = "inventory"
	loadedKeywords.commandAliases["inv"] = "inventory"
	loadedKeywords.commandAliases["eq"] = "inventory"
	loadedKeywords.commandAliases["a"] = "attack"
	loadedKeywords.commandAliases["k"] = "attack"
	loadedKeywords.commandAliases["kill"] = "attack"
	loadedKeywords.commandAliases["g"] = "get"
	loadedKeywords.commandAliases["take"] = "get"
	loadedKeywords.commandAliases["sta"] = "status"
	loadedKeywords.commandAliases["stat"] = "status"
	loadedKeywords.commandAliases["stats"] = "status"
	loadedKeywords.commandAliases["score"] = "status"
	loadedKeywords.commandAliases["q"] = "quests"
	loadedKeywords.commandAliases["quest"] = "quests"
	loadedKeywords.commandAliases["examine"] = "look"
	loadedKeywords.commandAliases["enter"] = "go"
	loadedKeywords.commandAliases["wear"] = "equip"
	loadedKeywords.commandAliases["wield"] = "equip"
	loadedKeywords.commandAliases["grab"] = "get"
}
