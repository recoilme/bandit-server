package strategies

import (
	"math"
	"sort"

	"github.com/recoilme/bandit-server/repository"
)

type UCB1 struct{}

func NewUCB1() UCB1 {
	return UCB1{}
}

func (_ UCB1) Choose(repo repository.Repository, experiment string, arms []string) (arm string) {
	arm = getHighestScoreArm(repo, experiment, arms)
	repo.Hit(experiment, arm)
	return arm
}

func getHighestScoreArm(repo repository.Repository, experiment string, arms []string) string {
	var highestArm string
	var highestScore float64

	var expData = repo.Get(experiment, arms)
	for arm, armData := range expData.Arms {
		if armData.Hits == 0 {
			return arm
		}

		var score = calcScore(expData.TotalHits, armData.Hits, armData.Rewards)
		//log.Println(arm, expData.TotalHits, armData.Hits, armData.Rewards, score)
		if score > highestScore {
			highestArm = arm
			highestScore = score
		}
	}

	return highestArm
}

func calcScore(totalHits int64, hits int64, rewards int64) float64 {
	return float64(float64(rewards)/float64(hits)) + math.Sqrt((2*math.Log(float64(totalHits)))/float64(hits))
}

func (_ UCB1) ChooseMany(repo repository.Repository, experiment string, arms []string, cnt int) (result []string) {
	type kv struct {
		Key   string
		Value float64
	}
	var ss []kv
	var expData = repo.Get(experiment, arms)
	for arm, armData := range expData.Arms {
		if armData.Hits == 0 {
			ss = append(ss, kv{arm, float64(100)})
			continue
		}
		var score = calcScore(expData.TotalHits, armData.Hits, armData.Rewards)
		ss = append(ss, kv{arm, score})
	}
	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Value > ss[j].Value
	})
	//log.Println("ss", ss)
	for i, kv := range ss {
		if i > 0 {
			if i >= cnt {
				break
			}
		}
		repo.Hit(experiment, kv.Key)
		result = append(result, kv.Key)
	}
	return result
}
