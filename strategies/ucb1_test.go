package strategies

import (
	"testing"

	"github.com/recoilme/bandit-server/repository"
	"github.com/stretchr/testify/assert"
)

var strgy = NewUCB1()

func TestReturnTheOneWithZeroHits(t *testing.T) {
	repo := repository.NewMemory()
	repo.Hit("exp", "arm1")

	choosenOne := strgy.Choose(repo, "exp", []string{"arm1", "armWithZeroHits"})

	assert.Equal(t, choosenOne, "armWithZeroHits")
}
