package characters

import (
	"fmt"

	"github.com/GoMudEngine/GoMud/internal/buffs"
	"github.com/GoMudEngine/GoMud/internal/statmods"
	"github.com/GoMudEngine/GoMud/internal/util"
)

// returns description unless description is a hash
// which points to another description location.
func (c *Character) TrackPlayerDamage(userId int, damageAmt int) {

	roundNow := util.GetRoundCount()
	if len(c.PlayerDamage) == 0 {
		c.PlayerDamage = map[int]int{}
	} else {
		if roundNow-c.LastPlayerDamage > 30 {
			clear(c.PlayerDamage)
		}
	}

	c.PlayerDamage[userId] = c.PlayerDamage[userId] + damageAmt
	c.LastPlayerDamage = roundNow

}

func (c *Character) DeductActionPoints(amount int) bool {

	if c.ActionPoints < amount {
		return false
	}
	c.ActionPoints -= amount
	if c.ActionPoints < 0 {
		c.ActionPoints = 0
	}
	return true
}

// USERNAME appears to be <BLANK>
func (c *Character) GetHealthAppearance() string {

	className := util.HealthClass(c.Health, c.HealthMax.GetValue())
	pct := int(float64(c.Health) / float64(c.HealthMax.GetValue()) * 100)

	if pct < 15 {
		return fmt.Sprintf(`<ansi fg="username">%s</ansi> looks like they're <ansi fg="%s">about to die!</ansi>`, c.Name, className)
	}

	if pct < 50 {
		return fmt.Sprintf(`<ansi fg="username">%s</ansi> looks to be in <ansi fg="%s">pretty bad shape.</ansi>`, c.Name, className)
	}

	if pct < 80 {
		return fmt.Sprintf(`<ansi fg="username">%s</ansi> has some <ansi fg="%s">cuts and bruises.</ansi>`, c.Name, className)
	}

	if pct < 100 {
		return fmt.Sprintf(`<ansi fg="username">%s</ansi> has <ansi fg="%s">a few scratches.</ansi>`, c.Name, className)
	}

	return fmt.Sprintf(`<ansi fg="username">%s</ansi> is in <ansi fg="%s">perfect health.</ansi>`, c.Name, className)
}

func (c *Character) ApplyHealthChange(healthChange int) int {
	oldHealth := c.Health
	newHealth := c.Health + healthChange
	if newHealth < 0 {
		c.CancelBuffsWithFlag(buffs.CancelIfCombat)

		// If they haven't dropped yet, require a drop before going straight to death.
		// Don't allow players to drop under -5 in a single hit.
		if newHealth < -5 && oldHealth > 0 {
			newHealth = -5
		} else if newHealth <= -10 {
			newHealth = -10
		}
	} else if newHealth > c.HealthMax.GetValue() {
		newHealth = c.HealthMax.GetValue()
	}

	c.Health = newHealth

	return newHealth - oldHealth
}

func (c *Character) ApplyManaChange(manaChange int) int {
	oldMana := c.Mana
	c.Mana += manaChange
	if c.Mana < 0 {
		c.Mana = 0
	} else if c.Mana > c.ManaMax.GetValue() {
		c.Mana = c.ManaMax.GetValue()
	}
	return c.Mana - oldMana
}

func (c *Character) BarterPrice(startPrice int) int {
	factor := (float64(c.Stats.ActionValueAdj("BarterDiscount")) / 3) / 100 // 100 = 33% discount, 0 = 0% discount, 300 = 100% discount
	if factor > .75 {
		factor = .75
	}
	return int(factor * float64(startPrice))
}

func (c *Character) Heal(hp int, mana int) (int, int) {
	startHP := c.Health
	startMP := c.Mana

	c.Health += hp
	if c.Health > c.HealthMax.GetValue() {
		c.Health = c.HealthMax.GetValue()
	}
	c.Mana += hp
	if c.Mana > c.ManaMax.GetValue() {
		c.Mana = c.ManaMax.GetValue()
	}

	return c.Health - startHP, c.Mana - startMP
}

func (c *Character) HealthPerRound() int {
	return 1 + c.StatMod(string(statmods.HealthRecovery))
	/*
		healAmt := math.Round(float64(c.Stats.Get("Vitality").ValueAdj)/8) +
			math.Round(float64(c.Level)/12) +
			1.0

		return int(healAmt)
	*/
}

func (c *Character) ManaPerRound() int {
	return 1 + c.StatMod(string(statmods.ManaRecovery))
	/*
		healAmt := math.Round(float64(c.Stats.Get("Mysticism").ValueAdj)/8) +
			math.Round(float64(c.Level)/12) +
			1.0

		return int(healAmt)
	*/
}

// Where 1000 = a full round
func (c *Character) MovementCost() int {
	modifier := 3                                       // by default they should be able to move 3 times per round.
	modifier += int(c.Level / 15)                       // Every 15 levels, get an extra movement.
	modifier += int(c.Stats.ActionValueAdj("MovementSpeed") / 15) // Every 15 speed, get an extra movement
	return int(1000 / modifier)
}
