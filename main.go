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

	brewingLookup := map[string]string{}

	for _, recipeName := range purchase.RecipeList {
		brewingLookup[recipeName] = recipeName
	}

	leftOvers := copy(purchase.PurchaseList)

	for _, recipeName := range purchase.RecipeList {
		recipe, found := recipeIndex.Recipes[recipeName]
		if !found {
			panic("Could not find recipe " + recipeName)
		}

		leftOvers = minus(leftOvers, recipe.IngredientList)
	}

	extraRecipes := map[string]map[string]Quantity{}
	recipes := map[string]string{}

	var smallestQuantity, smallestQuantityNotBrewing Quantity
	smallestQuantity.Pounds = -10000000
	smallestQuantityNotBrewing.Pounds = -10000000
	fewestAdditionalIngredients := 10000000
	fewestAdditionalIngredientsNowBrewing := 10000000

	for recipeName, recipe := range recipeIndex.Recipes {
		extraRecipes[recipeName] = minus(leftOvers, recipe.IngredientList)

		which, value := cumulativeNegativeAmount(extraRecipes[recipeName] )

		count := 0
		for _, ingredient := range which {
			_, found := purchase.PurchaseList[ingredient]
			if !found {
				count += 1
			}
		}

		if count < fewestAdditionalIngredients {
			fewestAdditionalIngredients = count
			recipes["fewestAdditionalIngredientsRecipe"] = recipeName
		}

		if value.more(smallestQuantity) {
			smallestQuantity = value
			recipes["smallestQuantiyRecipe"] = recipeName
		}


		_, found := brewingLookup[recipeName]
		if !found {
			if count < fewestAdditionalIngredientsNowBrewing {
				fewestAdditionalIngredientsNowBrewing = count
				recipes["fewestAdditionalIngredientsRecipeNotBrewing"] = recipeName
			}

			if value.more(smallestQuantityNotBrewing) {
				smallestQuantityNotBrewing = value
				recipes["smallestQuantiyRecipeNotBrewing"] = recipeName
			}
		}

	}

	fmt.Print("Already making: ")
	for _, r := range purchase.RecipeList {
		fmt.Print(r + ", ")
	}
	fmt.Println("\n")

	fmt.Println("Could make (smallest additional weight): " + recipes["smallestQuantiyRecipe"])
	printNeeds(extraRecipes[recipes["smallestQuantiyRecipe"]])

	fmt.Printf("\nOr Could make (fewest new ingredients: %d): %s\n", fewestAdditionalIngredients, recipes["fewestAdditionalIngredientsRecipe"])
	printNeeds(extraRecipes[recipes["fewestAdditionalIngredientsRecipe"]])

	fmt.Println("\nOr Could make (additional weight not brewing already): " + recipes["smallestQuantiyRecipeNotBrewing"])
	printNeeds(extraRecipes[recipes["smallestQuantiyRecipeNotBrewing"]])

	fmt.Printf("\nOr Could make (fewest new ingredients not brewing already: %d): %s\n", fewestAdditionalIngredientsNowBrewing, recipes["fewestAdditionalIngredientsRecipeNotBrewing"])
	printNeeds(extraRecipes[recipes["fewestAdditionalIngredientsRecipeNotBrewing"]])
}

func printNeeds(list map[string]Quantity) {
	var zero Quantity
	fmt.Println("Need more:")
	for ingredient, quantity := range list {
		if quantity.less(zero) {
			fmt.Printf("\t%s %glbs %goz\n", ingredient, math.Abs(quantity.Pounds), math.Abs(quantity.Ounces))
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

func cumulativeNegativeAmount(m map[string]Quantity) ([]string, Quantity) {
	var total Quantity
	missingIngredients := []string{}
	for ingredient, quantity := range m {
		if quantity.Pounds < 0 || quantity.Ounces < 0 {
			total = total.add(quantity)
			missingIngredients = append(missingIngredients, ingredient)
		}
	}
	return missingIngredients, total
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
		} else {
			difference[ingredient] = leftOverQuantity.minus(quantity)
		}
	}

	return difference
}

func (q1 *Quantity) minus(q2 Quantity) Quantity {
	var q3 Quantity

	ounces1 := q1.Pounds * 16.0 + q1.Ounces
	ounces2 := q2.Pounds * 16.0 + q2.Ounces
	ounces3 := ounces1 - ounces2

	q3.Pounds = float64(int64(ounces3 / 16))
	q3.Ounces = math.Mod(ounces3, 16)

	return q3
}

func (q1 *Quantity) add(q2 Quantity) Quantity {
	var q3 Quantity

	ounces1 := q1.Pounds * 16.0 + q1.Ounces
	ounces2 := q2.Pounds * 16.0 + q2.Ounces
	ounces3 := ounces1 + ounces2

	q3.Pounds = float64(int64(ounces3 / 16))
	q3.Ounces = math.Mod(ounces3, 16)

	return q3
}

func (q1 *Quantity) negative() Quantity {
	var q2 Quantity
	q2.Pounds = -q1.Pounds
	q2.Ounces = -q1.Ounces
	return q2
}