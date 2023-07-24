package lkh

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"os"
	"sort"
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
		G := [][]float64{
			{0, 21472, 24943, 26700},
			{21112, 0, 19515, 22141},
			{25633, 17347, 0, 4937},
			{28365, 22667, 5319, 0},
		}
		nodeName := []string{"New York", "Los Angeles", "Chicago", "Minneapolis", "Denver", "Dallas", "Seattle", "Boston", "San Francisco", "St. Louis", "Houston", "Phoenix", "Salt Lake City"}
		//selectNum := 79
		//var ori [][]float64
		//data, _ := Read("/Users/jinhanxie/Downloads/tsp_matrix.txt")
		//json.Unmarshal([]byte(data), &ori)
		////distanceJson, _ := json.Marshal(ori)
		////ioutil.WriteFile("/Users/jinhanxie/Downloads/200_matrix.json", distanceJson, os.ModePerm)
		//selectNum = len(ori)
		//G := make([][]float64, selectNum)
		//nodeName := make([]string, selectNum)
		//for i := 0; i < selectNum; i++ {
		//	G[i] = make([]float64, selectNum)
		//	for j := 0; j < selectNum; j++ {
		//		G[i][j] = ori[i][j]
		//	}
		//	nodeName[i] = strconv.Itoa(i)
		//}
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
	for selectNum := 150; selectNum > 110; selectNum = selectNum - 5 {
		lis := []float64{}
		row := []string{}
		//selectNum := 150
		row = append(row, fmt.Sprintf("%v", selectNum))
		for k := 0; k < 20; k++ {
			fmt.Printf("第%d轮：\n", k)
			var ori [][]float64
			data, _ := Read("/Users/jinhanxie/Downloads/tsp_matrix/CTBR230600345546_rota6_straight.txt")
			//data, _ := Read("/Users/jinhanxie/Downloads/200_matrix.json")
			//data, _ := Read("/Users/jinhanxie/Downloads/tsp_matrix.txt")
			json.Unmarshal([]byte(data), &ori)
			//distanceJson, _ := json.Marshal(ori)
			//ioutil.WriteFile("/Users/jinhanxie/Downloads/200_matrix.json", distanceJson, os.ModePerm)

			rawSlice := make([]int, 200)
			for i := 0; i < 200; i++ {
				rawSlice[i] = i
			}
			//dest := MicsSlice(rawSlice, selectNum)
			selectNum = len(ori)
			G := make([][]float64, selectNum)
			nodeName := make([]string, selectNum)
			for i := 0; i < selectNum; i++ {
				G[i] = make([]float64, selectNum)
				for j := 0; j < selectNum; j++ {
					//G[i][j] = ori[dest[i]][dest[j]]
					G[i][j] = ori[i][j]
				}
				nodeName[i] = strconv.Itoa(i)
			}
			//dis1 := G[2][6] + G[6][4]
			//dis2 := G[1][3] + G[3][5] + G[5][7] + G[7][0]
			//dis1 += 0
			//dis2 += 0
			//reserve := []int{0, 1, 2, 4, 6}
			//G1 := make([][]float64, 5)
			//nodeName = make([]string, 5)
			//for i := 0; i < 5; i++ {
			//	G1[i] = make([]float64, 5)
			//	nodeName[i] = strconv.Itoa(i)
			//	for j := 0; j < 5; j++ {
			//		G1[i][j] = G[reserve[i]][reserve[j]]
			//	}
			//}
			//G1[0][1] = 0
			//G1[1][0] = 0
			//G1 := [][]float64{{0, 24943, 26700, 21472}, {25633, 0, 4937, 17347}, {28365, 5319, 0, 22667}, {21112, 19515, 22141, 0}}
			//dis1 := G1[0][2] + G1[2][1] + G1[1][3] + G1[3][0]
			//dis1 += 0
			output := SolveTspOneway(TSPSolverInput{nodeName, G, nil, 0, 0, 500000, false})

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
			//innerDistDiffs := fmt.Sprintf("%v", innerDistDiff)
			//row := []string{solverInput.TaskId, total_routes1, total_avgDur1, totalCost1, totalDist1, totalInnerDist1, totalElapsed1, total_routes2, total_avgDur2, totalCost2, totalDist2, totalInnerDist2, totalElapsed2, costDiffs, distDiffs, innerDistDiffs}
			lis = append(lis, output.Distance)
		}
		sort.SliceStable(lis, func(i, j int) bool {
			return lis[i] < lis[j]
		})
		filePath := "/Users/jinhanxie/Downloads/result_tsp_0530f.csv"
		for i := 0; i < len(lis); i++ {
			row = append(row, fmt.Sprintf("%v", lis[i]))
		}
		WriteCSVFile(filePath, row)
	}
}

func TestSolveTspOnewayExhaustive(t *testing.T) {
	for k := 0; k < 1; k++ {
		fmt.Printf("第%d轮：\n", k)
		selectNum := 12
		var ori [][]float64
		//data, _ := Read("/Users/jinhanxie/Downloads/200_matrix.json")
		data, _ := Read("/Users/jinhanxie/Downloads/tsp_matrix.txt")
		json.Unmarshal([]byte(data), &ori)
		//distanceJson, _ := json.Marshal(ori)
		//ioutil.WriteFile("/Users/jinhanxie/Downloads/200_matrix.json", distanceJson, os.ModePerm)
		//selectNum = len(ori)
		rawSlice := make([]int, 200)
		for i := 0; i < 200; i++ {
			rawSlice[i] = i
		}
		//dest := MicsSlice(rawSlice, selectNum)
		G := make([][]float64, selectNum)
		nodeName := make([]string, selectNum)
		for i := 0; i < selectNum; i++ {
			G[i] = make([]float64, selectNum)
			for j := 0; j < selectNum; j++ {
				//G[i][j] = ori[dest[i]][dest[j]]
				G[i][j] = ori[i][j]
			}
			nodeName[i] = strconv.Itoa(i)
		}
		output := SolveTspOnewayExhaustive(TSPSolverInput{nodeName, G, nil, 0, 0, 3000, false})
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
func WriteCSVFile(fileName string, row []string) {
	//这样打开，每次都会清空文件内容
	//nfs, err := os.Create(newFileName)

	//这样可以追加写
	nfs, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalf("can not create file, err is %+v", err)
	}
	defer nfs.Close()
	nfs.Seek(0, io.SeekEnd)

	w := csv.NewWriter(nfs)
	//设置属性
	w.Comma = ','
	w.UseCRLF = true
	err = w.Write(row)
	if err != nil {
		log.Fatalf("can not write, err is %+v", err)
	}
	//这里必须刷新，才能将数据写入文件。
	w.Flush()
}
