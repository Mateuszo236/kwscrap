package main

import (
	"fmt"
	"log"
	"strconv"
)

// Weights for Polish property registry control digit calculation (modulo-10 algorithm)
// Each position in the registry number is multiplied by its corresponding weight
var wagi = []int{1, 3, 7, 1, 3, 7, 1, 3, 7, 1, 3, 7}

// Character to numeric value mapping per Polish registry standard
// Used for control digit validation (A=11, B=12, etc.)
var wartosciZnakow = map[rune]int{
	'0': 0, '1': 1, '2': 2, '3': 3, '4': 4, '5': 5, '6': 6, '7': 7, '8': 8, '9': 9,
	'X': 10, 'A': 11, 'B': 12, 'C': 13, 'D': 14, 'E': 15, 'F': 16, 'G': 17, 'H': 18,
	'I': 19, 'J': 20, 'K': 21, 'L': 22, 'M': 23, 'N': 24, 'O': 25, 'P': 26, 'R': 27,
	'S': 28, 'T': 29, 'U': 30, 'W': 31, 'Y': 32, 'Z': 33,
}

// ObliczCyfreKontrolna calculates the control digit for a registry number
// using the Polish modulo-10 algorithm (court code + number weighted sum % 10)
func ObliczCyfreKontrolna(kodSadu string, numer string) (string, error) {
	calosc := kodSadu + numer
	suma := 0

	for i, znak := range calosc {
		val, ok := wartosciZnakow[znak]
		if !ok {
			return "", fmt.Errorf("invalid character in registry number: %c", znak)
		}
		suma += val * wagi[i]
	}

	cyfra := suma % 10
	return strconv.Itoa(cyfra), nil
}

// GeneratorKW yields registry records as they are generated via a channel
// This prevents memory overhead by not storing all records in memory at once
func GeneratorKW(kodSadu string, startNum, endNum int) <-chan Ksiega {
	out := make(chan Ksiega)

	go func() {
		defer close(out)

		for i := startNum; i <= endNum; i++ {
			numerStr := fmt.Sprintf("%08d", i)

			if len(numerStr) != 8 {
				continue
			}

			cyfraKontrolna, err := ObliczCyfreKontrolna(kodSadu, numerStr)
			if err != nil {
				log.Printf("Failed to calculate control digit for %s/%s: %v", kodSadu, numerStr, err)
				continue
			}

			out <- Ksiega{
				KodSadu:        kodSadu,
				Numer:          numerStr,
				CyfraKontrolna: cyfraKontrolna,
			}
		}
	}()

	return out
}
