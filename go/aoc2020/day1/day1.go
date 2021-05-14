package main

import (
	"bufio"
	"log"
	"os"
	"strconv"
)

func main() {
	log.SetFlags(0)

	f, err := os.Open("report.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	var entries []int64

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		i, err := strconv.ParseInt(scanner.Text(), 10, 64)
		if err != nil {
			log.Fatal(err)
		}

		entries = append(entries, i)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	pairs := [][2]int64{}
	for _, a := range entries {
		for _, b := range entries {
			pairs = append(pairs, [2]int64{a, b})
		}
	}

	for _, pair := range pairs {
		a, b := pair[0], pair[1]
		if a+b == 2020 {
			log.Printf("%d + %d = 2020", a, b)
			log.Printf("%d * %d = %d", a, b, a*b)
		}
	}

	threes := [][3]int64{}
	for _, a := range entries {
		for _, b := range entries {
			for _, c := range entries {
				threes = append(threes, [3]int64{a, b, c})
			}
		}
	}

	for _, three := range threes {
		a, b, c := three[0], three[1], three[2]
		if a+b+c == 2020 {
			log.Printf("%d + %d + %d = 2020", a, b, c)
			log.Printf("%d * %d * %d = %d", a, b, c, a*b*c)
		}
	}
}
