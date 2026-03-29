package rooms

import (
	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/util"
)

func (r *Room) MarkVisited(id int, vType VisitorType, subtrackTurns ...int) {

	if r.visitors == nil {
		r.visitors = make(map[VisitorType]map[int]uint64)
	}

	if _, ok := r.visitors[vType]; !ok {
		r.visitors[vType] = make(map[int]uint64)
	}

	lastSeen := util.GetTurnCount() + uint64(visitorTrackingTimeout*configs.GetTimingConfig().TurnsPerSecond())

	if len(subtrackTurns) > 0 {
		if uint64(subtrackTurns[0]) > lastSeen {
			lastSeen = 0
		} else {
			lastSeen -= uint64(subtrackTurns[0])
		}
	}

	r.visitors[vType][id] = lastSeen
	r.lastVisited = util.GetRoundCount()
}

// Returns a list of recent visitors and how cold the trail is getting
func (r *Room) Visitors(vType VisitorType) map[int]float64 {

	ret := make(map[int]float64)

	tps := configs.GetTimingConfig().TurnsPerSecond()
	if _, ok := r.visitors[vType]; ok {
		for userId, expires := range r.visitors[vType] {
			ret[userId] = float64(expires-util.GetTurnCount()) / float64(visitorTrackingTimeout*tps)
		}

	}

	return ret
}

func (r *Room) HasVisited(id int, vType VisitorType) bool {
	//	r.PruneVisitors()

	if _, ok := r.visitors[vType]; !ok {
		return false
	}

	_, ok := r.visitors[vType][id]

	return ok
}

func (r *Room) HasRecentVisitors() bool {

	return r.visitors != nil && len(r.visitors) > 0
}

func (r *Room) PruneVisitors() int {

	if r.visitors == nil {
		r.visitors = make(map[VisitorType]map[int]uint64)
		return 0
	}

	c := configs.GetTimingConfig()

	// Make sure whoever is here has the freshest mark.
	for _, userId := range r.players {
		if _, ok := r.visitors[VisitorUser]; ok {
			r.visitors[VisitorUser][userId] = util.GetTurnCount() + uint64(visitorTrackingTimeout*c.TurnsPerSecond())
		}
	}

	for _, mobId := range r.mobs {
		if _, ok := r.visitors[VisitorMob]; ok {
			r.visitors[VisitorMob][mobId] = util.GetTurnCount() + uint64(visitorTrackingTimeout*c.TurnsPerSecond())
		}
	}

	pruneCt := 0

	for vType, _ := range r.visitors {

		for id, expires := range r.visitors[vType] {
			// Check whether expires is older than now
			if expires < util.GetTurnCount() {
				delete(r.visitors[vType], id)
				pruneCt++

				if len(r.visitors[vType]) < 1 {
					delete(r.visitors, vType)
				}
			}
		}

	}
	return pruneCt
}
