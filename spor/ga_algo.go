package solver

import (
	"math"
	"math/rand"
	"sort"
	"time"
)

const (
	MaxPopulation        = 100
	MaxIllegalPopulation = 30
	BreedRate            = 0.3
	MutationRate         = 0.1
	BreedLogBase         = 0.75
	MutationLogBase      = 0.15
)

// **************************************** draft **********************************************************
// **************************************** add accept *******************************************************
//var innerSeqs [][]int
//var InnerAsgmts []int
//var TaskLoc []int   // loc idx for each task idx
//var Cost [][][][]float64 // Cost between loc idx [resIdx][type(0:dist,1:dur)][][]
type Ethnicity struct {
	Seq [][]int
	Asg []int
}

func MethodGA(ethnicity *[]Ethnicity, optEndTime int64, gPara *GPara) (Ethnicity, bool) {
	var LoopNum int = 300
	InResMap := make(map[float64]int)
	tmp := []Ethnicity{}
	for i, per := range *ethnicity {
		if len(per.Seq) != len(per.Asg) {
			return (*ethnicity)[0], false
		}
		tmp = append(tmp, per)
		dis := GetDistance(per.Seq, per.Asg, gPara)
		_, ok := InResMap[dis]
		if !ok {
			InResMap[dis] = i
		}
	}
	ethnicity = &tmp
	//startTime := time.Now() // 获取当前时间
	population := len(InResMap)
	for LoopNum > 0 {
		nowTime := time.Now().Unix()
		if nowTime >= optEndTime {
			break
		}
		Breed(ethnicity, BreedRate, BreedLogBase, optEndTime, gPara)
		Mutation(ethnicity, MutationRate, MutationLogBase, optEndTime, gPara)
		sort.SliceStable(*ethnicity, func(i, j int) bool {
			return SortPerson((*ethnicity)[i].Seq, (*ethnicity)[i].Asg, gPara) < SortPerson((*ethnicity)[i].Seq, (*ethnicity)[i].Asg, gPara)
		})
		//去重并限制人口
		res := []Ethnicity{}
		disMap := make(map[float64]int)
		legalPerson := 0
		illegalPerson := 0
		for i, per := range *ethnicity {
			dis := GetDistance(per.Seq, per.Asg, gPara)
			_, ok := disMap[dis]
			if !ok {
				disMap[dis] = i
				if IsLegal(per.Seq, gPara, per.Asg) && legalPerson < MaxPopulation {
					legalPerson += 1
					res = append(res, (*ethnicity)[i])
				} else if !IsLegal(per.Seq, gPara, per.Asg) && illegalPerson <= MaxIllegalPopulation {
					illegalPerson += 1
					res = append(res, (*ethnicity)[i])
				}
			}
		}
		ethnicity = &res
		if LoopNum%25 == 1 {
			if legalPerson == 0 {
				return (*ethnicity)[0], false
			}
			//elapsed := time.Since(startTime)
			//latency := float64(elapsed) / 1000000000
			//startTime = time.Now()
			//fmt.Printf("==========Info:=======\n%d\n distance: %f\n use time: %f s\n population:%d\n InputPopulation:%d\n", LoopNum, GetDistance((*ethnicity)[0].Seq, (*ethnicity)[0].Asg, gPara), latency, legalPerson, len(InResMap))
			if legalPerson == population && population != MaxPopulation {
				//fmt.Printf("==========Info:=======\n%d\n distance: %f\n use time: %f s\n population:%d\n InputPopulation:%d\n", LoopNum, GetDistance((*ethnicity)[0].Seq, (*ethnicity)[0].Asg, gPara), latency, legalPerson, len(InResMap))
				break
			}
			population = legalPerson
		}
		//if len((*ethnicity)) > MaxPopulation {
		//	(*ethnicity) = (*ethnicity)[:MaxPopulation]
		//}
		LoopNum--
	}
	for index, person := range *ethnicity {
		if !IsLegal(person.Seq, gPara, person.Asg) {
			continue
		}
		res := GetDistance(person.Seq, person.Asg, gPara)
		_, ok := InResMap[res]
		if !ok {
			//fmt.Printf("距离：%f\n", res)
			return (*ethnicity)[index], true
		}
	}
	return (*ethnicity)[0], false
}
func IsLegalTask(newPerson [][]int, lenth int, asg []int) bool {
	xx := 0
	maps := make(map[int]bool)
	for i := 0; i < len(newPerson); i++ {
		xx += len(newPerson[i])
		for j := 0; j < len(newPerson[i]); j++ {
			if _, ok := maps[newPerson[i][j]]; ok {
				return false
			} else {
				maps[newPerson[i][j]] = true
			}
		}
	}
	if xx != lenth || len(newPerson) != len(asg) {
		return false
	}
	return true
}
func IsLegal(newPerson [][]int, gPara *GPara, asg []int) bool {
	for i := 0; i < len(newPerson); i++ {
		var totalParcel float64 = 0.0
		var totalDistance float64 = 0.0
		var totalDuration float64 = 0.0
		var totalWeight float64 = 0.0
		for j := 0; j < len(newPerson[i]); j++ {
			totalParcel += gPara.CapTask[newPerson[i][j]][0]
			totalWeight += gPara.CapTask[newPerson[i][j]][1]
			totalDuration += gPara.CapTask[newPerson[i][j]][2]
		}
		totalDistance = GetSeqDistance(newPerson[i], i, asg, gPara)
		totalDuration += GetSeqDuration(newPerson[i], i, asg, gPara)
		if !(totalDistance >= gPara.CapRes[asg[i]][0] && totalDuration >= gPara.CapRes[asg[i]][6] && totalParcel >= gPara.CapRes[asg[i]][2] && totalWeight >= gPara.CapRes[asg[i]][4]) {
			return false
		}
		if !(totalDistance <= gPara.CapRes[asg[i]][1] && totalDuration <= gPara.CapRes[asg[i]][7] && totalParcel <= gPara.CapRes[asg[i]][3] && totalWeight <= gPara.CapRes[asg[i]][5]) {
			return false
		}
	}
	return true
}
func Breed(ethnicity *[]Ethnicity, rate, logBase float64, optEndTime int64, gPara *GPara) {
	num := int(float64(len(*ethnicity))*rate) + 1 //繁殖couple数量
	for num > 0 {
		nowTime := time.Now().Unix()
		if nowTime >= optEndTime {
			break
		}
		num--
		BreedCal(ethnicity, logBase, optEndTime, gPara)
	}
}
func BreedCal(ethnicity *[]Ethnicity, logBase float64, optEndTime int64, gPara *GPara) {
	if len(*ethnicity) <= 1 {
		return
	}
	loop := int(math.Log(rand.Float64()) / math.Log(logBase))
	for ; loop > 0; loop-- {
		nowTime := time.Now().Unix()
		if nowTime >= optEndTime {
			break
		}
		parent1 := FirstProbability(*ethnicity, rand.Float64(), "best", gPara) //表现好的个体繁殖概率大
		parent2 := FirstProbability(*ethnicity, rand.Float64(), "best", gPara)
		tryNum := 50
		for parent2 == parent1 && tryNum > 0 {
			tryNum--
			parent2 = FirstProbability(*ethnicity, rand.Float64(), "best", gPara)
		}
		father := (*ethnicity)[parent1]
		mother := (*ethnicity)[parent2]
		if len(father.Seq) != len(mother.Seq) {
			continue
		}
		usedMessage := []int{}
		for i := 0; i < gPara.NTask; i++ {
			usedMessage = append(usedMessage, 1)
		}
		for i := 0; i < len(father.Seq); i++ {
			for j := 0; j < len(father.Seq[i]); j++ {
				usedMessage[father.Seq[i][j]] = 0
			}
		}
		for i := 0; i < len(mother.Seq); i++ {
			for j := 0; j < len(mother.Seq[i]); j++ {
				usedMessage[mother.Seq[i][j]] = 0
			}
		}
		childLenth := 0
		for i := 0; i < len(usedMessage); i++ {
			if usedMessage[i] == 0 {
				childLenth++
			}
		}
		child := [][]int{}
		//x := math.Minfloat64(float64(len(father)), float64(len(mother)))
		for i := 0; i < len(father.Asg); i++ {
			if i%2 == 0 {
				ParentSwap(father.Seq[i], mother.Seq[i], &child, &usedMessage)
			} else {
				ParentSwap(mother.Seq[i], father.Seq[i], &child, &usedMessage)
			}
		}
		if len(child) == 0 {
			continue
		}
		for i := 0; i < len(child); i++ {
			if len(child[i]) == 0 {
				for j := 0; j < len(father.Seq[i]); j++ {
					if usedMessage[father.Seq[i][j]] == 0 {
						child[i] = append(child[i], father.Seq[i][j])
						usedMessage[father.Seq[i][j]] = 1
					}
				}
				for j := 0; j < len(mother.Seq[i]); j++ {
					if usedMessage[mother.Seq[i][j]] == 0 {
						child[i] = append(child[i], mother.Seq[i][j])
						usedMessage[mother.Seq[i][j]] = 1
					}
				}
				for k := 0; k < gPara.NTask; k++ {
					if usedMessage[k] == 0 {
						ind := rand.Intn(len(child))
						child[ind] = append(child[ind], k)
						usedMessage[k] = 1
					}
				}
			}
		}
		childAsg := make([]int, len(father.Asg))
		for i := 0; i < len(father.Asg); i++ {
			valueA := gPara.CapRes[father.Asg[i]][3] / gPara.CapResCost[father.Asg[i]][1][3]
			valueB := gPara.CapRes[mother.Asg[i]][3] / gPara.CapResCost[mother.Asg[i]][1][3]
			if valueB < valueA {
				childAsg[i] = father.Asg[i]
			} else {
				childAsg[i] = mother.Asg[i]
			}
		}
		//新人入队
		if IsLegalTask(child, childLenth, childAsg) {
			*ethnicity = append(*ethnicity, Ethnicity{child, childAsg})
		}
	}
}
func ParentSwap(father, mother []int, child *[][]int, usedMessage *[]int) {
	if IsSeqAllUnused(father, *usedMessage) {
		*child = append(*child, CopySliceInt(father))
		for j := 0; j < len(father); j++ {
			(*usedMessage)[father[j]] = 1
		}
	} else if IsSeqAllUnused(mother, *usedMessage) {
		*child = append(*child, CopySliceInt(mother))
		for j := 0; j < len(mother); j++ {
			(*usedMessage)[mother[j]] = 1
		}
	} else {
		*child = append(*child, []int{})
	}
}
func IsSeqAllUnused(seq []int, usedMess []int) bool {
	for i := 0; i < len(seq); i++ {
		if usedMess[seq[i]] == 1 {
			return false
		}
	}
	return true
}
func Mutation(ethnicity *[]Ethnicity, rate, logBase float64, optEndTime int64, gPara *GPara) {
	num := int(float64(len(*ethnicity))*rate) + 1 //变异数量
	for num > 0 {
		nowTime := time.Now().Unix()
		if nowTime >= optEndTime {
			break
		}
		num--
		index := FirstProbability(*ethnicity, rand.Float64(), "worst", gPara) //表现差的个体变异概率大
		MutationCal(ethnicity, index, logBase, optEndTime)
	}
}
func MutationCal(ethnicity *[]Ethnicity, index int, logBase float64, optEndTime int64) { //变异操作函数
	person := (*ethnicity)[index]
	lenth := len(person.Seq)
	if lenth == 0 {
		return
	}
	loop := int(math.Log(rand.Float64()) / math.Log(logBase))
	for loop > 0 {
		nowTime := time.Now().Unix()
		if nowTime >= optEndTime {
			break
		}
		loop--
		FromIndex := rand.Intn(len(person.Seq))
		ToIndex := rand.Intn(len(person.Seq))
		FromSeq := person.Seq[FromIndex]
		ToSeq := person.Seq[ToIndex]
		tryNum := 50
		for (len(FromSeq) == 0 || len(ToSeq) == 0) && tryNum > 0 {
			tryNum--
			FromSeq = person.Seq[rand.Intn(len(person.Seq))]
			ToSeq = person.Seq[rand.Intn(len(person.Seq))]
		}
		if len(FromSeq) == 0 || len(ToSeq) == 0 {
			continue
		}
		startIndex := rand.Intn(len(FromSeq))
		endIndex := rand.Intn(len(FromSeq)-startIndex) + startIndex + 1
		insertIndex := rand.Intn(len(ToSeq) + 1)
		newFromeSeq := []int{}
		newToSeq := []int{}
		newFromeSeq = append(newFromeSeq, FromSeq[:startIndex]...)
		newFromeSeq = append(newFromeSeq, FromSeq[endIndex:]...)
		newToSeq = append(newToSeq, ToSeq[:insertIndex]...)
		newToSeq = append(newToSeq, FromSeq[startIndex:endIndex]...)
		newToSeq = append(newToSeq, ToSeq[insertIndex:]...)
		//新人入队
		newPerson := [][]int{}
		for i := 0; i < len(person.Seq); i++ {
			newPerson = append(newPerson, CopySliceInt(person.Seq[i]))
		}
		newPerson[FromIndex] = newFromeSeq
		newPerson[ToIndex] = newToSeq
		if IsLegalTask(newPerson, lenth, person.Asg) {
			*ethnicity = append(*ethnicity, Ethnicity{newPerson, person.Asg})
		}
	}
}
func FirstProbability(ethnicity []Ethnicity, randNum float64, mode string, gPara *GPara) int {
	var maxl float64 = 0.0
	var minl float64 = 1000000000000.0
	distanceSlice := []float64{}
	ProbabilitySlice := []float64{}
	for _, person := range ethnicity {
		var tmp float64 = 100000000.0
		if IsLegal(person.Seq, gPara, person.Asg) {
			tmp = GetDistance(person.Seq, person.Asg, gPara)
		}
		distanceSlice = append(distanceSlice, tmp)
		maxl = Maxfloat64(maxl, tmp)
		minl = Minfloat64(minl, tmp)
	}
	var total float64 = 0.0
	for i := 0; i < len(distanceSlice); i++ {
		if mode == "best" {
			va := -(distanceSlice[i] - minl) / (maxl - minl + 0.1)
			distanceSlice[i] = float64(math.Exp(float64(va)))
		} else {
			va := (distanceSlice[i]-minl)/(maxl-minl+0.1) - 1
			distanceSlice[i] = float64(math.Exp(float64(va)))
		}
		total += distanceSlice[i]
	}
	for i := 0; i < len(distanceSlice); i++ {
		ProbabilitySlice = append(ProbabilitySlice, distanceSlice[i]/total)
		if i > 0 {
			ProbabilitySlice[i] += ProbabilitySlice[i-1]
		}
	}
	for i := 0; i < len(ProbabilitySlice); i++ {
		if randNum <= ProbabilitySlice[i] {
			return i
		}
	}
	return len(ProbabilitySlice) - 1
}
func GetSeqDistance(seq []int, index int, asgmets []int, gPara *GPara) float64 {
	var distance float64 = 0.0
	if len(seq) == 0 {
		return 0.0
	}
	for j := 0; j < len(seq)-1; j++ {
		distance += gPara.Cost[asgmets[index]][0][gPara.TaskLoc[seq[j]]][gPara.TaskLoc[seq[j+1]]]
	}
	distance += gPara.Cost[asgmets[index]][0][gPara.TaskLoc[seq[len(seq)-1]]][0]
	distance += gPara.Cost[asgmets[index]][0][0][gPara.TaskLoc[seq[0]]]
	return distance
}
func GetSeqDuration(seq []int, index int, asgmets []int, gPara *GPara) float64 {
	// attention 无加stop内时间！！
	var duration float64 = 0.0
	if len(seq) == 0 {
		return 0.0
	}
	for j := 0; j < len(seq)-1; j++ {
		duration += gPara.Cost[asgmets[index]][1][gPara.TaskLoc[seq[j]]][gPara.TaskLoc[seq[j+1]]]
	}
	duration += gPara.Cost[asgmets[index]][1][gPara.TaskLoc[seq[len(seq)-1]]][0]
	duration += gPara.Cost[asgmets[index]][1][0][gPara.TaskLoc[seq[0]]]
	return duration
}
func GetDistance(seqs [][]int, asgmets []int, gPara *GPara) float64 { //距离函数
	var distance float64 = 0.0
	for i := 0; i < len(seqs); i++ {
		distance += GetSeqDistance(seqs[i], i, asgmets, gPara)
	}
	return distance
}
func GetDuration(seqs [][]int, asgmets []int, gPara *GPara) float64 { //时间函数
	var duration float64 = 0.0
	for i := 0; i < len(seqs); i++ {
		duration += GetSeqDuration(seqs[i], i, asgmets, gPara)
	}
	return duration
}
func SortPerson(seqs [][]int, asgmets []int, gPara *GPara) float64 {
	Lenth := 0
	for i := 0; i < len(seqs); i++ {
		Lenth += len(seqs[i])
	}
	if gPara.Obj == 2 {
		// todo !!!加mapcost
		return GetSeqsMapCost(seqs, asgmets, gPara) + 10000000*float64(gPara.NTask-Lenth)
	}
	return GetDistance(seqs, asgmets, gPara) + 10000000*float64(gPara.NTask-Lenth)
}
