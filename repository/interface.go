package repository

type ExperimentData struct {
	TotalHits int64
	Arms      map[string]ArmData
}

type ArmData struct {
	Hits    int64
	Rewards int64
}

type Repository interface {
	Get(experiment string, arms []string) ExperimentData
	//	Gets(experiment string, arms []string, count int) []string
	Hit(experiment string, arm string)
	Reward(experiment string, arm string)
	Rewards(experiment string, arm string, reward int)
}
