package locksmith

import (
	"fmt"

	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/items"
)

// KeyResult describes the result of attempting to find a key for a lock.
type KeyResult struct {
	HasKeyRingKey  bool       // Player has the key on their key ring
	HasBackpackKey bool       // Player has a key item in backpack
	BackpackKey    items.Item // The key item found in backpack (if any)
	HasSequence    bool       // Player knows the lock sequence (for lockpicking)
}

// BuildLockId creates a standardized lock identifier from a room ID and target name.
func BuildLockId(roomId int, targetName string) string {
	return fmt.Sprintf(`%d-%s`, roomId, targetName)
}

// FindKey checks the character's key ring and backpack for a key matching the lock.
func FindKey(char *characters.Character, lockId string, difficulty int) KeyResult {
	hasKey, hasSequence := char.HasKey(lockId, difficulty)

	result := KeyResult{
		HasKeyRingKey: hasKey,
		HasSequence:   hasSequence,
	}

	if !hasKey {
		backpackKey, found := char.FindKeyInBackpack(lockId)
		if found {
			result.HasBackpackKey = true
			result.BackpackKey = backpackKey
		}
	}

	return result
}

// ConsumeBackpackKey moves a key from the character's backpack to their key ring
// and emits an ItemOwnership event.
func ConsumeBackpackKey(char *characters.Character, keyItem items.Item, lockId string, userId int) {
	char.SetKey(`key-`+lockId, fmt.Sprintf(`%d`, keyItem.ItemId))
	char.RemoveItem(keyItem)

	events.AddToQueue(events.ItemOwnership{
		UserId: userId,
		Item:   keyItem,
		Gained: false,
	})
}
