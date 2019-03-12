package dispatcher

import (
	. "github.com/iotaledger/iota.go/trinary"
	"github.com/lunfardo314/goq/utils"
)

type EntityCore interface {
	Call(Trits, Trits) bool
}

type Entity struct {
	dispatcher *Dispatcher
	name       string
	inSize     int64
	outSize    int64
	affecting  []*environment // list of affected environments where effects are sent
	joined     []*environment // list of environments which are being listened to
	inChan     chan Trits     // chan for incoming effects
	entityCore EntityCore     // function called for each effect
}

func NewEntity(disp *Dispatcher, name string, inSize, outSize int64, effectCallable EntityCore) *Entity {
	ret := &Entity{
		dispatcher: disp,
		name:       name,
		inSize:     inSize,
		outSize:    outSize,
		affecting:  make([]*environment, 0),
		joined:     make([]*environment, 0),
		entityCore: effectCallable,
	}
	return ret
}

func (ent *Entity) GetName() string {
	return ent.name
}

func (ent *Entity) InSize() int64 {
	return ent.inSize
}

func (ent *Entity) OutSize() int64 {
	return ent.outSize
}

func (ent *Entity) affectEnvironment(env *environment) {
	ent.affecting = append(ent.affecting, env)
}

func (ent *Entity) joinEnvironment(env *environment) {
	ent.joined = append(ent.joined, env)
	ent.checkStart()
}

func (ent *Entity) stopAffectingEnvironment(env *environment) {
	tmpList := make([]*environment, 0)
	for _, e := range ent.affecting {
		if e != env {
			tmpList = append(tmpList, e)
		}
	}
}

func (ent *Entity) stopListeningToEnvironment(env *environment) {
	tmpList := make([]*environment, 0)
	for _, e := range ent.joined {
		if e != env {
			tmpList = append(tmpList, e)
		}
	}
	ent.joined = tmpList
	ent.checkStop()
}

func (ent *Entity) checkStop() {
	if ent.inChan != nil && len(ent.joined) == 0 {
		c := ent.inChan
		ent.inChan = nil
		close(c)
	}
}

func (ent *Entity) checkStart() {
	if ent.inChan == nil && len(ent.joined) != 0 {
		ent.inChan = make(chan Trits)
		go ent.effectsLoop()
	}
}

func (ent *Entity) effectsLoop() {
	logf(4, "entity '%v': effects loop STARTED", ent.name)
	defer logf(4, "entity '%v': effects loop STOPPED", ent.name)

	res := make(Trits, ent.outSize)

	for effect := range ent.inChan {
		logf(2, "Entity '%v' <- '%v'", ent.name, utils.TritsToString(effect))
		// calculate result
		if !ent.entityCore.Call(effect, res) {
			// is not null
			// mark it is done with entity
			// distribute result to affected environments
			for _, env := range ent.affecting {
				env.postEffect(res)
			}
		}
		ent.dispatcher.holdWaveWG.Done()
		ent.dispatcher.quantWG.Done()
		logf(4, "---------------- DONE (entity '%v')", ent.name)
	}
}
