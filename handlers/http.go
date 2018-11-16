package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/recoilme/bandit-server/repository"
	"github.com/recoilme/bandit-server/strategies"
)

type postStruct struct {
	Group    string   `json:"group"`
	Variants []string `json:"variants"`
	Count    int      `json:"count"`
}

type variants struct {
	Variants []string `json:"variants"`
}

func toJson(r map[string]string) string {
	b, err := json.Marshal(r)
	if err != nil {
		return ""
	}
	return string(b)
}

func doDefaultHeaders(w http.ResponseWriter, r *http.Request) {
	if origin := r.Header.Get("Origin"); origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
	}
	w.Header().Set("Access-Control-Allow-Methods", "PUT, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
}

func doGet(strategy strategies.Strategy, repo repository.Repository, w http.ResponseWriter, r *http.Request) {
	doDefaultHeaders(w, r)

	var result map[string]string = make(map[string]string)

	r.ParseForm()
	for context, values := range r.Form {
		experiments := strings.Split(values[0], ",")
		result[context] = strategy.Choose(repo, context, experiments)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprint(w, toJson(result))
}

func doPut(repo repository.Repository, w http.ResponseWriter, r *http.Request) {
	doDefaultHeaders(w, r)

	r.ParseForm()
	//log.Println("Put")
	for context, values := range r.Form {
		//fmt.Println(values[0], context)
		if strings.Contains(values[0], ",") {
			arr := strings.Split(values[0], ",")
			rew, _ := strconv.Atoi(arr[1])
			fmt.Println(context, arr[0], rew)
			repo.Rewards(context, arr[0], rew)
		} else {
			fmt.Println(context, values[0])
			repo.Reward(context, values[0])
		}

	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(w, "ok")
}

func NewHttpHandler(strategy strategies.Strategy, repo repository.Repository) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			doGet(strategy, repo, w, r)
		case "POST":
			doPost(strategy, repo, w, r)
		case "PUT":
			doPut(repo, w, r)
		case "OPTIONS":
			doDefaultHeaders(w, r)
		default:
			http.Error(w, "Method Not Allowed", 405)
		}
	}
}

func doPost(strategy strategies.Strategy, repo repository.Repository, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	decoder := json.NewDecoder(r.Body)
	var p postStruct
	err := decoder.Decode(&p)

	var result map[string]string = make(map[string]string)
	if err != nil {
		result["error"] = err.Error()
		fmt.Fprint(w, toJson(result))
		return
	}
	log.Println("P:", p)
	res := strategy.ChooseMany(repo, p.Group, p.Variants, p.Count)
	//var v variants
	b, err := json.Marshal(res)
	if err != nil {
		result["error"] = err.Error()
		fmt.Fprint(w, toJson(result))
		return
	}
	fmt.Fprint(w, string(b))
}
