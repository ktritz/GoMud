package mobs

// Test helpers for registering mobs in the global mob registry.
// Used by the test harness package for integration testing.

// RegisterTestMobInstance registers a mob instance in the global registry.
func RegisterTestMobInstance(mob *Mob) {
	if mobInstances == nil {
		mobInstances = make(map[int]*Mob)
	}
	mobInstances[mob.InstanceId] = mob
}

// DeregisterTestMobInstance removes a mob instance from the global registry.
func DeregisterTestMobInstance(instanceId int) {
	delete(mobInstances, instanceId)
}

// NewTestMob creates a minimal Mob for testing.
func NewTestMob(instanceId int, mobId MobId, name string) *Mob {
	m := &Mob{
		MobId:         mobId,
		InstanceId:    instanceId,
		Hostile:       false,
		tempDataStore: make(map[string]any),
	}
	m.Character.Name = name
	m.Character.Gold = 100
	m.Character.Health = 50
	return m
}
