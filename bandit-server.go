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
	"runtime"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/recoilme/pudge"
)

var (
	port  = 3000
	debug = false
	koef  = 1.0
	//cfg   *pudge.Config
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
	flag.Float64Var(&koef, "koef", 1.0, "--koef=1.0")
	pudge.Open("hits/relap", nil)
	pudge.Open("rewards/relap", nil)
	// Workaround for issue #17393.
	//signal.Notify(make(chan os.Signal), syscall.SIGPIPE)
}

func main() {

	flag.Parse()
	log.Println("port:", port, "debug:", debug, "koef:", koef, "maxproc:", runtime.GOMAXPROCS(0))
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: InitRouter(),
	}

	// Wait for interrupt signal to gracefully shutdown the server with
	// setup signal catching
	quit := make(chan os.Signal, 1)
	// catch all signals since not explicitly listing
	signal.Notify(quit)

	//signal.Notify(sigs,syscall.SIGQUIT)
	// method invoked upon seeing signal
	go func() {
		q := <-quit
		log.Printf("RECEIVED SIGNAL: %s", q)
		//if q == syscall.SIGPIPE || q.String() == "broken pipe" {
		//return
		//}
		//log.Println("Shutdown Server ...")
		ctx, cancel := context.WithTimeout(context.Background(), 2000*time.Millisecond)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Println("Server Shutdown:", err)
		}
		// Close db
		log.Println("Close db ...")
		if err := pudge.CloseAll(); err != nil {
			log.Println("Database Shutdown err:", err)
		}
		log.Println("Close db")
		log.Println("Server exiting")
		time.Sleep(2 * time.Second)
		os.Exit(1)
	}()

	//go func() {
	// service connections
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen: %s\n", err)
	}
	//}()

}

func globalRecover(c *gin.Context) {
	defer func(c *gin.Context) {

		if err := recover(); err != nil {
			if err := pudge.CloseAll(); err != nil {
				log.Println("Database Shutdown err:", err)
			}
			log.Printf("Server recovery with err:%+v\n", err)
			gin.RecoveryWithWriter(gin.DefaultErrorWriter)
			c.AbortWithStatus(500)
			return
			//time.Sleep(200 * time.Millisecond)
			//time.Sleep(2 * time.Second)
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
	r.GET("/backup/:dir", backup)
	r.GET("/stats/:group/:count", stats)
	r.POST("/stats/:group/:count", stats)
	r.POST("/write/:param/:group", write)

	return r
}

func ok(c *gin.Context) {
	c.String(http.StatusOK, "%s", "ok")
}

func backup(c *gin.Context) {
	clean(c)
	dir := c.Param("dir")
	log.Println("backup")
	pudge.BackupAll(dir)
	log.Println("end")
	c.String(http.StatusOK, "%s", "ok")
}

func clean(c *gin.Context) {
	hits, err := pudge.Keys("hits/relap", nil, 0, 0, true)
	log.Println(len(hits))
	if err != nil {
		panic(err)
	}
	rew, err := pudge.Open("rewards/relap", nil)
	rew2, err := pudge.Open("rewards/relap.new2", nil)
	log.Println(rew.Count())
	for _, k := range hits {
		var b []byte
		has, e := rew.Has(k)
		if has && e == nil {
			err := rew.Get(k, &b)
			if err != nil {
				panic(err)
			}
			e := rew2.Set(k, b)
			if e != nil {
				panic(e)
			}
		}
	}
	log.Println(rew2.Count())
	rew2.Close()
	pudge.Open("rewards/relap.new2", nil)

	/*
		if dir == "" {
			dir = "backup"
		}
		dbs.Lock()
		stores := dbs.dbs
		dbs.Unlock()
		//tmp := make(map[string]string)
		for _, db := range stores {
			backup := dir + "/" + db.name
			DeleteFile(backup)
			keys, err := db.Keys(nil, 0, 0, true)
			if err == nil {
				for _, k := range keys {
					var b []byte
					db.Get(k, &b)
					Set(backup, k, b)
				}
			}
			Close(backup)
		}
	*/
}

func write(c *gin.Context) {

	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered in write", r)
		}
	}()

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
	/*
		db, err := pudge.Open(dbPrefix+"/"+group, cfg)
		if err != nil {
			renderError(c, err)
			return
		}*/
	for _, h := range hits {
		if dbPrefix == "rewards" {
			has, err := pudge.Has("hits/"+group, h.Arm)
			if !has || err != nil {
				continue
			}
		}
		_, err = pudge.Counter(dbPrefix+"/"+group, h.Arm, h.Cnt)
		if err != nil {
			log.Println("invalid_argument:", "'"+dbPrefix+"'", "'"+group+"'", "'"+h.Arm+"'", h.Cnt)
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
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered in stats", r)
		}
	}()
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

	/*
		dbhits, err := pudge.Open("hits/"+group, cfg)
		if err != nil {
			renderError(c, err)
			return
		}
		dbrew, err := pudge.Open("rewards/"+group, cfg)
		if err != nil {
			renderError(c, err)
			return
		}*/
	var data = make([][]byte, 0, 0)
	if len(arms) > 0 {
		for _, arm := range arms {
			data = append(data, []byte(arm.Arm))
		}
	} else {
		//log.Println("Error: stats without arms:", arms)
		err = errors.New("Error: stats without arms")
		renderError(c, err)
		return
		//data, err = pudge.Keys("hits/"+group, nil, 0, 0, true) //dbhits.Keys(nil, 0, 0, true)
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
		t22 := time.Now()
		errGet := pudge.Get("hits/"+group, key, &hit) ///dbhits.Get(key, &hit)
		if errGet != nil && errGet != pudge.ErrKeyNotFound {
			err = errGet
			break
		}
		pudge.Get("rewards/"+group, key, &rew)
		t23 := time.Now()
		if t23.Sub(t22) > (20 * time.Millisecond) {
			//fmt.Printf("The t23 took %v to run. With  %s key \n", t23.Sub(t22), string(key))
		}
		//dbrew.Get(key, &rew)

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
	if t3.Sub(t2) > (200 * time.Millisecond) {
		//fmt.Printf("The t3 took %v to run. With %d %v  %+v req \n", t3.Sub(t2), count, t3.Sub(t1), c.Request.Method)
	}

	var scores = make([]Stat, 0, 0)
	for _, s := range stats {
		s.Score = s.calcScore(totalHits)
		scores = append(scores, s)
	}
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].Score > scores[j].Score
	})
	//t4 := time.Now()
	//if t4.Sub(t2) > (10 * time.Millisecond) {
	//	fmt.Printf("The t4 took %v to run.\n", t4.Sub(t2))
	//}

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
	return (koef*float64(st.Rew))/float64(st.Hit) + math.Sqrt((2*math.Log(float64(totalHits)))/float64(st.Hit))
}
