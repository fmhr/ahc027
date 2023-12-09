package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"sort"
	"sync"
	"time"
)

// ./bin/a.out -cpuprofile cpu.prof < tools/in/0000.txt
// ./bin/a.out -cpuprofile cpu.prof -memprofile mem.prof < tools/in/0000.txt
// go tool pprof -http=localhost:8888 bin/a.out cpu.prof
// go tool pprof -http=localhost:8888 bin/a.out mem.prof
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
var memprofile = flag.String("memprofile", "", "write memory profile to `file`")

func main() {
	log.SetFlags(log.Lshortfile)
	// GCの閾値を高く設定して、GCの実行頻度を減らす
	//debug.SetGCPercent(2000)
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
	var m runtime.MemStats
	//runtime.ReadMemStats(&m)
	//fmt.Printf("Allocations before: %v\n", m.Mallocs)
	// 実際の処理 --------------------------------------------------
	startTime := time.Now()
	readInput()
	beamSearch()
	duration := time.Since(startTime)
	log.Printf("time=%vs", duration.Seconds())
	//log.Println("getCount:", getCount, "putCount:", putCount)
	// -----------------------------------------------------------
	// メモリ使用量を表示
	runtime.ReadMemStats(&m)
	//log.Printf("Allocations after: %v\n", m.Mallocs)
	//log.Printf("TotalAlloc: %v\n", m.TotalAlloc)
	//log.Printf("NumGC: %v\n", m.NumGC)
	//log.Printf("NumForcedGC: %v\n", m.NumForcedGC)
	//log.Printf("MemPauseTotal: %vms\n", float64(m.PauseTotalNs)/1000/1000) // ナノ、マイクロ、ミリ
	// Allocは現在ヒープに割り当てられているバイト数を返します
	//log.Printf("Alloc = %v MiB", m.Alloc/1024/1024)
	// TotalAllocはプログラム開始以来割り当てられた全バイト数を返します
	//log.Printf("TotalAlloc = %v MiB", m.TotalAlloc/1024/1024)
	// SysはOSから取得した全バイト数を返します
	log.Printf("Sys=%v MiB", m.Sys/1024/1024)
	// NumGCはプログラム開始以来のGC実行回数を返します
	//log.Printf("NumGC = %v\n", m.NumGC)
	// memory profile
	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		runtime.GC()    // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
	}
}

var N int
var hWall [40][40]bool
var vWall [40][40]bool
var dirtiness [40][40]uint16

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
			sumDirtiness += int(dirtiness[i][j])
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

type State struct {
	flag                 bool
	turn                 int
	position             Point
	collectedTrashAmount int
	lastVistidTime       [40][40]uint16
	nodeAddress          *Node
}

// sync.Pool
var pool = sync.Pool{
	New: func() interface{} {
		return &State{}
	},
}
var getCount int
var putCount int

func GetState() *State {
	getCount++
	return pool.Get().(*State)
}

func PutState(s *State) {
	if s == nil {
		return
	}
	putCount++
	s.turn = 0
	s.position = Point{0, 0}
	s.collectedTrashAmount = 0
	s.lastVistidTime = [40][40]uint16{}
	pool.Put(s)
}

func (s *State) outputToStringForTree() string {
	var buffer bytes.Buffer
	node := s.nodeAddress
	for node.Parent != nil {
		buffer.WriteString(rdluName[node.Move])
		node = node.Parent
	}
	bytes := buffer.Bytes()
	for i, j := 0, len(bytes)-1; i < j; i, j = i+1, j-1 {
		bytes[i], bytes[j] = bytes[j], bytes[i]
	}
	return string(bytes)
}

func (s *State) Clone() *State {
	rtn := GetState()
	rtn.turn = s.turn
	rtn.position = s.position
	rtn.collectedTrashAmount = s.collectedTrashAmount
	rtn.lastVistidTime = s.lastVistidTime
	rtn.nodeAddress = s.nodeAddress
	//log.Printf("rtn=%p s=%p %v\n", &rtn, s, &rtn == s)
	return rtn
}

func (src *State) Copy(dst *State) {
	dst.flag = src.flag
	dst.turn = src.turn
	dst.position.y = src.position.y
	dst.position.x = src.position.x
	dst.lastVistidTime = src.lastVistidTime
	dst.collectedTrashAmount = src.collectedTrashAmount
	dst.nodeAddress = src.nodeAddress
}

// func (s *State) nextState(next *[beamWidth * 4]State, nextIndex *int) {
func (s *State) nextState(next []*State, nextIndex *int, tree *Tree) {
	var cnt int = 0
	for i := 0; i < 4; i++ {
		s.Copy(next[*nextIndex])
		if next[*nextIndex].move(i) {
			next[*nextIndex].flag = true
			// tree update
			p := s.nodeAddress
			c := tree.AddChild(p, uint8(i), s.turn)
			next[*nextIndex].nodeAddress = c
			*nextIndex++
			cnt++
		}
	}
	// 四方塞がれているマスには到達しないので、ここは未達
	//if cnt == 0 {
	//tree.Release(s.nodeAddress)
	//}
}

// move returns true if the move was successful
func (s *State) move(d int) bool {
	if !canMove(s.position.y, s.position.x, d) {
		return false
	}
	s.position.y += rdluPoint[d].y
	s.position.x += rdluPoint[d].x
	// 汚れの総和
	s.collectedTrashAmount += int(dirtiness[s.position.y][s.position.x]) * (s.turn - int(s.lastVistidTime[s.position.y][s.position.x]))
	if s.lastVistidTime[s.position.y][s.position.x] == 0 {
		// 初めて訪れるマスにボーナス
		s.collectedTrashAmount += 100 * (s.turn + 1)
	} else {
		// 久しぶりに訪れるマスにボーナス
		s.collectedTrashAmount += 10 * (s.turn - int(s.lastVistidTime[s.position.y][s.position.x]))
	}
	s.lastVistidTime[s.position.y][s.position.x] = uint16(s.turn)
	s.turn++
	return true
}

// Goal (0,0)
func (s *State) toGoal() string {
	var buffer bytes.Buffer
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
					buffer.WriteString(rdluName[i])
					break
				}
			}
		}
	}
	return buffer.String()
}

const beamWidth = 40
const beamDepth = 10000

var nowArr, nextArr [beamWidth * 4]State

func beamSearch() {
	tree := NewTree()
	tree.Root = tree.NewNode(nil, 0, 0)
	nowSlice := make([]*State, beamWidth*4)
	nextSlice := make([]*State, beamWidth*4)
	for i := 0; i < beamWidth*4; i++ {
		nowSlice[i] = &nowArr[i]
	}
	for i := 0; i < beamWidth*4; i++ {
		nextSlice[i] = &nextArr[i]
	}
	now, next := nowSlice, nextSlice
	//	now, next := &nowArr, &nextArr
	now[0].flag = true // first(0, 0)
	now[0].nodeAddress = tree.Root
	for i := 0; beamDepth > i; i++ {
		nextIndex := 0
		for j := 0; j < beamWidth; j++ {
			if now[j].flag && now[j].turn == i {
				now[j].nextState(next, &nextIndex, tree) // nextに追加
			}
		}
		sort.Slice(next[:nextIndex], func(i, j int) bool {
			return next[i].collectedTrashAmount > next[j].collectedTrashAmount
		})
		for j := 0; j < beamWidth*4; j++ {
			n := now[j].nodeAddress
			if n == nil || n == tree.Root {
				continue
			}
			if n.Children[0] == nil && n.Children[1] == nil && n.Children[2] == nil && n.Children[3] == nil {
				tree.Release(n)
			}
			now[j].nodeAddress = nil
		}
		if nextIndex == 0 {
			break
		}
		now, next = next, now
	}
	// 最後にゴールに向かうのはnext
	rtn := now[0].toGoal()
	//checkOutput(now[0].outputToString())
	//fmt.Println(now[0].outputToString())
	//log.Println(len(now[0].outputToStringForTree()), len(now[0].outputToString()))
	//log.Println(len(rtn), rtn)
	ans := now[0].outputToStringForTree() + rtn
	fmt.Println(ans)
	//log.Println(time.Now())
	//log.Println(tree.Root)
	//tree.TraverseFromChildren(tree.Root)
	//log.Println(cnt)
	//log.Println(time.Now())
}

func MinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// 行動履歴を探索木を作って、コピーコストを減らす
type Node struct {
	Parent   *Node
	Children [4]*Node // RDLU毎に子ノードを持つ
	Move     uint8
}

type Tree struct {
	Root *Node
	pool sync.Pool
}

func NewTree() *Tree {
	return &Tree{
		pool: sync.Pool{
			New: func() interface{} {
				return &Node{}
			},
		},
	}
}

func (t *Tree) NewNode(parent *Node, move uint8, turn int) *Node {
	node := t.pool.Get().(*Node)
	node.Parent = parent
	node.Move = move
	return node
}

func (t *Tree) AddChild(parent *Node, move uint8, turn int) *Node {
	child := t.NewNode(parent, move, turn)
	parent.Children[move] = child
	return child
}

func (t *Tree) Release(node *Node) {
	node.Parent.Children[node.Move] = nil
	//if node.Parent.Children[0] == nil && node.Parent.Children[1] == nil && node.Parent.Children[2] == nil && node.Parent.Children[3] == nil {
	//if node.Parent.Parent != nil {
	//t.Release(node.Parent)
	//}
	//}
	// reset
	node.Parent = nil
	node.Move = 0
	for i := 0; i < 4; i++ {
		node.Children[i] = nil
	}
	t.pool.Put(node)
}

var cnt int

func (t *Tree) TraverseFromChildren(node *Node) {
	if node == nil {
		return
	}

	for _, child := range node.Children {
		t.TraverseFromChildren(child)
	}
	// ここで処理
	//if node.Children[0] == nil && node.Children[1] == nil && node.Children[2] == nil && node.Children[3] == nil {
	//t.Release(node)
	//}
	//log.Println(node)
	cnt++
}

//Error seeds: []
//avarageTime=0.30  maxTime=0.31
//(Score)sum=477,862,241.00 avarage=9,557,244.82 log=19.984833
//477862241.00
