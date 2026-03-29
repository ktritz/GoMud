package characters

import (
	"math"

	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/races"
)

func (c *Character) GetDefaultDiceRoll() (attacks int, dCount int, dSides int, bonus int, buffOnCrit []int) {
	// default racial
	raceInfo := races.GetRace(c.RaceId)

	attacks = raceInfo.Damage.Attacks
	dCount = raceInfo.Damage.DiceCount
	dSides = raceInfo.Damage.SideCount
	bonus = raceInfo.Damage.BonusDamage
	buffOnCrit = raceInfo.Damage.CritBuffIds

	dCount += int(math.Floor((float64(c.Stats.ActionValueAdj("DiceCount")) / 50)))
	dSides += int(math.Floor((float64(c.Stats.ActionValueAdj("DiceSides")) / 12)))
	bonus += int(math.Floor((float64(c.Stats.ActionValueAdj("DamageBonus")) / 25)))

	if dCount < raceInfo.Damage.DiceCount {
		dCount = raceInfo.Damage.DiceCount
	}
	if dSides < raceInfo.Damage.SideCount {
		dSides = raceInfo.Damage.SideCount
	}

	return attacks, dCount, dSides, bonus, buffOnCrit
}

// Returns an integer representing a % damage reduction
func (c *Character) GetDefense() int {

	reduction := c.Equipment.Weapon.GetDefense() +
		c.Equipment.Offhand.GetDefense() +
		c.Equipment.Head.GetDefense() +
		c.Equipment.Neck.GetDefense() +
		c.Equipment.Body.GetDefense() +
		c.Equipment.Belt.GetDefense() +
		c.Equipment.Gloves.GetDefense() +
		c.Equipment.Ring.GetDefense() +
		c.Equipment.Legs.GetDefense() +
		c.Equipment.Feet.GetDefense()

	//reduction = int(float64(reduction) / 9)

	// If wearing an offhand item like a shield, defense gets a 50% boost
	// Holdables are not considered "shield" type items.
	// Anything held in the offhand that provides a damage reduction is considered a shield.
	if c.Equipment.Offhand.ItemId != 0 && c.Equipment.Offhand.GetSpec().Type != items.Weapon && c.Equipment.Offhand.GetSpec().DamageReduction > 0 {
		reduction = int(float64(reduction) * 1.5)
	}

	if reduction > 100 {
		reduction = 100
	}

	return reduction
}

func (c *Character) GetAllWornItems() []items.Item {
	wornItems := []items.Item{}
	if c.Equipment.Weapon.ItemId > 0 {
		wornItems = append(wornItems, c.Equipment.Weapon)
	}
	if c.Equipment.Offhand.ItemId > 0 {
		wornItems = append(wornItems, c.Equipment.Offhand)
	}
	if c.Equipment.Head.ItemId > 0 {
		wornItems = append(wornItems, c.Equipment.Head)
	}
	if c.Equipment.Neck.ItemId > 0 {
		wornItems = append(wornItems, c.Equipment.Neck)
	}
	if c.Equipment.Body.ItemId > 0 {
		wornItems = append(wornItems, c.Equipment.Body)
	}
	if c.Equipment.Belt.ItemId > 0 {
		wornItems = append(wornItems, c.Equipment.Belt)
	}
	if c.Equipment.Gloves.ItemId > 0 {
		wornItems = append(wornItems, c.Equipment.Gloves)
	}
	if c.Equipment.Ring.ItemId > 0 {
		wornItems = append(wornItems, c.Equipment.Ring)
	}
	if c.Equipment.Legs.ItemId > 0 {
		wornItems = append(wornItems, c.Equipment.Legs)
	}
	if c.Equipment.Feet.ItemId > 0 {
		wornItems = append(wornItems, c.Equipment.Feet)
	}
	return wornItems
}

func (c *Character) GetGearValue() int {
	value := 0
	if c.Equipment.Weapon.ItemId > 0 {
		value += c.Equipment.Weapon.GetSpec().Value
	}
	if c.Equipment.Offhand.ItemId > 0 {
		value += c.Equipment.Offhand.GetSpec().Value
	}
	if c.Equipment.Head.ItemId > 0 {
		value += c.Equipment.Head.GetSpec().Value
	}
	if c.Equipment.Neck.ItemId > 0 {
		value += c.Equipment.Neck.GetSpec().Value
	}
	if c.Equipment.Body.ItemId > 0 {
		value += c.Equipment.Body.GetSpec().Value
	}
	if c.Equipment.Belt.ItemId > 0 {
		value += c.Equipment.Belt.GetSpec().Value
	}
	if c.Equipment.Gloves.ItemId > 0 {
		value += c.Equipment.Gloves.GetSpec().Value
	}
	if c.Equipment.Ring.ItemId > 0 {
		value += c.Equipment.Ring.GetSpec().Value
	}
	if c.Equipment.Legs.ItemId > 0 {
		value += c.Equipment.Legs.GetSpec().Value
	}
	if c.Equipment.Feet.ItemId > 0 {
		value += c.Equipment.Feet.GetSpec().Value
	}
	return value
}

func (c *Character) Wear(i items.Item) (returnItems []items.Item, newItemWorn bool, failureReason string) {

	i.Validate()

	spec := i.GetSpec()

	if spec.Type != items.Weapon && spec.Subtype != items.Wearable {
		return returnItems, false, `That item cannot be equipped.`
	}

	iHandsRequired := c.HandsRequired(i)
	if iHandsRequired > 2 {
		return returnItems, false, `That requires too many hands.`
	}

	// are botht he currently equipped weapon and this weapon claws?
	bothMartial := false
	if spec.Subtype == items.Claws && c.Equipment.Weapon.GetSpec().Subtype == items.Claws {
		bothMartial = true
	}

	canDualWield := c.CanDualWield()

	// Weapons can go in either hand.
	// Only do this if this is a 1 handed weapon
	if spec.Type == items.Weapon && iHandsRequired < 2 {

		// If they can dual wield
		if canDualWield || bothMartial {

			// If they have a weapon equippment and it is 1 handed
			if c.Equipment.Weapon.ItemId != 0 && c.HandsRequired(c.Equipment.Weapon) == 1 {
				// If nothing is in their offhand
				if c.Equipment.Offhand.ItemId == 0 {
					// Put it in the offhand.
					//returnItems = append(returnItems, c.Equipment.Offhand)
					c.Equipment.Offhand = i

					c.reapplyPermabuffs()

					return returnItems, true, ``
				}
			}

		}

	}

	// First handle weapon/offhand, since they are special cases
	switch spec.Type {
	case items.Weapon:
		if c.Equipment.Weapon.IsDisabled() { // Don't allow equipping on a disabled slot
			return returnItems, false, `You can't use a weapon.`
		}

		if !c.Equipment.Offhand.IsDisabled() { // Don't allow equipping on a disabled slot
			// If it's a 2 handed weapon, remove whatever is in the offhand
			if iHandsRequired == 2 || !canDualWield && c.Equipment.Offhand.GetSpec().Type == items.Weapon {
				returnItems = append(returnItems, c.Equipment.Offhand)
				c.Equipment.Offhand = items.Item{}
			}
		}

		if c.Equipment.Weapon.IsCursed() {
			return returnItems, false, `Your ` + c.Equipment.Weapon.DisplayName() + ` is cursed and prevents you from removing it.`
		}

		returnItems = append(returnItems, c.Equipment.Weapon)
		c.Equipment.Weapon = i
	case items.Offhand:
		if c.Equipment.Offhand.IsDisabled() { // Don't allow equipping on a disabled slot
			return returnItems, false, `You can't hold things in an offhand.`
		}

		if !c.Equipment.Weapon.IsDisabled() { // Don't allow equipping on a disabled slot
			// If they have a 2h weapon equipped, remove it
			if c.HandsRequired(c.Equipment.Weapon) == 2 {
				// If the weapon is cursed, do not allow the offhand to be equipped
				if c.Equipment.Weapon.IsCursed() {
					return returnItems, false, `Your ` + c.Equipment.Weapon.DisplayName() + ` is cursed and prevents you from removing it.`
				}
				returnItems = append(returnItems, c.Equipment.Weapon)
				c.Equipment.Weapon = items.Item{}
			}
		}
		returnItems = append(returnItems, c.Equipment.Offhand)
		c.Equipment.Offhand = i
	case items.Head:
		if c.Equipment.Head.IsDisabled() { // Don't allow equipping on a disabled slot
			return returnItems, false, `You can't wear things on your head.`
		}
		returnItems = append(returnItems, c.Equipment.Head)
		c.Equipment.Head = i
	case items.Neck:
		if c.Equipment.Neck.IsDisabled() { // Don't allow equipping on a disabled slot
			return returnItems, false, `You can't wear things on your neck.`
		}
		returnItems = append(returnItems, c.Equipment.Neck)
		c.Equipment.Neck = i
	case items.Body:
		if c.Equipment.Body.IsDisabled() { // Don't allow equipping on a disabled slot
			return returnItems, false, `You can't wear things on your body.`
		}
		returnItems = append(returnItems, c.Equipment.Body)
		c.Equipment.Body = i
	case items.Belt:
		if c.Equipment.Belt.IsDisabled() { // Don't allow equipping on a disabled slot
			return returnItems, false, `You can't wear things on your head.`
		}
		returnItems = append(returnItems, c.Equipment.Belt)
		c.Equipment.Belt = i
	case items.Gloves:
		if c.Equipment.Gloves.IsDisabled() { // Don't allow equipping on a disabled slot
			return returnItems, false, `You can't wear things as gloves.`
		}
		returnItems = append(returnItems, c.Equipment.Gloves)
		c.Equipment.Gloves = i
	case items.Ring:
		if c.Equipment.Ring.IsDisabled() { // Don't allow equipping on a disabled slot
			return returnItems, false, `You can't wear rings.`
		}
		returnItems = append(returnItems, c.Equipment.Ring)
		c.Equipment.Ring = i
	case items.Legs:
		if c.Equipment.Legs.IsDisabled() { // Don't allow equipping on a disabled slot
			return returnItems, false, `You can't wear things on your legs.`
		}
		returnItems = append(returnItems, c.Equipment.Legs)
		c.Equipment.Legs = i
	case items.Feet:
		if c.Equipment.Feet.IsDisabled() { // Don't allow equipping on a disabled slot
			return returnItems, false, `You can't wear things on your feet.`
		}
		returnItems = append(returnItems, c.Equipment.Feet)
		c.Equipment.Feet = i
	default:
		return returnItems, false, `Unrecognized object.`
	}

	c.reapplyPermabuffs(returnItems...)

	return returnItems, true, ``
}

func (c *Character) RemoveFromBody(i items.Item) bool {

	if i.Equals(c.Equipment.Weapon) {
		c.Equipment.Weapon = items.Item{}
	} else if i.Equals(c.Equipment.Offhand) {
		c.Equipment.Offhand = items.Item{}
	} else if i.Equals(c.Equipment.Head) {
		c.Equipment.Head = items.Item{}
	} else if i.Equals(c.Equipment.Neck) {
		c.Equipment.Neck = items.Item{}
	} else if i.Equals(c.Equipment.Body) {
		c.Equipment.Body = items.Item{}
	} else if i.Equals(c.Equipment.Belt) {
		c.Equipment.Belt = items.Item{}
	} else if i.Equals(c.Equipment.Gloves) {
		c.Equipment.Gloves = items.Item{}
	} else if i.Equals(c.Equipment.Ring) {
		c.Equipment.Ring = items.Item{}
	} else if i.Equals(c.Equipment.Legs) {
		c.Equipment.Legs = items.Item{}
	} else if i.Equals(c.Equipment.Feet) {
		c.Equipment.Feet = items.Item{}
	} else {
		return false
	}

	c.reapplyPermabuffs(i)

	return true
}

// Used with SpawnInfo to gift spawning mobs with permabuffs
func (c *Character) SetPermaBuffs(buffIds []int) {
	c.permaBuffIds = buffIds
}

func (c *Character) reapplyPermabuffs(removedItems ...items.Item) {

	buffIdCount := map[int]int{}

	for _, buffId := range c.permaBuffIds {
		buffIdCount[buffId] = 100 // Special case permabuffs associated with certain mobs
	}

	// Apply any buffs that come from a race
	if rInfo := races.GetRace(c.RaceId); rInfo != nil {
		for _, buffId := range rInfo.BuffIds {
			buffIdCount[buffId] = 100 // Don't allow racial buffs to be removed, keep this number high
		}
	}

	// Apply any buffs from pet
	if c.Pet.Exists() {
		for _, buffId := range c.Pet.GetBuffs() {
			buffIdCount[buffId] = 100 // Don't allow pet buffs to be removed, keep this number high
		}
	}

	// Track any buffs that come from an item
	// If these don't show up as still being required by an item (such as a yaml file was changed)
	// This will cause them to be removed.
	for _, b := range c.Buffs.List {
		if b.PermaBuff {
			if _, ok := buffIdCount[b.BuffId]; !ok {
				buffIdCount[b.BuffId] = 0
			}
		}
	}

	// Make a list of all item buffs provided by existing worn items
	for _, itm := range c.GetAllWornItems() {
		spec := itm.GetSpec()
		for _, buffId := range spec.WornBuffIds {
			buffIdCount[buffId] = buffIdCount[buffId] + 1
		}

	}
	// Remove any buffs that come specifically from item
	for _, removedItem := range removedItems {
		iSpec := removedItem.GetSpec()
		if len(iSpec.WornBuffIds) > 0 {
			for _, buffId := range iSpec.WornBuffIds {
				buffIdCount[buffId] = buffIdCount[buffId] - 1
			}
		}
	}

	for buffId, ct := range buffIdCount {
		if ct < 1 {
			c.RemoveBuff(buffId)
		} else {
			c.AddBuff(buffId, true)
		}
	}
}

func (c *Character) Uncurse() []items.Item {

	uncursedList := []items.Item{}

	if c.Equipment.Weapon.IsCursed() {
		c.Equipment.Weapon.Uncursed = true
		uncursedList = append(uncursedList, c.Equipment.Weapon)
	}

	if c.Equipment.Offhand.IsCursed() {
		c.Equipment.Offhand.Uncursed = true
		uncursedList = append(uncursedList, c.Equipment.Offhand)
	}

	if c.Equipment.Head.IsCursed() {
		c.Equipment.Head.Uncursed = true
		uncursedList = append(uncursedList, c.Equipment.Head)
	}

	if c.Equipment.Neck.IsCursed() {
		c.Equipment.Neck.Uncursed = true
		uncursedList = append(uncursedList, c.Equipment.Neck)
	}

	if c.Equipment.Body.IsCursed() {
		c.Equipment.Body.Uncursed = true
		uncursedList = append(uncursedList, c.Equipment.Body)
	}

	if c.Equipment.Belt.IsCursed() {
		c.Equipment.Belt.Uncursed = true
		uncursedList = append(uncursedList, c.Equipment.Belt)
	}

	if c.Equipment.Gloves.IsCursed() {
		c.Equipment.Gloves.Uncursed = true
		uncursedList = append(uncursedList, c.Equipment.Gloves)
	}

	if c.Equipment.Ring.IsCursed() {
		c.Equipment.Ring.Uncursed = true
		uncursedList = append(uncursedList, c.Equipment.Ring)
	}

	if c.Equipment.Legs.IsCursed() {
		c.Equipment.Legs.Uncursed = true
		uncursedList = append(uncursedList, c.Equipment.Legs)
	}

	if c.Equipment.Feet.IsCursed() {
		c.Equipment.Feet.Uncursed = true
		uncursedList = append(uncursedList, c.Equipment.Feet)
	}

	return uncursedList
}
