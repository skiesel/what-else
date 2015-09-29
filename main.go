package main

import (
	"encoding/json"
	"math"
	"io/ioutil"
	"fmt"
	"flag"
)

type Quantity struct {
	Pounds, Ounces float64
}

type Recipe struct {
	IngredientList map[string]Quantity
}

type RecipeIndex struct {
	Recipes map[string]Recipe
}

type Purchase struct {
	PurchaseList map[string]Quantity
	RecipeList []string
}


var (
	recipeFile = flag.String("RecipeFile", "recipes.json", "The recipe file index")
	purchaseFile = flag.String("PurchaseFile", "purchase.json", "The purchase file")
)

func main() {
	flag.Parse()

	recipeFileContents, err := ioutil.ReadFile(*recipeFile)
	if err != nil {
		panic(err)
	}

	purchaseFileContents, err := ioutil.ReadFile(*purchaseFile)
	if err != nil {
		panic(err)
	}

	var recipeIndex RecipeIndex
	json.Unmarshal(recipeFileContents, &recipeIndex)
	

	var purchase Purchase
	json.Unmarshal(purchaseFileContents, &purchase)

	leftOvers := map[string]Quantity{}
	for key, val := range purchase.PurchaseList {
		leftOvers[key] = val
	}

	for _, recipeName := range purchase.RecipeList {
		recipe, found := recipeIndex.Recipes[recipeName]
		if !found {
			panic("Could not find recipe " + recipeName)
		}

		leftOvers = minus(leftOvers, recipe.IngredientList)
	}

	// fmt.Println("Left Over")
	// for key, value := range leftOvers {
	// 	fmt.Printf("%s : \t\t%g lbs %g oz\n", key, value.Pounds, value.Ounces)
	// }

	extraRecipes := map[string]map[string]Quantity{}
	var closestRecipe string
	var smallestQuantity Quantity
	smallestQuantity.Pounds = -10000000

	for recipeName, recipe := range recipeIndex.Recipes {
		extraRecipes[recipeName] = minus(leftOvers, recipe.IngredientList)
		value := cumulativeNegativeAmount(extraRecipes[recipeName] )

		if value.more(smallestQuantity) {
			smallestQuantity = value
			closestRecipe = recipeName
		}
	}


	fmt.Print("Already making: ")
	for _, r := range purchase.RecipeList {
		fmt.Print(r + " ")
	}
	fmt.Println()

	fmt.Println("Could make: " + closestRecipe)

	var zero Quantity

	fmt.Println("Need more:")
	for ingredient, quantity := range extraRecipes[closestRecipe] {
		if quantity.less(zero) {
			quantity = quantity.negative()
			fmt.Printf("\t%s %glbs %goz\n", ingredient, quantity.Pounds, quantity.Ounces)
		}
	}
}


func (q1 *Quantity) more(q2 Quantity) bool {
	ounces1 := q1.Pounds * 16.0 + q1.Ounces
	ounces2 := q2.Pounds * 16.0 + q2.Ounces
	return ounces1 > ounces2
}

func (q1 *Quantity) less(q2 Quantity) bool {
	ounces1 := q1.Pounds * 16.0 + q1.Ounces
	ounces2 := q2.Pounds * 16.0 + q2.Ounces
	return ounces1 < ounces2
}

func cumulativeNegativeAmount(m map[string]Quantity) Quantity {
	var total Quantity
	for _, quantity := range m {
		if quantity.Pounds < 0 || quantity.Ounces < 0 {
			total = total.add(quantity)
		}
	}
	return total
}

func copy(m map[string]Quantity) map[string]Quantity {
	copy := map[string]Quantity{}
	for ingredient, quantity := range m {
		copy[ingredient] = quantity
	}
	return copy
}

func minus(left map[string]Quantity, right map[string]Quantity) map[string]Quantity {
	difference := copy(left)

	for ingredient, quantity := range right {
		leftOverQuantity, found := difference[ingredient]
		if !found {
			difference[ingredient] = quantity.negative()
			continue
		}
		difference[ingredient] = leftOverQuantity.minus(quantity)
	}

	return difference
}

func (q1 *Quantity) minus(q2 Quantity) Quantity {
	var q3 Quantity

	ounces1 := q1.Pounds * 16.0 + q1.Ounces
	ounces2 := q2.Pounds * 16.0 + q2.Ounces
	ounces3 := ounces1 - ounces2

	q3.Pounds = math.Floor(ounces3 / 16)
	q3.Ounces = math.Mod(ounces3, 16)

	return q3
}

func (q1 *Quantity) add(q2 Quantity) Quantity {
	var q3 Quantity

	ounces1 := q1.Pounds * 16.0 + q1.Ounces
	ounces2 := q2.Pounds * 16.0 + q2.Ounces
	ounces3 := ounces1 + ounces2

	q3.Pounds = math.Floor(ounces3 / 16)
	q3.Ounces = math.Mod(ounces3, 16)

	return q3
}

func (q1 *Quantity) negative() Quantity {
	var q2 Quantity
	q2.Pounds = -q1.Pounds
	q2.Ounces = -q1.Ounces
	return q2
}