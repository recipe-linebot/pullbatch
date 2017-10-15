/**
 * Copyright 2017 recipe-linebot
 */
package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
)

type RecipeLinebotConfig struct {
	PullBatch struct {
		ProgressFilePath string `json:"progress_filepath"`
	} `json:"pull_batch"`
	RakutenAPI struct {
		AppId        string `json:"app_id"`
		CallInterval int    `json:"call_interval_sec"`
	} `json:"rakuten_api"`
	RecipeDB struct {
		Host           string `json:"host"`
		Index          string `json:"index"`
		RecipeDoctype  string `json:"recipe_doctype"`
		RankingDoctype string `json:"ranking_doctype"`
	} `json:"recipe_db"`
}

func main() {
	confpath := flag.String("c", "", "config file path")
	flag.Parse()
	confdata, err := ioutil.ReadFile(*confpath)
	if err != nil {
		log.Fatal(err)
	}
	var conf RecipeLinebotConfig
	err = json.Unmarshal(confdata, &conf)
	if err != nil {
		log.Fatal(err)
	}
	pullRecipes(&conf)
}
