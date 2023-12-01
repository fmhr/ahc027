package main

import (
	"fmt"
	"log"
)

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	readInput()
}

func readInput() {
	var N int
	_, err := fmt.Scan(&N)
	if err != nil {
		log.Fatal(err)
	}
	var h [40][40]int
	var v [40][40]int
	var d [40][40]int
	for i := 0; i < N-1; i++ {
		for j := 0; j < N; j++ {
			fmt.Scan(&h[i][j])
			_, err := fmt.Scan(&h[i][j])
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	for i := 0; i < N-1; i++ {
		for j := 0; j < N; j++ {
			_, err := fmt.Scan(&v[i][j])
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	for i := 0; i < N; i++ {
		for j := 0; j < N; j++ {
			_, err := fmt.Scan(&d[i][j])
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
