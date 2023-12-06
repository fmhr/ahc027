package main

import (
	"bytes"
	"fmt"
	"log"
)

func main() {
	log.SetFlags(log.Lshortfile)
	readInput()
}

var N int
var hWall [40][40]bool
var vWall [40][40]bool
var dirtiness [40][40]int

func readInput() {
	_, err := fmt.Scan(&N)
	if err != nil {
		log.Fatal(err)
	}
	for i := 0; i < N-1; i++ {
		var s string
		_, err := fmt.Scan(&s)
		if err != nil {
			log.Fatal(err)
		}
		for j := 0; j < N; j++ {
			if s[j] == '1' {
				hWall[i][j] = true
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
				vWall[i][j] = true
			}
		}
	}
	for i := 0; i < N; i++ {
		for j := 0; j < N; j++ {
			_, err := fmt.Scan(&dirtiness[i][j])
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	gridView(dirtiness)
}

func gridView(grid [40][40]int) {
	var buffer bytes.Buffer
	buffer.WriteString("\n")
	for i := 0; i <= 2*N; i++ {
		for j := 0; j <= 2*N; j++ {
			switch {
			case i%2 == 0 && j%2 == 0:
				buffer.WriteString("+")
			case i == 0 || i == 2*N:
				buffer.WriteString("---")
			case j == 0 || j == 2*N:
				buffer.WriteString("|")
			case i%2 == 0:
				if hWall[i/2-1][(j-1)/2] {
					buffer.WriteString("---")
				} else {
					buffer.WriteString("   ")
				}
			case j%2 == 0:
				if vWall[(i-1)/2][j/2-1] {
					buffer.WriteString("|")
				} else {
					buffer.WriteString(" ")
				}
			default:
				y := (i - 1) / 2
				x := (j - 1) / 2
				buffer.WriteString(fmt.Sprintf("%3d", grid[y][x]))
			}
		}
		buffer.WriteString("\n")
	}
	fmt.Print(buffer.String())
}

// --------------------------------------------------------------------
// 共通
type Point struct {
	y, x int
}

const (
	Right = iota
	Down
	Left
	Up
)

var rdluPoint = []Point{{0, 1}, {1, 0}, {0, -1}, {-1, 0}}
var rdluName = []string{"R", "D", "L", "U"} // +2%4 で反対向き

// wallExists check if there is a wall in the direction d from (i, j)
func wallExists(i, j, d int) bool {
	switch d {
	case Right:
		return vWall[i][j]
	case Down:
		return hWall[i][j]
	case Left:
		return vWall[i][j-1]
	case Up:
		return hWall[i-1][j]
	default:
		panic("invalid direction")
	}
}

// --------------------------------------------------------------------
var visited [40][40]bool

func dfs(i, j int) {
	visited[i][j] = true
	for d := 0; d < 4; d++ {
		ny := i + rdluPoint[d].y
		nx := j + rdluPoint[d].x
		if ny < 0 || ny >= N || nx < 0 || nx >= N {
			continue
		}
		if visited[ny][nx] {
			continue
		}
		if wallExists(i, j, d) {
			continue
		}
		fmt.Println(rdluName[d])
		dfs(ny, nx)
	}
}
