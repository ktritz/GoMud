package rooms

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/exit"
	"github.com/GoMudEngine/GoMud/internal/gametime"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/mutators"
	"github.com/GoMudEngine/GoMud/internal/users"
	"github.com/GoMudEngine/GoMud/internal/util"
)

const visitorTrackingTimeout = 180 // 180 seconds (3 minutes?)
const defaultMapSymbol = `•`

var (
	MapSymbolOverrides = map[string]string{
		"*": defaultMapSymbol,
		//"•": "*",
	}
)

type FindFlag uint16
type VisitorType string

const (
	AffectsNone   = ""
	AffectsPlayer = "player" // Does it affect only the player who triggered it?
	AffectsRoom   = "room"   // Does it affect everyone in the room?

	// Useful for finding mobs/players
	FindCharmed        FindFlag = 0b00000000001 // charmed
	FindNeutral        FindFlag = 0b00000000010 // Not aggro, not charmed, not Hostile
	FindFightingPlayer FindFlag = 0b00000000100 // aggro vs. a player
	FindFightingMob    FindFlag = 0b00000001000 // aggro vs. a mob
	FindHostile        FindFlag = 0b00000010000 // will auto-attack players
	FindMerchant       FindFlag = 0b00000100000 // is a merchant
	FindDowned         FindFlag = 0b00001000000 // hp < 1
	FindBuffed         FindFlag = 0b00010000000 // has a buff
	FindHasLight       FindFlag = 0b00100000000 // has a light source
	FindHasPet         FindFlag = 0b01000000000 // has a pet
	FindNative         FindFlag = 0b10000000000 // spawns in this room

	// Combinatorial flags
	FindFighting          = FindFightingPlayer | FindFightingMob // Currently in combat (aggro)
	FindIdle              = FindCharmed | FindNeutral            // Not aggro or hostile
	FindAll      FindFlag = 0b111111111
	// Visitor types
	VisitorUser = "user"
	VisitorMob  = "mob"
)

type Room struct {
	//mutex
	RoomId            int                               `yaml:"roomid"`                              // a unique numeric index of the room. Also the filename.
	Zone              string                            `yaml:"zone"`                                // zone is a way to partition rooms into groups. Also into folders.
	MusicFile         string                            `yaml:"musicfile,omitempty"`                 // background music to play when in this room
	IsBank            bool                              `yaml:"isbank,omitempty"`                    // Is this a bank room? If so, players can deposit/withdraw gold here.
	IsStorage         bool                              `yaml:"isstorage,omitempty"`                 // Is this a storage room? If so, players can add/remove objects here.
	IsCharacterRoom   bool                              `yaml:"ischaracterroom,omitempty"`           // Is this a room where characters can create new characters to swap between them?
	Title             string                            `yaml:"title"`                               // Title shown to the user
	Description       string                            `yaml:"description"`                         // Description shown to the user
	MapSymbol         string                            `yaml:"mapsymbol,omitempty"`                 // The symbol to use when generating a map of the zone
	MapLegend         string                            `yaml:"maplegend,omitempty"`                 // The text to display in the legend for this room. Should be one word.
	Biome             string                            `yaml:"biome,omitempty"`                     // The biome of the room. Used for weather generation.
	Containers        map[string]Container              `yaml:"containers,omitempty"`                // If this room has a chest, what is in it?
	Exits             map[string]exit.RoomExit          `yaml:"exits"`                               // Exits to other rooms
	ExitsTemp         map[string]exit.TemporaryRoomExit `yaml:"-"`                                   // Temporary exits that will be removed after a certain time. Don't bother saving on sever shutting down.
	Nouns             map[string]string                 `yaml:"nouns,omitempty"`                     // Interesting nouns to highlight in the room or reveal on succesful searches.
	Items             []items.Item                      `yaml:"items,omitempty"`                     // Items on the floor
	Stash             []items.Item                      `yaml:"stash,omitempty"`                     // list of items in the room that are not visible to players
	Corpses           []Corpse                          `yaml:"-"`                                   // Any corpses laying around from recent deaths
	Gold              int                               `yaml:"gold,omitempty"`                      // How much gold is on the ground?
	SpawnInfo         []SpawnInfo                       `yaml:"spawninfo,omitempty" instance:"skip"` // key is creature ID, value is spawn chance
	SkillTraining     map[string]TrainingRange          `yaml:"skilltraining,omitempty"`             // list of skills that can be trained in this room
	Signs             []Sign                            `yaml:"sign,omitempty"`                      // list of scribbles in the room
	IdleMessages      []string                          `yaml:"idlemessages,omitempty" `             // list of messages that can be displayed to players in the room
	LastIdleMessage   uint8                             `yaml:"-"`                                   // index of the last idle message displayed
	LongTermDataStore map[string]any                    `yaml:"longtermdatastore,omitempty"`         // Long term data store for the room
	Mutators          mutators.MutatorList              `yaml:"mutators,omitempty"`                  // mutators this room spawns with.
	Pvp               bool                              `yaml:"pvp,omitempty"`                       // if config pvp is set to `limited`, uses this value
	// Unexported/private
	players       []int                          // list of user IDs currently in the room
	mobs          []int                          // list of mob instance IDs currently in the room. Does not get saved.
	visitors      map[VisitorType]map[int]uint64 // list of user IDs that have visited this room, and the last round they did
	lastVisited   uint64                         // last round a visitor was in the room
	tempDataStore map[string]any                 // Temporary data store for the room
}

type TrainingRange struct {
	Min int
	Max int
}

func NewRoom(zone string) *Room {
	r := &Room{
		RoomId:        GetNextRoomId(),
		Zone:          zone,
		Title:         "An empty room.",
		Description:   "This is an empty room that was never given a description.",
		MapSymbol:     ``,
		Exits:         make(map[string]exit.RoomExit),
		players:       []int{},
		visitors:      make(map[VisitorType]map[int]uint64),
		tempDataStore: make(map[string]any),
	}

	SetNextRoomId(r.RoomId + 1)

	return r
}

func (r *Room) IsEphemeral() bool {
	return r.RoomId >= ephemeralRoomIdMinimum
}

// 0 = none (darkness). 1 = can see this room. 2 = can see this room and all exits
func (r *Room) GetVisibility() int {

	visibility := 2 // default to max visibility
	// At night visibility decreases by one
	if gametime.IsNight() {
		visibility -= 1
	}

	biome := r.GetBiome()
	// First calculate natural lighting level for biome
	if biome.IsDark() { // If a naturally dark biome (cave), minimize visibility
		visibility -= 2
		if visibility < 0 {
			visibility = 0
		}
	} else if biome.IsLit() { // If the biome is naturally lit (streets with lanterns), increase visibility by one
		visibility += 1
		if visibility > 2 {
			visibility = 2
		}
	}

	// Apply any mutators
	for mut := range r.ActiveMutators {
		spec := mut.GetSpec()
		if spec.LightMod != 0 {
			visibility += spec.LightMod
		}
	}

	// min/max visibility
	if visibility < 0 {
		visibility = 0
	} else if visibility > 2 {
		visibility = 2
	}

	// If someone has light, cancel the darkness
	if visibility < 2 { // no need to increase light if it's already maxed
		if len(r.GetMobs(FindHasLight)) > 0 || len(r.GetPlayers(FindHasLight)) > 0 {
			visibility += 1
			if visibility > 2 {
				visibility = 2
			}
		}
	}

	return visibility
}

func (r *Room) SetLongTermData(key string, value any) {

	if r.LongTermDataStore == nil {
		r.LongTermDataStore = make(map[string]any)
	}

	if value == nil {
		delete(r.LongTermDataStore, key)
		return
	}
	r.LongTermDataStore[key] = value
}

func (r *Room) GetLongTermData(key string) any {

	if r.LongTermDataStore == nil {
		r.LongTermDataStore = make(map[string]any)
	}

	if value, ok := r.LongTermDataStore[key]; ok {
		return value
	}
	return nil
}

func (r *Room) SetTempData(key string, value any) {

	if r.tempDataStore == nil {
		r.tempDataStore = make(map[string]any)
	}

	if value == nil {
		delete(r.tempDataStore, key)
		return
	}
	r.tempDataStore[key] = value
}

func (r *Room) GetTempData(key string) any {

	if r.tempDataStore == nil {
		r.tempDataStore = make(map[string]any)
	}

	if value, ok := r.tempDataStore[key]; ok {
		return value
	}
	return nil
}

func (r *Room) GetScript() string {

	scriptPath := r.GetScriptPath()

	// Load the script into a string
	if _, err := os.Stat(scriptPath); err == nil {
		if bytes, err := os.ReadFile(scriptPath); err == nil {
			return string(bytes)
		}
	}

	return ``
}

func (r *Room) GetScriptPath() string {
	// Load any script for the room
	return strings.Replace(configs.GetFilePathsConfig().DataFiles.String()+`/rooms/`+r.Filepath(), `.yaml`, `.js`, 1)
}

// The purpose of Prepare() is to ensure a room is properly setup before anyone looks into it or enters it
// That way if there should be anything in the room prior, it will already be there.
// For example, mobs shouldn't ENTER the room right as the player arrives, they should already be there.
func (r *Room) Prepare(checkAdjacentRooms bool) {

	roundNow := util.GetRoundCount()

	r.Mutators.Update(roundNow)

	if len(r.Containers) > 0 {
		for k, c := range r.Containers {
			if c.DespawnRound > 0 && c.DespawnRound <= roundNow {
				r.SendText(fmt.Sprintf(`The <ansi fg="container">%s</ansi> crumbles to dust, and is gone.`, k))
				delete(r.Containers, k)
			}
		}
	}

	// First ensure any mobs that should be here are spawned
	for idx, spawnInfo := range r.SpawnInfo {

		// Make sure to clean up any instances that may be dead
		if spawnInfo.InstanceId > 0 {
			// Mob gone missing. Reset the spawn info.
			if mob := mobs.GetInstance(spawnInfo.InstanceId); mob == nil {
				spawnInfo.InstanceId = 0
				spawnInfo.DespawnedRound = roundNow
				r.SpawnInfo[idx] = spawnInfo
				continue
			}
			continue
		}

		// If a despawn was tracked, check whether the time has been reached, else skip
		if spawnInfo.DespawnedRound > 0 {

			if roundNow < gametime.GetDate(spawnInfo.DespawnedRound).AddPeriod(spawnInfo.RespawnRate) { // Not yet ready to respawn.
				continue
			}
		}

		//
		// At this point we are good to attempt respawns
		//

		// New instances needed? Spawn them
		if spawnInfo.MobId > 0 {

			forceLevel := 0

			if spawnInfo.Level > 0 {
				forceLevel = spawnInfo.Level
			} else {

				// Get the zone settings, check for scaling
				if zConfig := GetZoneConfig(r.Zone); zConfig != nil {

					if zConfig.MobAutoScale.Minimum > 0 {
						forceLevel = zConfig.GenerateRandomLevel()
					}

					if forceLevel > 0 {
						forceLevel += spawnInfo.LevelMod
						if forceLevel < 1 {
							forceLevel = 1
						}
					}

				}
			}

			if mob := mobs.NewMobById(mobs.MobId(spawnInfo.MobId), r.RoomId, forceLevel); mob != nil {

				// If a merchant, fill up stocks on first time being loaded in
				if mob.HasShop() {
					mob.Character.Shop.Restock()
				}

				if len(spawnInfo.BuffIds) > 0 {
					mob.Character.SetPermaBuffs(spawnInfo.BuffIds)
				}

				// If there are idle commands for this spawn, overwrite.
				if len(spawnInfo.IdleCommands) > 0 {
					mob.IdleCommands = append([]string{}, spawnInfo.IdleCommands...)
				}

				if len(spawnInfo.ScriptTag) > 0 {
					mob.ScriptTag = spawnInfo.ScriptTag
				}

				if len(spawnInfo.QuestFlags) > 0 {
					mob.QuestFlags = spawnInfo.QuestFlags
				}

				// Does this mob have a special name?
				if len(spawnInfo.Name) > 0 {
					mob.Character.Name = spawnInfo.Name
				}

				if spawnInfo.ForceHostile {
					mob.Hostile = true
				}

				if spawnInfo.MaxWander != 0 {
					mob.MaxWander = spawnInfo.MaxWander
				}

				mob.Character.Zone = r.Zone
				mob.Validate()

				r.mobs = append(r.mobs, mob.InstanceId)

				spawnInfo.InstanceId = mob.InstanceId
				spawnInfo.DespawnedRound = 0

				r.SpawnInfo[idx] = spawnInfo
			}

			roomManager.roomsWithMobs[r.RoomId] = len(r.mobs)

			// Since mob spanws cannot be combined with item/gold spawns, go next loop
			continue
		}

		if spawnInfo.ItemId > 0 || spawnInfo.Gold > 0 {

			// If no container specified, or the container specified exists, then spawn the item
			if spawnInfo.Container == `` {

				if _, alreadyExists := r.FindOnFloor(fmt.Sprintf(`!%d`, spawnInfo.ItemId), false); !alreadyExists {

					if item := items.New(spawnInfo.ItemId); item.ItemId != 0 {
						r.Items = append(r.Items, item) // just append to avoid a mutex double lock
					}

				}

				if r.Gold < spawnInfo.Gold {
					r.Gold = spawnInfo.Gold
				}

				spawnInfo.DespawnedRound = roundNow

				r.SpawnInfo[idx] = spawnInfo

				continue
			}

			if containerName := r.FindContainerByName(spawnInfo.Container); containerName != `` {

				container := r.Containers[containerName]

				if _, alreadyExists := container.FindItem(fmt.Sprintf(`!%d`, spawnInfo.ItemId)); !alreadyExists {
					if item := items.New(spawnInfo.ItemId); item.ItemId != 0 {
						container.AddItem(item)
					}
				}

				if container.Gold < spawnInfo.Gold {
					container.Gold = spawnInfo.Gold
				}

				r.Containers[containerName] = container

				spawnInfo.DespawnedRound = roundNow

				r.SpawnInfo[idx] = spawnInfo

			}

		}

	}

	// Reach out one more room to prepare those exit rooms
	if !checkAdjacentRooms {
		return
	}

	prepRoomIds := []int{}
	for _, exit := range r.Exits {
		if exit.RoomId == r.RoomId {
			continue
		}
		prepRoomIds = append(prepRoomIds, exit.RoomId)
	}

	for _, exitRoomId := range prepRoomIds {

		if exitRoom := LoadRoom(exitRoomId); exitRoom != nil {

			if exitRoom.PlayerCt() < 1 { // Don't prepare rooms that players are already in
				exitRoom.Prepare(false) // Don't continue checking adjacent rooms or else gets in recursion trouble
			}
		}

	}

}

func (r *Room) CleanupMobSpawns(noCooldown bool) {

	roundNow := util.GetRoundCount()
	// First ensure any mobs that should be here are spawned
	for idx, spawnInfo := range r.SpawnInfo {

		// Make sure to clean up any instances that may be dead
		if spawnInfo.InstanceId > 0 {

			if mob := mobs.GetInstance(spawnInfo.InstanceId); mob == nil {

				spawnInfo.InstanceId = 0
				if noCooldown {
					spawnInfo.DespawnedRound = 0
				} else {
					spawnInfo.DespawnedRound = roundNow
				}

			}
		}

		r.SpawnInfo[idx] = spawnInfo
	}
}

func (r *Room) GetDescriptionFormatted(lineSplit int, highlightNouns bool) string {

	desc := util.SplitStringNL(r.GetDescription(), 80)

	if highlightNouns {
		for noun, _ := range r.Nouns {
			desc = strings.ReplaceAll(desc, noun, fmt.Sprintf(`<ansi fg="noun">%s</ansi>`, noun))
		}
	}

	return desc
}

func (r *Room) GetDescription() string {
	return r.Description
}

func (r *Room) RoundTick() {

	roundNow := util.GetRoundCount()

	//
	// Apply any mutators from the zone or room
	// This will only add mutators that the player
	// doesn't already have.
	//
	r.Mutators.Update(roundNow)

	for mut := range r.ActiveMutators {
		spec := mut.GetSpec()
		r.ApplyBuffIdToPlayers(spec.PlayerBuffIds, `area`)
		r.ApplyBuffIdToMobs(spec.MobBuffIds, `area`)
		r.ApplyBuffIdToNativeMobs(spec.NativeBuffIds, `area`)
	}
	//
	// Done adding mutator buffs
	//

	for idx, spawnInfo := range r.SpawnInfo {

		// Make sure to clean up any instances that may be dead
		if spawnInfo.InstanceId > 0 {
			if mob := mobs.GetInstance(spawnInfo.InstanceId); mob == nil {
				spawnInfo.InstanceId = 0
				spawnInfo.DespawnedRound = roundNow
				r.SpawnInfo[idx] = spawnInfo
			}
		}

	}

	// If any players are in the room
	// Update all mobs in the room that they've seen a player
	if len(r.players) > 0 {
		for _, mobInstanceId := range r.mobs {
			if mob := mobs.GetInstance(mobInstanceId); mob != nil {
				mob.BoredomCounter = 0
			}
		}
	}

	//
	// Decay any corpses
	//
	r.UpdateCorpses(roundNow)
}

func (r *Room) Id() int {
	return r.RoomId
}

func (r *Room) Validate() error {
	if r.Title == "" {
		return errors.New("title cannot be empty")
	}
	if r.GetDescription() == "" {
		return errors.New("description cannot be empty")
	}

	if len(r.SpawnInfo) > 0 {

		for idx, sInfo := range r.SpawnInfo {

			// Make sure that mob spawns remain separately defined from item/gold spawns.
			if sInfo.MobId > 0 {
				if sInfo.ItemId > 0 || sInfo.Gold > 0 {
					return errors.New(`a given spawn info cannot have a mobid if it has gold or an item as well. Theese must be separate spawn info entries.`)
				}
			}

			// Spawn periods if left empty default to 15 minutes
			if sInfo.RespawnRate == `` {
				sInfo.RespawnRate = `15 real minutes`
				r.SpawnInfo[idx] = sInfo
			}

		}
	}


	// Make sure all items are validated (and have uids)
	for i := range r.Items {
		r.Items[i].Validate()
	}

	for i := range r.Stash {
		r.Stash[i].Validate()
	}

	for cName, c := range r.Containers {
		for i := range c.Items {
			c.Items[i].Validate()
		}
		r.Containers[cName] = c
	}

	return nil
}

func (r *Room) GetMapSymbol() string {
	if newSymbol, ok := MapSymbolOverrides[r.MapSymbol]; ok {
		return newSymbol
	}
	return r.MapSymbol
}

func (r *Room) Filename() string {
	return fmt.Sprintf("%d.yaml", GetOriginalRoom(r.RoomId))
}

func (r *Room) Filepath() string {
	zone := ZoneNameSanitize(r.Zone)
	return util.FilePath(zone, `/`, r.Filename())
}

func (r *Room) GetBiome() *BiomeInfo {

	if r.Biome == `` {
		if r.Zone != `` {
			r.Biome = GetZoneBiome(r.Zone)
		}
	}

	bInfo, ok := GetBiome(r.Biome)
	if !ok {
		// If biome not found, try to get the default biome
		bInfo, _ = GetBiome(``)
	}

	return bInfo
}

func (r *Room) ActiveMutators(yield func(mutators.Mutator) bool) {

	var activeMutators mutators.MutatorList
	if zoneConfig := GetZoneConfig(r.Zone); zoneConfig != nil {
		activeMutators = append(r.Mutators.GetActive(), zoneConfig.Mutators.GetActive()...)
	}

	for _, mut := range activeMutators {
		if !yield(mut) {
			return
		}
	}
}

// Returns true if Pvp is allowed in this room
func (r *Room) IsPvp() bool {
	roomPvp := r.Pvp
	for mut := range r.ActiveMutators {
		spec := mut.GetSpec()
		if spec.Pvp.Enabled {
			roomPvp = true
		} else if spec.Pvp.Disabled {
			roomPvp = false
		}
	}
	return roomPvp
}

// Returns an error with a reason why they cannot PVP, or nil
func (r *Room) CanPvp(attUser *users.UserRecord, defUser *users.UserRecord) error {

	if attUser.Character.RoomId == -1 || attUser.Character.RoomId == int(configs.GetSpecialRoomsConfig().DeathRecoveryRoom) {
		return errors.New(`Fighting is not allowed here.`)
	}

	c := configs.GetGamePlayConfig()

	// Possible settings are `enabled`, `disabled`, `limited`
	pvpSetting := string(c.PVP)
	minLevel := int(c.PVPMinimumLevel)

	if pvpSetting == configs.PVPDisabled {
		return errors.New(`PVP is disabled.`)
	}

	if attUser.Character.Level < minLevel || defUser.Character.Level < minLevel {
		return fmt.Errorf(`Players must be at least level %d to PVP.`, minLevel)
	}

	if pvpSetting == configs.PVPLimited {
		if r.IsPvp() {
			return nil
		}
		return errors.New(`This is not a PVP area.`)
	}

	return nil
}
