package main

import (
	"bytes"
	"fmt"
	"log"
	"sort"
)

func main() {
	log.SetFlags(log.Lshortfile)
	readInput()
	beamSearch()
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

func rdluNameToDirection(name string) int {
	for i, n := range rdluName {
		if n == name {
			return i
		}
	}
	panic("invalid name")
}

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

// canMove check if you can move from (i, j) in the direction d
func canMove(i, j, d int) bool {
	y := i + rdluPoint[d].y
	x := j + rdluPoint[d].x
	if y < 0 || x < 0 || y >= N || x >= N {
		return false
	}
	return !wallExists(i, j, d)
}

// --------------------------------------------------------------------
// beamsearch
// TODO: あらかじめ、各マスから行動できるマスを計算しておく

type Cell struct {
	lastVistidTime int
}

type State struct {
	turn                 int
	position             Point
	collectedTrashAmount int
	fields               [40][40]Cell
	OUTPUT               string
}

func (s *State) Clone() *State {
	rtn := *s
	rtn.position = s.position
	rtn.collectedTrashAmount = s.collectedTrashAmount
	rtn.fields = s.fields
	rtn.OUTPUT = s.OUTPUT
	return &rtn
}

func (s *State) nextState() (rtn *[]*State) {
	rtn = &[]*State{}
	for i := 0; i < 4; i++ {
		n := s.Clone()
		if n.move(i) {
			if n != nil {
				*rtn = append(*rtn, n)
			}
		}
	}
	return
}

// move returns true if the move was successful
func (s *State) move(d int) bool {
	if !canMove(s.position.y, s.position.x, d) {
		return false
	}
	s.turn++
	s.position.y += rdluPoint[d].y
	s.position.x += rdluPoint[d].x
	s.collectedTrashAmount += dirtiness[s.position.y][s.position.x] * (s.turn - s.fields[s.position.y][s.position.x].lastVistidTime)
	s.fields[s.position.y][s.position.x].lastVistidTime = s.turn
	s.OUTPUT += rdluName[d]
	return true
}

func beamSearch() {
	beamWidth := 200
	beamDepth := 2000
	nowState := State{position: Point{0, 0}}
	states := []*State{&nowState}
	for beamDepth > 0 {
		beamDepth--
		nextStates := []*State{}
		for _, state := range states {
			tmpstates := state.nextState()
			nextStates = append(nextStates, *tmpstates...)
		}
		sort.Slice(nextStates, func(i, j int) bool {
			return nextStates[i].collectedTrashAmount > nextStates[j].collectedTrashAmount
		})
		states = make([]*State, 0, beamWidth)
		states = nextStates[:MinInt(beamWidth, len(nextStates))]
	}
	checkOutput(states[0].OUTPUT)
	fmt.Println(states[0].OUTPUT)
}

func MinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func checkOutput(output string) {
	state := State{position: Point{0, 0}}
	var reached [40][40]bool
	for _, o := range output {
		state.move(rdluNameToDirection(string(o)))
		reached[state.position.y][state.position.x] = true
	}

	noReaches := make([]Point, 0, N*N)
	for i := 0; i < N; i++ {
		for j := 0; j < N; j++ {
			if !reached[i][j] {
				noReaches = append(noReaches, Point{i, j})
			}
		}
	}
	log.Println("noReaches:", len(noReaches), noReaches)

	if state.position.y != 0 || state.position.x != 0 {
		log.Println("not reached goal")
		return
	}
}
