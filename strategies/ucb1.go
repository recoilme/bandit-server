package strategies

import (
	"math"

	"github.com/peleteiro/bandit-server/repository"
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
