package main

import (
	"fmt"
	"log"
)

func main() {
	log.SetFlags(log.Lshortfile)
	readInput()
}

func readInput() {
	var N int
	_, err := fmt.Scan(&N)
	if err != nil {
		log.Fatal(err)
	}
	var h [40][40]bool
	var v [40][40]bool
	var d [40][40]int
	for i := 0; i < N-1; i++ {
		var s string
		_, err := fmt.Scan(&s)
		if err != nil {
			log.Fatal(err)
		}
		for j := 0; j < N; j++ {
			if s[j] == '1' {
				h[i][j] = true
			}
		}
	}
	for i := 0; i < N; i++ {
		var s string
		_, err := fmt.Scan(&s)
		if err != nil {
			log.Fatal(err)
		}
		for j := 0; j < N-1; j++ {
			if s[j] == '1' {
				v[i][j] = true
			}
		}
	}
	for i := 0; i < N; i++ {
		for j := 0; j < N-1; j++ {
			_, err := fmt.Scan(&d[i][j])
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	log.Println(N)
	for i := 0; i < N; i++ {
		log.Println(h[i])
	}
	for j := 0; j < N; j++ {
		log.Println(v[j])
	}
	for k := 0; k < N; k++ {
		log.Println(d[k][0 : N-1])
	}
}
