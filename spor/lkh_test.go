package solver

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

func genGraph(size int) (weights [][]float64) {
	points := make([][2]int, 0)
	k := 1
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			points = append(points, [2]int{i * 100, j * 100})
			fmt.Printf("%d %d %d\n", k, i*100, j*100)
			k += 1
		}
	}
	weights = make([][]float64, len(points))
	for i := range weights {
		weights[i] = make([]float64, size*size)
		for j := 0; j < len(points); j++ {
			weights[i][j] = math.Sqrt(float64((points[i][0]-points[j][0])*(points[i][0]-points[j][0]) + (points[i][1]-points[j][1])*(points[i][1]-points[j][1])))
		}
	}

	return weights
}

func TestPrim(t *testing.T) {
	distance := genGraph(3)
	G, _ := checkDistMat(distance)
	nodes, _ := genNodes(nil, G, nil)
	err := prim(nodes, G)
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range nodes {
		if n.mstDad == nil {
			t.Logf("%d is the root", n.idx)
			continue
		}
		t.Logf("%d -> %d", n.idx, n.mstDad.idx)
	}
}

func TestMin1Tree(t *testing.T) {
	distance := genGraph(3)
	G, _ := checkDistMat(distance)
	nodes, _ := genNodes(nil, G, nil)
	cost, err := min1Tree(nodes, G)
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range nodes {
		if n.mstDad == nil {
			t.Logf("%d is the root", n.idx)
			continue
		}
		t.Logf("%d -> %d", n.idx, n.mstDad.idx)
		if isSpecialNode(n) {
			t.Logf("%d <-S-> %d", n.idx, n.oneTreeSucc.idx)
		}
	}
	t.Logf("Cost: %f", float64(cost)/float64(PRECISION))
}

func TestAscent(t *testing.T) {
	distance := genGraph(3)
	G, _ := checkDistMat(distance)
	nodes, _ := genNodes(nil, G, nil)
	t.Log("ascending")
	cost, err := ascent(nodes, G)
	if err != nil {
		t.Fatal(err)
	}
	//for i := 0; i < len(nodes); i++ {
	//	d := make([]float64, 0)
	//	for j := 0; j < len(nodes); j++ {
	//		d = append(d, nodes[i].distanceTo(nodes[j]))
	//	}
	//	fmt.Println(d)
	//}
	for _, n := range nodes {
		if n.mstDad == nil {
			t.Logf("%d is the root", n.idx)
			continue
		}
		t.Logf("%d -> %d", n.idx, n.mstDad.idx)
		if isSpecialNode(n) {
			t.Logf("%d <-S-> %d", n.idx, n.oneTreeSucc.idx)
		}
	}
	t.Logf("Cost: %f", float64(cost)/float64(PRECISION))
}

func TestGenCandidates(t *testing.T) {
	distance := genGraph(3)
	G, _ := checkDistMat(distance)
	nodes, _ := genNodes(nil, G, nil)
	_, err := min1Tree(nodes, G)
	if err != nil {
		t.Fatal(err)
	}
	err = genCandidates(nodes, G, math.MaxInt64, math.MaxFloat64)
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range nodes {
		fmt.Printf("Node %d: ", n.idx+1)
		for _, c := range n.candidates {
			fmt.Printf("(Candidate=%d) ", c.idx+1)
		}
		fmt.Printf("\n")
	}
}

func TestSolveTspOnewayInit(t *testing.T) {
	G := [][]float64{
		{0, 2451, 713, 1018, 1631, 1374, 2408, 213, 2571, 875, 1420, 2145, 1972},  // New York
		{2451, 0, 1745, 1524, 831, 1240, 959, 2596, 403, 1589, 1374, 357, 579},    // Los Angeles
		{713, 1745, 0, 355, 920, 803, 1737, 851, 1858, 262, 940, 1453, 1260},      // Chicago
		{1018, 1524, 355, 0, 700, 862, 1395, 1123, 1584, 466, 1056, 1280, 987},    // Minneapolis
		{1631, 831, 920, 700, 0, 663, 1021, 1769, 949, 796, 879, 586, 371},        // Denver
		{1374, 1240, 803, 862, 663, 0, 1681, 1551, 1765, 547, 225, 887, 999},      // Dallas
		{2408, 959, 1737, 1395, 1021, 1681, 0, 2493, 678, 1724, 1891, 1114, 701},  // Seattle
		{213, 2596, 851, 1123, 1769, 1551, 2493, 0, 2699, 1038, 1605, 2300, 2099}, // Boston
		{2571, 403, 1858, 1584, 949, 1765, 678, 2699, 0, 1744, 1645, 653, 600},    // San Francisco
		{875, 1589, 262, 466, 796, 547, 1724, 1038, 1744, 0, 679, 1272, 1162},     // St. Louis
		{1420, 1374, 940, 1056, 879, 225, 1891, 1605, 1645, 679, 0, 1017, 1200},   // Houston
		{2145, 357, 1453, 1280, 586, 887, 1114, 2300, 653, 1272, 1017, 0, 504},    // Phoenix
		{1972, 579, 1260, 987, 371, 999, 701, 2099, 600, 1162, 1200, 504, 0},      // Salt Lake City
	}
	nodeName := []string{"New York", "Los Angeles", "Chicago", "Minneapolis", "Denver", "Dallas", "Seattle", "Boston", "San Francisco", "St. Louis", "Houston", "Phoenix", "Salt Lake City"}
	output := SolveTspOnewayInit(TSPSolverInput{nodeName, G, nil, 9, 9, 2500, false})
	if output.Retcode != Success {
		t.Fatal(output.Err)
	} else {
		t.Log(output.Sequence)
		t.Log(output.Distance)
	}
}

func TestSolve(t *testing.T) {
	G := [][]float64{
		{0, 2451, 713, 1018, 1631, 1374, 2408, 213, 2571, 875, 1420, 2145, 1972},  // New York
		{2451, 0, 1745, 1524, 831, 1240, 959, 2596, 403, 1589, 1374, 357, 579},    // Los Angeles
		{713, 1745, 0, 355, 920, 803, 1737, 851, 1858, 262, 940, 1453, 1260},      // Chicago
		{1018, 1524, 355, 0, 700, 862, 1395, 1123, 1584, 466, 1056, 1280, 987},    // Minneapolis
		{1631, 831, 920, 700, 0, 663, 1021, 1769, 949, 796, 879, 586, 371},        // Denver
		{1374, 1240, 803, 862, 663, 0, 1681, 1551, 1765, 547, 225, 887, 999},      // Dallas
		{2408, 959, 1737, 1395, 1021, 1681, 0, 2493, 678, 1724, 1891, 1114, 701},  // Seattle
		{213, 2596, 851, 1123, 1769, 1551, 2493, 0, 2699, 1038, 1605, 2300, 2099}, // Boston
		{2571, 403, 1858, 1584, 949, 1765, 678, 2699, 0, 1744, 1645, 653, 600},    // San Francisco
		{875, 1589, 262, 466, 796, 547, 1724, 1038, 1744, 0, 679, 1272, 1162},     // St. Louis
		{1420, 1374, 940, 1056, 879, 225, 1891, 1605, 1645, 679, 0, 1017, 1200},   // Houston
		{2145, 357, 1453, 1280, 586, 887, 1114, 2300, 653, 1272, 1017, 0, 504},    // Phoenix
		{1972, 579, 1260, 987, 371, 999, 701, 2099, 600, 1162, 1200, 504, 0},      // Salt Lake City
	}
	nodeName := []string{"New York", "Los Angeles", "Chicago", "Minneapolis", "Denver", "Dallas", "Seattle", "Boston", "San Francisco", "St. Louis", "Houston", "Phoenix", "Salt Lake City"}

	//fix := [][2]int{
	//	{12, 0},
	//}

	seq, dis, rc, err := Solve(nodeName, G, nil, time.Now().Add(2500*time.Millisecond))
	if rc != Success {
		t.Fatal(err)
	} else {
		t.Log(seq)
		t.Log(dis)
	}
}

func TestSolveTspAsymmetrical(t *testing.T) {
	for k := 0; k < 10; k++ {
		//ori := [][]float64{
		//	{0, 2451, 713, 1018, 1631, 1374, 2408, 213, 2571, 875, 1420, 2145, 1972},  // New York
		//	{2451, 0, 1745, 1524, 831, 1240, 959, 2596, 403, 1589, 1374, 357, 579},    // Los Angeles
		//	{713, 1745, 0, 355, 920, 803, 1737, 851, 1858, 262, 940, 1453, 1260},      // Chicago
		//	{1018, 1524, 355, 0, 700, 862, 1395, 1123, 1584, 466, 1056, 1280, 987},    // Minneapolis
		//	{1631, 831, 920, 700, 0, 663, 1021, 1769, 949, 796, 879, 586, 371},        // Denver
		//	{1374, 1240, 803, 862, 663, 0, 1681, 1551, 1765, 547, 225, 887, 999},      // Dallas
		//	{2408, 959, 1737, 1395, 1021, 1681, 0, 2493, 678, 1724, 1891, 1114, 701},  // Seattle
		//	{213, 2596, 851, 1123, 1769, 1551, 2493, 0, 2699, 1038, 1605, 2300, 2099}, // Boston
		//	{2571, 403, 1858, 1584, 949, 1765, 678, 2699, 0, 1744, 1645, 653, 600},    // San Francisco
		//	{875, 1589, 262, 466, 796, 547, 1724, 1038, 1744, 0, 679, 1272, 1162},     // St. Louis
		//	{1420, 1374, 940, 1056, 879, 225, 1891, 1605, 1645, 679, 0, 1017, 1200},   // Houston
		//	{2145, 357, 1453, 1280, 586, 887, 1114, 2300, 653, 1272, 1017, 0, 504},    // Phoenix
		//	{1972, 579, 1260, 987, 371, 999, 701, 2099, 600, 1162, 1200, 504, 0},      // Salt Lake City
		//}
		//nodeName := []string{"New York", "Los Angeles", "Chicago", "Minneapolis", "Denver", "Dallas", "Seattle", "Boston", "San Francisco", "St. Louis", "Houston", "Phoenix", "Salt Lake City"}
		selectNum := 79
		var ori [][]float64
		data, _ := Read("/Users/jinhanxie/Downloads/200_matrix.json")
		json.Unmarshal([]byte(data), &ori)
		//distanceJson, _ := json.Marshal(ori)
		//ioutil.WriteFile("/Users/jinhanxie/Downloads/200_matrix.json", distanceJson, os.ModePerm)
		G := make([][]float64, selectNum)
		nodeName := make([]string, selectNum)
		for i := 0; i < selectNum; i++ {
			G[i] = make([]float64, selectNum)
			for j := 0; j < selectNum; j++ {
				G[i][j] = ori[i][j]
			}
			nodeName[i] = strconv.Itoa(i)
		}
		//fix := [][2]int{
		//	{12, 0},
		//}
		seq, dis, rc, err := SolveTspAsymmetrical(nodeName, G, nil, time.Now().Add(2500*time.Millisecond))
		if rc != Success {
			t.Fatal(err)
		} else {
			t.Log(seq)
			t.Log(dis)
		}
	}
}

func TestSolveTspOneway(t *testing.T) {
	for k := 0; k < 20; k++ {
		fmt.Printf("第%d轮：\n", k)
		selectNum := 79
		var ori [][]float64
		data, _ := Read("/Users/jinhanxie/Downloads/200_matrix.json")
		json.Unmarshal([]byte(data), &ori)
		//distanceJson, _ := json.Marshal(ori)
		//ioutil.WriteFile("/Users/jinhanxie/Downloads/200_matrix.json", distanceJson, os.ModePerm)
		rawSlice := make([]int, 200)
		for i := 0; i < 200; i++ {
			rawSlice[i] = i
		}
		dest := MicsSlice(rawSlice, selectNum)
		G := make([][]float64, selectNum)
		nodeName := make([]string, selectNum)
		for i := 0; i < selectNum; i++ {
			G[i] = make([]float64, selectNum)
			for j := 0; j < selectNum; j++ {
				G[i][j] = ori[dest[i]][dest[j]]
			}
			nodeName[i] = strconv.Itoa(i)
		}
		//fix := [][2]int{
		//	{12, 0},
		//}
		output := SolveTspOneway(TSPSolverInput{nodeName, G, nil, 5, 5, 2500, false})
		//seq, dis, rc, err := SolveTspOneway(nodeName, G, nil, 0, 5, false)
		//seq, dis, rc, err := SolveTspOneway(nodeName, G, nil, 0, -1, false)
		//seq, dis, rc, err := SolveTspOneway(nodeName, G, nil, 5, 5, true)
		//seq, dis, rc, err := SolveTspOneway(nodeName, G, nil, 0, 5, true)
		//seq, dis, rc, err := SolveTspOneway(nodeName, G, nil, 0, -1, true)
		if output.Retcode != Success {
			t.Fatal(output.Err)
		} else {
			t.Log(output.Sequence)
			t.Log(output.Distance)
		}
	}
}
func BenchmarkSolve(b *testing.B) {
	G := [][]float64{
		{0, 2451, 713, 1018, 1631, 1374, 2408, 213, 2571, 875, 1420, 2145, 1972},  // New York
		{2451, 0, 1745, 1524, 831, 1240, 959, 2596, 403, 1589, 1374, 357, 579},    // Los Angeles
		{713, 1745, 0, 355, 920, 803, 1737, 851, 1858, 262, 940, 1453, 1260},      // Chicago
		{1018, 1524, 355, 0, 700, 862, 1395, 1123, 1584, 466, 1056, 1280, 987},    // Minneapolis
		{1631, 831, 920, 700, 0, 663, 1021, 1769, 949, 796, 879, 586, 371},        // Denver
		{1374, 1240, 803, 862, 663, 0, 1681, 1551, 1765, 547, 225, 887, 999},      // Dallas
		{2408, 959, 1737, 1395, 1021, 1681, 0, 2493, 678, 1724, 1891, 1114, 701},  // Seattle
		{213, 2596, 851, 1123, 1769, 1551, 2493, 0, 2699, 1038, 1605, 2300, 2099}, // Boston
		{2571, 403, 1858, 1584, 949, 1765, 678, 2699, 0, 1744, 1645, 653, 600},    // San Francisco
		{875, 1589, 262, 466, 796, 547, 1724, 1038, 1744, 0, 679, 1272, 1162},     // St. Louis
		{1420, 1374, 940, 1056, 879, 225, 1891, 1605, 1645, 679, 0, 1017, 1200},   // Houston
		{2145, 357, 1453, 1280, 586, 887, 1114, 2300, 653, 1272, 1017, 0, 504},    // Phoenix
		{1972, 579, 1260, 987, 371, 999, 701, 2099, 600, 1162, 1200, 504, 0},      // Salt Lake City
	}
	nodeName := []string{"New York", "Los Angeles", "Chicago", "Minneapolis", "Denver", "Dallas", "Seattle", "Boston", "San Francisco", "St. Louis", "Houston", "Phoenix", "Salt Lake City"}
	for n := 0; n < b.N; n++ {
		_, _, _, _ = Solve(nodeName, G, nil, time.Now().Add(2500*time.Millisecond))
	}
}
func MicsSlice(origin []int, count int) []int {
	tmpOrigin := make([]int, len(origin))
	copy(tmpOrigin, origin)
	//一定要seed
	rand.Seed(time.Now().Unix())
	rand.Shuffle(len(tmpOrigin), func(i int, j int) {
		tmpOrigin[i], tmpOrigin[j] = tmpOrigin[j], tmpOrigin[i]
	})

	result := make([]int, 0, count)
	for index, value := range tmpOrigin {
		if index == count {
			break
		}
		result = append(result, value)
	}
	return result
}
func Read(filePath string) (string, bool) {
	byteArr, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Println("read file error -path:", filePath, err)
		return "", false
	}
	return string(byteArr), true
}
