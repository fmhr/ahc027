package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"sort"
	"sync"
	"time"
)

// ./bin/main -cpuprofile cpuprof < tools/in/0000.txt
// go tool pprof -http=localhost:8888 bin/main cpuprof
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")

//var memprofile = flag.String("memprofile", "", "write memory profile to `file`")

func main() {
	log.SetFlags(log.Lshortfile)
	// CPU profile
	flag.Parse()
	if *cpuprofile != "" {
		log.Println("CPU profile enabled")
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}
	// メモリ使用量を記録
	//var m runtime.MemStats
	//runtime.ReadMemStats(&m)
	//fmt.Printf("Allocations before: %v\n", m.Mallocs)
	// 実際の処理
	startTime := time.Now()
	readInput()
	beamSearch()
	duration := time.Since(startTime)
	log.Printf("time=%vs", duration.Seconds())
	// メモリ使用量を表示
	//runtime.ReadMemStats(&m)
	//log.Printf("Allocations after: %v\n", m.Mallocs)
	//log.Printf("TotalAlloc: %v\n", m.TotalAlloc)
	//log.Printf("NumGC: %v\n", m.NumGC)
	//log.Printf("NumForcedGC: %v\n", m.NumForcedGC)
	//log.Printf("MemPauseTotal: %vms\n", float64(m.PauseTotalNs)/1000/1000) // ナノ、マイクロ、ミリ
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
	sumDirtiness := 0
	for i := 0; i < N; i++ {
		for j := 0; j < N; j++ {
			_, err := fmt.Scan(&dirtiness[i][j])
			if err != nil {
				log.Fatal(err)
			}
			sumDirtiness += dirtiness[i][j]
		}
	}
	log.Printf("N=%v dirty=%v sumdirty=%v\n", N, sumDirtiness/(N*N), sumDirtiness)
	//gridView(dirtiness)
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
	log.Printf("\n %s\n", buffer.String())
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
	output               [20000]int8
}

// sync.Pool
var pool = sync.Pool{
	New: func() interface{} {
		return &State{}
	},
}

func GetState() *State {
	return pool.Get().(*State)
}

func PutState(s *State) {
	if s == nil {
		return
	}
	s.turn = 0
	s.position = Point{0, 0}
	s.collectedTrashAmount = 0
	s.fields = [40][40]Cell{}
	s.output = [20000]int8{}
	pool.Put(s)
}

func (s *State) outputToString() string {
	var buffer bytes.Buffer
	for i, o := range s.output {
		if i >= s.turn {
			break
		}
		buffer.WriteString(rdluName[o])
	}
	return buffer.String()
}

func (s *State) Clone() *State {
	rtn := GetState()
	rtn.turn = s.turn
	rtn.position = s.position
	rtn.collectedTrashAmount = s.collectedTrashAmount
	rtn.fields = s.fields
	rtn.output = s.output
	//log.Printf("rtn=%p s=%p %v\n", &rtn, s, &rtn == s)
	return rtn
}

func (s *State) nextState() (rtn *[]*State) {
	rtn = &[]*State{}
	for i := 0; i < 4; i++ {
		n := s.Clone()
		if n.move(i) {
			*rtn = append(*rtn, n)
		} else {
			PutState(n)
		}
	}
	return
}

// move returns true if the move was successful
func (s *State) move(d int) bool {
	if !canMove(s.position.y, s.position.x, d) {
		return false
	}
	s.position.y += rdluPoint[d].y
	s.position.x += rdluPoint[d].x
	s.collectedTrashAmount += dirtiness[s.position.y][s.position.x] * (s.turn - s.fields[s.position.y][s.position.x].lastVistidTime)
	if s.fields[s.position.y][s.position.x].lastVistidTime == 0 {
		s.collectedTrashAmount += 100 * (s.turn + 1)
	} else {
		s.collectedTrashAmount += 10 * (s.turn - s.fields[s.position.y][s.position.x].lastVistidTime)
	}
	s.fields[s.position.y][s.position.x].lastVistidTime = s.turn
	s.output[s.turn] = int8(d)
	s.turn++
	return true
}

// Goal (0,0)
func (s *State) toGoal() {
	// goalからの距離を計算
	// 現在地からgoalを目指す
	var distance [40][40]int
	points := []Point{{0, 0}}
	reached := [40][40]bool{}
	reached[0][0] = true
	for len(points) > 0 {
		now := points[0]
		points = points[1:]
		for i := 0; i < 4; i++ {
			if canMove(now.y, now.x, i) {
				next := Point{now.y + rdluPoint[i].y, now.x + rdluPoint[i].x}
				if !reached[next.y][next.x] || distance[next.y][next.x] > distance[now.y][now.x]+1 {
					distance[next.y][next.x] = distance[now.y][now.x] + 1
					reached[next.y][next.x] = true
					points = append(points, next)
				}
			}
		}
	}
	//	gridView(distance)
	for s.position.y != 0 || s.position.x != 0 {
		for i := 0; i < 4; i++ {
			if canMove(s.position.y, s.position.x, i) {
				next := Point{s.position.y + rdluPoint[i].y, s.position.x + rdluPoint[i].x}
				if distance[next.y][next.x] < distance[s.position.y][s.position.x] {
					s.move(i)
					break
				}
			}
		}
	}
}

const beamWidth = 10
const beamDepth = 10000

func beamSearch() {
	nowState := State{}
	var states [beamWidth]*State
	states[0] = &nowState
	var nextStates [beamWidth * 4]*State
	for i := 0; beamDepth > i; i++ {
		var nindex int
		for _, state := range states {
			if state == nil {
				continue
			}
			tmpstates := state.nextState()
			for _, tmpstate := range *tmpstates {
				nextStates[nindex] = tmpstate
				nindex++
			}
		}
		sort.Slice(nextStates[:nindex], func(i, j int) bool {
			return nextStates[i].collectedTrashAmount > nextStates[j].collectedTrashAmount
		})
		for i, state := range states {
			PutState(state)
			states[i] = nil
		}
		for i, state := range nextStates {
			if i < beamWidth {
				states[i] = state
			} else {
				PutState(state)
			}
			nextStates[i] = nil
		}
	}
	states[0].toGoal()
	checkOutput(states[0].outputToString())
	fmt.Println(states[0].outputToString())
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
		rtn := state.move(rdluNameToDirection(string(o)))
		if !rtn {
			log.Println("invalid output")
		}
		reached[state.position.y][state.position.x] = true
	}

	noReaches := make([]Point, 0, N*N)
	for i := 0; i < N; i++ {
		//	log.Println(reached[i][:N])
		for j := 0; j < N; j++ {
			if !reached[i][j] {
				noReaches = append(noReaches, Point{i, j})
			}
		}
	}
	log.Println("noReaches:", len(noReaches), "/", N*N, noReaches)

	if state.position.y != 0 || state.position.x != 0 {
		log.Println("not reached goal")
		return
	}
}
