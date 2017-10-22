/**
 * Copyright 2016 tech0522
 */
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
)

const APIEndpoint = "https://app.rakuten.co.jp/services/api/"
const RecipeCategoryListAPIPath = "Recipe/CategoryList/"
const RecipeCategoryListAPILatestVersion = "20121121"
const RecipeCategoryRankingAPIPath = "Recipe/CategoryRanking/"
const RecipeCategoryRankingAPILatestVersion = "20121121"

type RecipeCategory struct {
	ID       string `json:"categoryId"`
	Name     string `json:"categoryName"`
	URL      string `json:"categoryUrl"`
	ParentID string `json:"parentCategoryId"`
}

type RecipeCategoryList []RecipeCategory

type RecipeCategoryListsByType struct {
	ByLarge  RecipeCategoryList `json:"large"`
	ByMedium RecipeCategoryList `json:"medium"`
	BySmall  RecipeCategoryList `json:"small"`
}

type RecipeCategoryListAPIResult struct {
	Categories RecipeCategoryListsByType `json:"result"`
}

type RecipeCategoryType string

const (
	RecipeCategoryLarge  RecipeCategoryType = "large"
	RecipeCategoryMedium RecipeCategoryType = "medium"
	RecipeCategorySmall  RecipeCategoryType = "small"
	RecipeCategoryAll    RecipeCategoryType = ""
)

func FetchRecipeCategories(categoryType RecipeCategoryType, appID string) (*RecipeCategoryListAPIResult, error) {
	apiURL, err := url.Parse(APIEndpoint)
	if err != nil {
		return nil, err
	}
	apiURL.Path = path.Join(apiURL.Path, RecipeCategoryListAPIPath, RecipeCategoryListAPILatestVersion)
	apiURL.RawQuery = "applicationId=" + appID
	if categoryType != RecipeCategoryAll {
		apiURL.RawQuery += "&categoryType=" + string(categoryType)
	}
	resp, err := http.Get(apiURL.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Bad status code: url=%v, code=%v, body=%v", apiURL.String(), resp.Status, string(body))
	}
	var result RecipeCategoryListAPIResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

type RecipeSummary struct {
	ID             int      `json:"recipeId"`
	Title          string   `json:"recipeTitle"`
	URL            string   `json:"recipeUrl"`
	LargeImageURL  string   `json:"foodImageUrl"`
	MediumImageURL string   `json:"mediumImageUrl"`
	SmallImageURL  string   `json:"smallImageUrl"`
	PickUp         int      `json:"pickup"`
	Shop           int      `json:"shop"`
	Nickname       string   `json:"nickname"`
	Description    string   `json:"recipeDescription"`
	Materials      []string `json:"recipeMaterial"`
	Indication     string   `json:"recipeIndication"`
	Cost           string   `json:"recipeCost"`
	PublishDay     string   `json:"recipePublishday"`
	Rank           string   `json:"rank"`
}

type RecipeRanking struct {
	Recipes []RecipeSummary `json:"result"`
}

func FetchRecipeRanking(categoryID string, appID string) (*RecipeRanking, error) {
	apiURL, err := url.Parse(APIEndpoint)
	if err != nil {
		return nil, err
	}
	apiURL.Path = path.Join(apiURL.Path, RecipeCategoryRankingAPIPath, RecipeCategoryRankingAPILatestVersion)
	apiURL.RawQuery = "applicationId=" + appID
	if categoryID != "" {
		apiURL.RawQuery += "&categoryId=" + categoryID
	}
	resp, err := http.Get(apiURL.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Bad status code: url=%v, code=%v, body=%v", apiURL.String(), resp.Status, string(body))
	}
	var ranking RecipeRanking
	err = json.Unmarshal(body, &ranking)
	if err != nil {
		return nil, err
	}
	return &ranking, nil
}
