package repository

import (
	"sync/atomic"
)

type memoryMap map[string]map[string]*int64

func (v memoryMap) Get(experiment string, arm string) int64 {
	ctx, ok := v[experiment]
	if !ok {
		return 0
	}

	count, ok := ctx[arm]
	if !ok {
		return 0
	}

	return *count
}

func (v memoryMap) Incr(experiment string, arm string, add int) {
	ctx, ok := v[experiment]
	if !ok {
		ctx = make(map[string]*int64)
		v[experiment] = ctx
	}

	count, ok := ctx[arm]
	if !ok {
		var inital int64 = 0
		count = &inital
		ctx[arm] = count
	}
	if add == 0 {
		atomic.AddInt64(count, 1)
	} else {
		atomic.AddInt64(count, int64(add))
	}

}

type Memory struct {
	hits, rewards memoryMap
}

func NewMemory() Memory {
	return Memory{make(map[string]map[string]*int64), make(map[string]map[string]*int64)}
}

func (v Memory) Get(experiment string, arms []string) ExperimentData {
	var expData = ExperimentData{0, make(map[string]ArmData)}

	for _, arm := range arms {
		armData := ArmData{Hits: v.hits.Get(experiment, arm), Rewards: v.rewards.Get(experiment, arm)}

		expData.Arms[arm] = armData
		expData.TotalHits += armData.Hits
	}

	return expData
}

func (v Memory) Hit(experiment string, arm string) {
	v.hits.Incr(experiment, arm, 0)
}

func (v Memory) Reward(experiment string, arm string) {
	v.rewards.Incr(experiment, arm, 0)
}

func (v Memory) Rewards(experiment string, arm string, incr int) {
	//log.Println(experiment, arm, incr)
	v.rewards.Incr(experiment, arm, incr)
}
