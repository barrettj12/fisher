package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/barrettj12/collections"
	"gonum.org/v1/gonum/stat/distuv"
)

const (
	INIT_POP   = 500
	INIT_RATIO = 0.8
	NUM_GENS   = 50

	// Beta distribution parameters
	// v controls the variation in sex ratios. Smaller v = more variation
	BETA_V = 20.0
)

func main() {
	rand.Seed(time.Now().Unix())

	// Initialise population
	numMales := int(INIT_RATIO * float64(INIT_POP))
	males := collections.NewList[Human](numMales)
	for i := 0; i < numMales; i++ {
		males.Append(Human{Male, randSexRatio(INIT_RATIO), nil, nil})
	}

	numFemales := INIT_POP - numMales
	females := collections.NewList[Human](numFemales)
	for i := 0; i < numFemales; i++ {
		females.Append(Human{Female, randSexRatio(INIT_RATIO), nil, nil})
	}

	printGen(males, females)

	for gen := 0; gen < NUM_GENS; gen++ {
		newMales := collections.NewList[Human](0)
		newFemales := collections.NewList[Human](0)

		for _, m := range *males {
			for _, f := range *females {
				// Aim for same population size
				// Chance of reproduction is popSize/(numMales*numFemales)
				chanceRepro := float64(males.Size()+females.Size()) / float64(males.Size()*females.Size())
				if rand.Float64() < chanceRepro {
					child := reproduce(m, f)
					if child.sex == Male {
						newMales.Append(child)
					} else {
						newFemales.Append(child)
					}
				}
			}
		}

		males = newMales
		females = newFemales
		printGen(males, females)
	}

	// Print ratios to file
	// Tends towards a beta distribution with mean=0.5, v=BETA_V
	f, err := os.Create("ratios.csv")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	for _, h := range *males {
		fmt.Fprintln(f, h.spermRatio)
	}
	for _, h := range *females {
		fmt.Fprintln(f, h.spermRatio)
	}
}

func printGen(males, females *collections.List[Human]) {
	ratio := float64(males.Size()) / float64(males.Size()+females.Size())
	fmt.Printf("%d males	%d females	sex ratio %g\n", males.Size(), females.Size(), ratio)
}

type Human struct {
	sex        Sex
	spermRatio float64 // what % of sperm are male

	mother, father *Human
}

type Sex string

const (
	Male   Sex = "male"
	Female Sex = "female"
)

func reproduce(m, f Human) Human {
	sex := Female
	if rand.Float64() < m.spermRatio {
		sex = Male
	}

	defer func() {
		if r := recover(); r != nil {
			printFamilyTree(m, "")
			printFamilyTree(f, "")
			panic(r)
		}
	}()
	// is this a good model for propagation of sex ratios?
	ratio := randSexRatio((m.spermRatio + f.spermRatio) / 2)
	// ratio := randSexRatio(f.spermRatio) // keeps panicking when someone's ratio gets to 0 or 1
	return Human{sex, ratio, &f, &m}
}

// Random sex ratios based on beta distribution
// m is the mean of the distribution
func randSexRatio(m float64) float64 {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("m = ", m)
			panic(r)
		}
	}()

	v := BETA_V
	return distuv.Beta{
		Alpha: m * v,
		Beta:  (1 - m) * v,
	}.Rand()
}

func printFamilyTree(h Human, indent string) {
	fmt.Printf(indent+"%v\n", h)
	if h.mother != nil {
		printFamilyTree(*h.mother, indent+"  ")
	}
	if h.father != nil {
		printFamilyTree(*h.father, indent+"  ")
	}
}
