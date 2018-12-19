package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/recoilme/pudge"
)

var (
	port  = 3000
	debug = false
	cfg   *pudge.Config
)

type Hit struct {
	Arm string `json:"arm"`
	Cnt int    `json:"cnt"`
}

type Arms struct {
	Arm string `json:"arm"`
}

type Stat struct {
	Arm   string  `json:"arm"`
	Hit   int     `json:"hit"`
	Rew   int     `json:"rew"`
	Score float64 `json:"score"`
}

func init() {
	flag.IntVar(&port, "port", 3000, "http port")
	flag.BoolVar(&debug, "debug", false, "--debug=true")
	cfg = pudge.DefaultConfig()
	cfg.StoreMode = 2
}

func main() {

	flag.Parse()

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: InitRouter(),
	}

	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	// Close db
	if err := pudge.CloseAll(); err != nil {
		log.Fatal("Database Shutdown:", err)
	}
	log.Println("Server exiting")

}

func globalRecover(c *gin.Context) {
	defer func(c *gin.Context) {

		if err := recover(); err != nil {
			if err := pudge.CloseAll(); err != nil {
				log.Println("Database Shutdown err:", err)
			}
			log.Println("Server recovery with err:", err)
			gin.RecoveryWithWriter(gin.DefaultErrorWriter)
			//c.AbortWithStatus(500)
		}
	}(c)
	c.Next()
}

// InitRouter - init router
func InitRouter() *gin.Engine {
	if debug {
		gin.SetMode(gin.DebugMode)

	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	if debug {
		r.Use(gin.Logger())
	}
	r.Use(globalRecover)

	r.GET("/", ok)
	r.GET("/stats/:group/:count", stats)
	r.POST("/stats/:group/:count", stats)
	r.POST("/write/:param/:group", write)

	return r
}

func ok(c *gin.Context) {
	c.String(http.StatusOK, "%s", "ok")
}

func write(c *gin.Context) {
	var err error
	group := c.Param("group")
	param := c.Param("param")
	dbPrefix := ""
	switch param {
	case "hits":
		dbPrefix = param
		break
	case "rewards":
		dbPrefix = param
		break
	default:
		renderError(c, errors.New("Unknown param:"+param))
		return
	}
	var hits = make([]Hit, 0, 0)
	err = c.ShouldBind(&hits)
	if err != nil {
		renderError(c, err)
		return
	}
	db, err := pudge.Open(dbPrefix+"/"+group, cfg)
	if err != nil {
		renderError(c, err)
		return
	}
	for _, h := range hits {
		_, err = db.Counter(h.Arm, h.Cnt)
		if err != nil {
			break
		}
	}
	if err != nil {
		renderError(c, err)
		return
	}
	c.String(http.StatusOK, "%s", "ok")
}

func renderError(c *gin.Context, err error) {
	if err != nil {
		log.Println(err)
		c.Error(err)
		c.JSON(http.StatusUnprocessableEntity, c.Errors)
		return
	}
}

func stats(c *gin.Context) {
	t1 := time.Now()
	var err error
	group := c.Param("group")
	var arms = make([]Arms, 0, 0)
	switch c.Request.Method {
	case "POST":
		err = c.ShouldBind(&arms)
		if err != nil {
			renderError(c, err)
			return
		}
	}

	count, _ := strconv.Atoi(c.Param("count"))
	dbhits, err := pudge.Open("hits/"+group, cfg)
	if err != nil {
		renderError(c, err)
		return
	}
	dbrew, err := pudge.Open("rewards/"+group, cfg)
	if err != nil {
		renderError(c, err)
		return
	}
	var data = make([][]byte, 0, 0)
	if len(arms) > 0 {
		for _, arm := range arms {
			data = append(data, []byte(arm.Arm))
		}
	} else {

		data, err = dbhits.Keys(nil, 0, 0, true)
	}

	if err != nil {
		renderError(c, err)
		return
	}
	t2 := time.Now()
	if debug {
		//log.Println("len", len(data))
		//fmt.Printf("The t2 took %v to run.\n", t2.Sub(t1))
	}
	var stats = make([]Stat, 0, 0)
	//var totalrew = 0
	var totalHits int
	for _, key := range data {
		var hit, rew int
		errGet := dbhits.Get(key, &hit)
		if errGet != nil && errGet != pudge.ErrKeyNotFound {
			err = errGet
			break
		}
		dbrew.Get(key, &rew)
		//totalrew += rew
		var stat Stat
		stat.Arm = string(key)
		stat.Hit = hit
		stat.Rew = rew
		totalHits += hit
		stats = append(stats, stat)
	}
	t3 := time.Now()
	if debug {
		_ = t1
		_ = t2
		_ = t3
		//fmt.Printf("The t3 took %v to run.\n", t3.Sub(t2))
	}
	var scores = make([]Stat, 0, 0)
	for _, s := range stats {
		s.Score = s.calcScore(totalHits)
		scores = append(scores, s)
	}
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].Score > scores[j].Score
	})
	if len(scores) > count {
		scores = scores[:count]
	}
	if err != nil {
		renderError(c, err)
		return
	}
	//log.Println("total rew:", totalrew)
	//log.Println("Score:", scores, stats[0].calcScore(12))

	c.JSON(http.StatusOK, scores)
}

func (st *Stat) calcScore(totalHits int) float64 {
	if st.Hit == 0 {
		return float64(math.Sqrt((2 * math.Log(float64(totalHits+1)))))
	}
	return float64(st.Rew)/float64(st.Hit) + math.Sqrt((2*math.Log(float64(totalHits)))/float64(st.Hit))
}
