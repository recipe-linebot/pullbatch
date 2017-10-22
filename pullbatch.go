/**
 * Copyright 2016 tech0522
 */
package main

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"time"
)

type BatchProgress struct {
	CategoriesByType  map[RecipeCategoryType]RecipeCategoryList
	CategoryIdxByType map[RecipeCategoryType]int
}

func restoreProgress(restorePath string, progress *BatchProgress) error {
	restoreFile, err := os.Open(restorePath)
	if err != nil {
		return err
	}
	defer restoreFile.Close()
	decoder := gob.NewDecoder(bufio.NewReader(restoreFile))
	return decoder.Decode(&progress)
}

func storeProgress(progress *BatchProgress, storePath string) error {
	storeFile, err := os.OpenFile(storePath, os.O_WRONLY+os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer storeFile.Close()
	encoder := gob.NewEncoder(bufio.NewWriter(storeFile))
	return encoder.Encode(&progress)
}

type RecipeDocument struct {
	Materials   []string `json:"materials"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	ImageURL    string   `json:"image_url"`
	DetailURL   string   `json:"detail_url"`
}

type RankingDocument struct {
	Concept string `json:"concept"`
	Recipes []int  `json:"recipes"`
}

func pullRecipesOnCategory(categoryID string, categoryName string, config *RecipeLinebotConfig) error {
	time.Sleep(time.Duration(config.RakutenAPI.CallInterval) * time.Second)
	ranking, err := FetchRecipeRanking(categoryID, config.RakutenAPI.AppID)
	if err != nil {
		return err
	}
	if len(ranking.Recipes) == 0 {
		log.Printf("recipe not found: category=%v(%v)", categoryID, categoryName)
	} else {
		var recipes []int
		for _, recipe := range ranking.Recipes {
			log.Printf("post recipe: id=%v, title=%v", recipe.ID, recipe.Title)
			apiURL := url.URL{Scheme: "http", Host: config.RecipeDB.Host,
				Path: path.Join(config.RecipeDB.Index, config.RecipeDB.RecipeDoctype, strconv.Itoa(recipe.ID))}
			imageURL, err := url.Parse(recipe.LargeImageURL)
			if err != nil {
				log.Fatal(err)
			}
			imageURL.Scheme = "https"
			document := RecipeDocument{Materials: recipe.Materials, Title: recipe.Title, Description: recipe.Description,
				ImageURL: imageURL.String(), DetailURL: recipe.URL}
			reqBody, nil := json.Marshal(document)
			if err != nil {
				log.Fatal(err)
			}
			resp, err := http.Post(apiURL.String(), "application/json", bytes.NewBuffer(reqBody))
			if err != nil {
				log.Fatal(err)
			}
			defer resp.Body.Close()
			if resp.StatusCode/100 != 2 {
				bodyAsString := "(failed to read)"
				body, err := ioutil.ReadAll(resp.Body)
				if err == nil {
					bodyAsString = string(body)
				}
				log.Fatalf("Bad status code: url=%v, code=%v, body=%v", apiURL.String(), resp.Status, bodyAsString)
			}
			recipes = append(recipes, recipe.ID)
		}
		log.Printf("post ranking: category=%v(%v), recipes=%v", categoryID, categoryName, recipes)
		apiURL := url.URL{Scheme: "http", Host: config.RecipeDB.Host,
			Path: path.Join(config.RecipeDB.Index, config.RecipeDB.RankingDoctype, categoryID)}
		document := RankingDocument{Concept: categoryName, Recipes: recipes}
		reqBody, nil := json.Marshal(document)
		if err != nil {
			log.Fatal(err)
		}
		resp, err := http.Post(apiURL.String(), "application/json", bytes.NewBuffer(reqBody))
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode/100 != 2 {
			bodyAsString := "(failed to read)"
			body, err := ioutil.ReadAll(resp.Body)
			if err == nil {
				bodyAsString = string(body)
			}
			log.Fatalf("Bad status code: url=%v, code=%v, body=%v", apiURL.String(), resp.Status, bodyAsString)
		}
	}
	return nil
}

func pullRecipesOnCategoryLevel(categoryType RecipeCategoryType,
	progress BatchProgress, config *RecipeLinebotConfig) error {
	for idx, category := range progress.CategoriesByType[categoryType] {
		categoryID := category.ID
		if categoryType != RecipeCategoryLarge {
			if idx <= progress.CategoryIdxByType[categoryType] {
				continue
			}
			categoryURL, err := url.Parse(category.URL)
			if err != nil {
				return err
			}
			categoryID = path.Base(categoryURL.Path)
		}
		err := pullRecipesOnCategory(categoryID, category.Name, config)
		if err != nil {
			return err
		}
		progress.CategoryIdxByType[categoryType] = idx
		err = storeProgress(&progress, config.PullBatch.ProgressFilePath)
		if err != nil {
			return err
		}
	}
	return nil
}

func pullRecipes(config *RecipeLinebotConfig) {
	log.Print("start pull batch")

	// Restore the progress up to the previous working
	var progress BatchProgress
	restored := true
	err := restoreProgress(config.PullBatch.ProgressFilePath, &progress)
	if err != nil {
		if os.IsNotExist(err) {
			restored = false
		} else {
			log.Fatal(err)
		}
	}

	if !restored {
		result, err := FetchRecipeCategories(RecipeCategoryAll, config.RakutenAPI.AppID)
		if err != nil {
			log.Fatal(err)
		}
		progress.CategoriesByType[RecipeCategoryLarge] = result.Categories.ByLarge
		progress.CategoriesByType[RecipeCategoryMedium] = result.Categories.ByMedium
		progress.CategoriesByType[RecipeCategorySmall] = result.Categories.BySmall
	}
	err = pullRecipesOnCategoryLevel(RecipeCategoryLarge, progress, config)
	if err != nil {
		log.Fatal(err)
	}
	err = pullRecipesOnCategoryLevel(RecipeCategoryMedium, progress, config)
	if err != nil {
		log.Fatal(err)
	}
	err = pullRecipesOnCategoryLevel(RecipeCategorySmall, progress, config)
	if err != nil {
		log.Fatal(err)
	}
	err = os.Remove(config.PullBatch.ProgressFilePath)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("pull batch finished")
}
