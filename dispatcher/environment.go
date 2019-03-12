package dispatcher

import (
	"fmt"
	. "github.com/iotaledger/iota.go/trinary"
	. "github.com/lunfardo314/goq/utils"
	"sync"
)

type environment struct {
	sync.RWMutex
	dispatcher *Dispatcher
	name       string
	invalid    bool
	joins      []*Entity
	affects    []*Entity
	size       int64
	effectChan chan Trits
	value      Trits // valid only between waves
}

func NewEnvironment(disp *Dispatcher, name string) *environment {
	ret := &environment{
		dispatcher: disp,
		name:       name,
		joins:      make([]*Entity, 0),
		affects:    make([]*Entity, 0),
		effectChan: make(chan Trits),
	}
	go ret.effectsLoop()
	return ret
}

//func (env *environment) Size() int64 {
//	return env.size
//}
//
func (env *environment) GetName() string {
	return env.name
}

func (env *environment) existsEntity_(name string) bool {
	for _, ei := range env.joins {
		if ei.name == name {
			return true
		}
	}
	return false
}

func (env *environment) checkNewSize(size int64) bool {
	if env.size != 0 {
		if env.size != size {
			return false
		}
	} else {
		env.size = size
	}
	return true
}

func (env *environment) join(entity *Entity) error {
	if !env.checkNewSize(entity.InSize()) {
		return fmt.Errorf("size mismach between joining entity '%v' and the environment '%v'",
			entity.name, env.name)
	}
	env.joins = append(env.joins, entity)
	entity.joinEnvironment(env)
	return nil
}

func (env *environment) affect(entity *Entity) error {
	if !env.checkNewSize(entity.OutSize()) {
		return fmt.Errorf("size mismach between affecting entity '%v' and the environment '%v'",
			entity.name, env.name)
	}
	env.affects = append(env.affects, entity)
	entity.affectEnvironment(env)
	return nil
}

func (env *environment) postEffect(effect Trits) {
	if effect != nil {
		dec, _ := TritsToBigInt(effect)
		logf(2, "environment '%v' <- '%v' (%v)", env.name, TritsToString(effect), dec)
	} else {
		logf(2, "environment '%v' <- 'null'", env.name)
	}
	env.setNewValue(effect)
	env.dispatcher.quantWG.Add(len(env.joins))

	logf(4, "---------------- ADD %v (env '%v')", len(env.joins), env.name)

	env.effectChan <- effect
}

// loop waits for effect in the environment and then process it
func (env *environment) effectsLoop() {
	logf(4, "environment '%v': effects loop STARTED", env.name)
	defer logf(4, "environment '%v': effects loop STOPPED", env.name)

	for effect := range env.effectChan {
		// only passes when wave ends.
		env.dispatcher.holdWaveWG.Wait()

		// released externally.
		env.dispatcher.releaseWaveWG.Wait()

		if len(env.joins) == 0 {
			continue
		}
		//  here starts new wave
		env.dispatcher.holdWaveWG.Add(len(env.joins)) // <<<< ???????????

		env.setNewValue(nil) // environment value becomes invalid during wave
		for _, entity := range env.joins {
			entity.inChan <- effect
		}
	}
}

func (env *environment) setNewValue(val Trits) Trits {
	env.Lock()
	defer env.Unlock()
	logf(3, "------ SET value env '%v' = '%v'", env.name, TritsToString(val))
	saveValue := env.value
	env.value = val
	return saveValue
}

func (env *environment) GetValue() Trits {
	env.RLock()
	defer env.RUnlock()
	return env.value
}

func (env *environment) invalidate() {
	if env.invalid {
		return
	}
	env.invalid = true
	close(env.effectChan)

	for _, entity := range env.joins {
		entity.stopListeningToEnvironment(env)
	}
	for _, entity := range env.affects {
		entity.stopAffectingEnvironment(env)
	}
}
