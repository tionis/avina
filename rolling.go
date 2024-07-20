package main

import (
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
)

var numberRegex = regexp.MustCompile("[0-9]+")

func roll_x_y_sided_dice(x, y int) []int {
	results := make([]int, x)
	for i := 0; i < x; i++ {
		results[i] = rand.Intn(y) + 1
	}
	return results
}

func codRoll(amount int, modifier string) (string, error) {
	var err error
	rote := strings.Contains(modifier, "r")
	noreroll := strings.Contains(modifier, "n")
	numbers := numberRegex.FindAllString(modifier, -1)
	rerollTreshold := 10
	if len(numbers) > 0 {
		rerollTreshold, err = strconv.Atoi(numbers[0])
		if err != nil {
			return "", err
		}
	}
	results := make([][]int, amount)
	for i := 0; i < amount; i++ {
		results[i] = roll_x_y_sided_dice(1, 10)
		if !noreroll {
			if rote {
				results[i] = append(results[i], roll_x_y_sided_dice(1, 10)[0])
			}
			lastResult := results[i][len(results[i])-1]
			for lastResult >= rerollTreshold {
				results[i] = append(results[i], roll_x_y_sided_dice(1, 10)[0])
				lastResult = results[i][len(results[i])-1]
			}
		}
	}
	var successes, ones int
	var diceResultStr string
	resultCount := 0
	for i := 0; i < amount; i++ {
		diceResultStr += "["
		for j := 0; j < len(results[i]); j++ {
			resultCount++
			diceResultStr += strconv.Itoa(results[i][j])
			if results[i][j] == 1 {
				ones++
			} else if results[i][j] >= 8 {
				successes++
			}
			if j != len(results[i])-1 {
				diceResultStr += " "
			}
		}
		diceResultStr += "]"
		if i != amount-1 {
			diceResultStr += " "
		}
	}
	var suffix string
	isCritFail := ones >= resultCount/2
	if isCritFail {
		suffix = "You have a [Critical Failure]!"
		if successes > 0 {
			suffix += "\nBut you also have " + strconv.Itoa(successes) + " successes"
		}
	} else {
		suffix = "You have " + strconv.Itoa(successes) + " successes"
		if successes >= 5 {
			suffix += "\nThat's an [Exceptional Success]!"
		}
	}
	response := "cod(" + strconv.Itoa(amount)
	if len(modifier) > 0 {
		response += "," + modifier
	}
	response += ")= [" + strconv.Itoa(successes) + "/" + strconv.Itoa(ones) + "]\n" + diceResultStr + "\n" + suffix
	return response, nil
}

func rollChance() string {
	result := roll_x_y_sided_dice(1, 10)[0]
	switch result {
	case 1:
		return "Chance Die:\n[1]\nCritical Failure!"
	case 10:
		return "Chance Die:\n[10]\nSuccess!"
	default:
		return "Chance Die:\n[" + strconv.Itoa(result) + "]\nFailure!"
	}
}

func roll(amountStr string, modifier string) (string, error) {
	amount, err := strconv.Atoi(amountStr)
	switch amountStr {
	case "chance":
		return rollChance(), nil
	default:
		if errors.Is(err, strconv.ErrSyntax) {
			parts := strings.Split(amountStr, "d")
			if len(parts) != 2 {
				return "Invalid roll format", errors.New("Invalid roll format")
			}
			amount, err = strconv.Atoi(parts[0])
			if err != nil {
				return "first argument before d not a number", err
			}
			sides, err := strconv.Atoi(parts[1])
			if err != nil {
				return "second argument after d not a number", err
			}
			dice_result := roll_x_y_sided_dice(amount, sides)
			var sum int
			var diceResultStr string
			for i := 0; i < amount; i++ {
				sum += dice_result[i]
				if i == amount-1 {
					diceResultStr += strconv.Itoa(dice_result[i]) + " = "
				} else {
					diceResultStr += strconv.Itoa(dice_result[i]) + " + "
				}
			}
			return diceResultStr + strconv.Itoa(sum), nil
		} else if err == nil {
			return codRoll(amount, modifier)
		} else {
			return "could not parse roll as either xdy or cod roll (first arg not a number)", err
		}
	}
}

// shadowroll rolls a number of dice using Shadowrun rules
// if exploding is true, 6s explode (are rerolled)
func shadowroll(amount int, exploding bool) string {
	results := make([][]int, amount)
	for i := 0; i < amount; i++ {
		if exploding {
			lastResult := 6
			results[i] = make([]int, 0)
			for lastResult == 6 {
				lastResult = roll_x_y_sided_dice(1, 6)[0]
				results[i] = append(results[i], lastResult)
			}
		} else {
			results[i] = roll_x_y_sided_dice(1, 6)
		}
	}
	var hits int
	var ones int
	var diceResultStr string
	var resultCount int
	for i := 0; i < amount; i++ {
		diceResultStr += "["
		for j := 0; j < len(results[i]); j++ {
			resultCount++
			diceResultStr += strconv.Itoa(results[i][j])
			if results[i][j] == 1 {
				ones++
			} else if results[i][j] >= 5 {
				hits++
			}
			if j != len(results[i])-1 {
				diceResultStr += " "
			}
		}
		diceResultStr += "]"
		if i != amount-1 {
			diceResultStr += " "
		}
	}
	var explodingStr string
	if exploding {
		explodingStr = "!!"
	}
	diceInfo := fmt.Sprintf("(((%dd6%s)>=5)cf=1): [%d/%d]", amount, explodingStr, hits, ones)
	var suffix string
	isGlitch := ones >= resultCount/2
	if isGlitch && hits == 0 {
		suffix = "You have a [Critical glitch]!\nGood luck..."
	} else if isGlitch {
		suffix = "That's a [Glitch]!\nLet's see where this goes..."
	} else {
		suffix = "You have " + strconv.Itoa(hits) + " hits"
	}
	return diceInfo + "\n" + diceResultStr + "\n" + suffix
}
