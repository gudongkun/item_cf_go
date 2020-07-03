package cf_lib

import (
	"bytes"
	"encoding/csv"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

//基于项目的余弦相似度计算类
type ItemCf struct {
	//计算属性
	MaxUserId      int                           //最大用户数决定向量长度
	TrainSet       map[string][]float64          //训练集合
	TrainSetRec    map[string]map[string]float64 //训练集合
	TestSet        map[string]map[string]float64 //测试集合
	SimilierMatrix map[string]map[string]float64 //相似度矩阵
	TrainNum       int //训练集总数据条数
	TestNum        int //训练集总数据条数
	//配置属性
	DataPath       string  // 数据文件路径
	SaveTestPath   string  // 测试集存储路径
	SaveTrainPath  string  // 训练集存储路径
	SaveSMXPath    string  //	相似度矩阵存储路径
	TrainSetPecent float64 // 训练集数量占比
	SimMovieNum    int     //产生相似电影数
	RecMovieNum    int     //产生推荐电影数
}

//获取计算对象，初始化配置
func GetItemCF() *ItemCf {
	cf := ItemCf{
		DataPath:       "../ml-1m/ratings.csv",
		SaveTestPath:   "../runtime/testset.data",
		SaveTrainPath:  "../runtime/trainset.data",
		SaveSMXPath:    "../runtime/itemSimliarMatrix.data",
		TrainSetPecent: 0.75,
		TrainSet:       make(map[string][]float64),
		TrainSetRec:    make(map[string]map[string]float64),
		TestSet:        make(map[string]map[string]float64),
		SimilierMatrix: make(map[string]map[string]float64),
		SimMovieNum:    20,
		RecMovieNum:    10,
		TrainNum:       0,
		TestNum:        0,
	}
	fmt.Println("配置初始化完成")
	return &cf
}

//--业务函数
//计算数据
func (cf *ItemCf) DoCalculate() {
	cf.LoadData()        //读取文件分成训练集合，和测试集合
	cf.GenerateMarix()   //生成相似度矩阵
	cf.SaveTrainedData() //持久化数据
}

func (cf *ItemCf) DoEvaluate() {
	cf.LoadTrainedData()
	cf.EvaluateData()
}

//读取文件分成训练集合，和测试集合
func (cf *ItemCf) LoadData() {
	cf.InitMaxUid()
	cf.InitDivedSet()

	cf.InitCalculateSet()

	cf.AdjustCalculateSet() //修正余弦相似度
}
//取得最大用户id
func  (cf *ItemCf) InitMaxUid()  {
	fs, err := os.Open(cf.DataPath)
	if err != nil {
		log.Fatalf("无法打开数据文件: %+v", err)
	}
	defer fs.Close()
	r := csv.NewReader(fs)
	//取得最大用户数，作为向量维度
	cf.MaxUserId = 0
	for {
		row, err := r.Read()
		if err != nil && err != io.EOF {
			log.Fatalf("文件读取错误 ： %+v", err)
		}
		if err == io.EOF {
			break
		}
		uid, _ := strconv.Atoi(row[0])
		if cf.MaxUserId < uid {
			cf.MaxUserId = uid
		}
	}
}
//生成测试数据集和训练数据集
func (cf *ItemCf)InitDivedSet()  {
	fs, err := os.Open(cf.DataPath)
	if err != nil {
		log.Fatalf("无法打开数据文件: %+v", err)
	}
	defer fs.Close()
	r := csv.NewReader(fs)
	r.Read()
	rand.Seed(time.Now().UnixNano())
	for {
		row, err := r.Read()
		if err != nil && err != io.EOF {
			log.Fatalf("文件读取错误 ： %+v", err)
		}
		if err == io.EOF {
			break
		}
		score, _ := strconv.ParseFloat(row[2], 64)
		if rand.Float64() <= cf.TrainSetPecent {
			cf.TestNum++
			if _, ok := cf.TrainSetRec[row[0]]; !ok {
				cf.TrainSetRec[row[0]] = make(map[string]float64)
			}
			cf.TrainSetRec[row[0]][row[1]] = float64(score)
		} else {
			cf.TrainNum++
			if _, ok := cf.TestSet[row[0]]; !ok {
				cf.TestSet[row[0]] = make(map[string]float64)
			}
			cf.TestSet[row[0]][row[1]] = float64(score)
		}
	}
	fmt.Printf("数据分组完成，训练集包含数据%d条,测试集包含数据%d条 \n", cf.TrainNum, cf.TestNum)
}



func (cf *ItemCf)InitCalculateSet()  {
	for uid, movieMap :=range cf.TrainSetRec{
		for movieId ,rating := range movieMap {
			if _, ok := cf.TrainSet[movieId]; !ok {
				cf.TrainSet[movieId] = make([]float64, cf.MaxUserId, cf.MaxUserId)
			}
			uidInt, _ := strconv.Atoi(uid)
			//余弦相似度
			cf.TrainSet[movieId][uidInt-1] = rating
		}
	}
}

//相似度仅考虑向量维度方向上的相似而没考虑到各个维度的量纲的差异性，所以在计算相似度的时候，做了每个维度减去均值的修正操作
func (cf *ItemCf)AdjustCalculateSet()  {
	for movieId, userMap :=range cf.TrainSet{
		totalRating := 0.0
		for _,rating := range userMap{
			totalRating += rating
		}
		AverageRating := totalRating/float64(len(userMap))
		for UserId,rating:= range userMap{
			if rating != 0 {
				cf.TrainSet[movieId][UserId] = rating - AverageRating
			}
		}
	}
	
}

//生成相似度矩阵
func (cf *ItemCf) GenerateMarix() {
	var listLine sync.Map

	var wg sync.WaitGroup
	wg.Add(len(cf.TrainSet))

	i := 0

	fmt.Printf("相似度矩阵计算开始共需计算%d条,已完成:\n", len(cf.TrainSet))
	for k, v := range cf.TrainSet {
		go func(k1 string, v1 []float64) {
			for k2, v2 := range cf.TrainSet {
				if k1 == k2 {
					continue
				}
				simVal := cf.Cosine(v1, v2)
				if simVal != 0 {
					key := fmt.Sprintf("%v-%v", k1, k2)
					listLine.Store(key, simVal)
				}
			}
			i++
			fmt.Printf("\r%d               ", i)
			wg.Done()
		}(k, v)
	}
	wg.Wait()
	totalNum := 0
	listLine.Range(func(k, v interface{}) bool {
		totalNum++
		keyArr := strings.Split(k.(string), "-")
		if _, ok := cf.SimilierMatrix[keyArr[0]]; !ok {
			cf.SimilierMatrix[keyArr[0]] = make(map[string]float64)
		}
		cf.SimilierMatrix[keyArr[0]][keyArr[1]] = v.(float64)
		return true
	})

	fmt.Printf("\n相似度矩阵计算完成，共计算电影：%d部，共产生数据%d条\n", len(cf.SimilierMatrix), totalNum)
}

//保存相似度矩阵，训练集合和测试集合到本地文件
func (cf *ItemCf) SaveTrainedData() {
	cf.SaveGob(cf.SaveSMXPath, cf.SimilierMatrix)
	cf.SaveGob(cf.SaveTestPath, cf.TestSet)
	cf.SaveGob(cf.SaveTrainPath, cf.TrainSetRec)

	//cf.SaveJson(cf.SaveSMXPath,cf.SimilierMatrix)
	//cf.SaveJson(cf.SaveTestPath,cf.TestSet)
	//cf.SaveJson(cf.SaveTrainPath,cf.TrainSet)
}

//加载训练数据集
func (cf *ItemCf) LoadTrainedData() {
	//加载相似度矩阵
	smxBytes, _ := ioutil.ReadFile(cf.SaveSMXPath)
	dec := gob.NewDecoder(bytes.NewBuffer(smxBytes))
	dec.Decode(&cf.SimilierMatrix)
	//加载训练集合
	trainSetBytes, _ := ioutil.ReadFile(cf.SaveTrainPath)
	dec = gob.NewDecoder(bytes.NewBuffer(trainSetBytes))
	dec.Decode(&cf.TrainSetRec)
	//加载测试集合
	testSetBytes, _ := ioutil.ReadFile(cf.SaveTestPath)
	dec = gob.NewDecoder(bytes.NewBuffer(testSetBytes))
	dec.Decode(&cf.TestSet)

	//加载训练电影数据总数
	cf.TrainNum = 0
	for _, v := range cf.TrainSetRec {
		for range v {
			cf.TrainNum++
		}
	}

	fmt.Println("数据加载完毕")
}
// 产生推荐电影的函数
func (cf *ItemCf) Recommend(uid string) []MapSortBean {
	watchedMovices := cf.TrainSetRec[uid] //用户已经看过的电影
	RecMap := map[string]float64{}
	for moviceId, _ := range watchedMovices {
		relatedMovies := cf.sortMap(cf.SimilierMatrix[moviceId], cf.SimMovieNum)
		for _, movie := range relatedMovies {
			if _,ok := watchedMovices[movie.Key];ok{
				continue
			}
			if _, ok := RecMap[movie.Key]; !ok {
				RecMap[movie.Key] = movie.Val
			} else {
				RecMap[movie.Key] += movie.Val
			}
		}
	}
	return cf.sortMap(RecMap, cf.RecMovieNum)
}

//测试预测准确率
func (cf *ItemCf) EvaluateData() {
	fmt.Println("评估测试开始...")
	// 准确率和召回率
	hit := 0
	recNum := 0
	testNum := 0
	//覆盖率
	var allRecMovies []MapSortBean
	for uid, watchedMovies := range cf.TestSet {
		recList := cf.Recommend(uid)
		for _, movie := range recList {
			if _, ok := watchedMovies[movie.Key]; ok {
				hit += 1
			}
			allRecMovies = append(allRecMovies, movie)
		}
		recNum += cf.RecMovieNum
		testNum += len(watchedMovies)
	}
	precision := float64(hit) / float64(recNum)
	recall := float64(hit) / float64(testNum)
	coverage := float64(len(allRecMovies)) / float64(cf.TrainNum)

	fmt.Printf("准确率为:%v, 回复率为: %v , 覆盖率为: %v\n", precision, recall, coverage)
}

//----底层函数--------
//排序用数据结构
type MapSortBean struct {
	Key string
	Val float64
}

//map排序
func (cf *ItemCf) sortMap(mp map[string]float64, num int) []MapSortBean {
	var sortList []MapSortBean
	for k, v := range mp {
		sortList = append(sortList, MapSortBean{k, v})
	}
	sort.Slice(sortList, func(i, j int) bool {
		return sortList[i].Val > sortList[j].Val // 降序
	})
	if len(sortList) > num {
		return sortList[:num]
	} else {
		return sortList
	}
}

//存储数据到本地gob格式
func (cf *ItemCf) SaveGob(filename string, data interface{}) {
	os.Remove(filename)
	//创建缓存
	buf := new(bytes.Buffer)
	//把指针丢进去
	enc := gob.NewEncoder(buf)
	enc.Encode(data)
	//写入文件
	ioutil.WriteFile(filename, buf.Bytes(), 0777)
}

//存储数据到本地json格式
func (cf *ItemCf) SaveJson(filename string, data interface{}) {
	os.Remove(filename)
	cf.CreateDateDir(path.Dir(filename))
	dataString, _ := json.Marshal(&data)
	f, _ := os.Create(filename)           //创建文件
	io.WriteString(f, string(dataString)) //写入文件(字符串)
}

//创建文件夹
func (cf *ItemCf) CreateDateDir(dirPath string) {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		// 必须分成两步
		// 先创建文件夹
		os.Mkdir(dirPath, 0777)
		// 再修改权限
		os.Chmod(dirPath, 0777)
	}
}

//余弦相似度算法
func (cf *ItemCf) Cosine(a []float64, b []float64) float64 {
	sa := float64(0)
	sb := float64(0)
	s := float64(0)
	for i, _ := range a {
		s += a[i] * b[i]
	}
	if s == 0 {
		return 0
	}
	for i, _ := range a {
		sa += math.Pow(a[i], 2)
		sb += math.Pow(b[i], 2)
		sa += math.Pow(a[i], 2)
		sb += math.Pow(b[i], 2)
	}
	return s / (math.Sqrt(sa) * math.Sqrt(sb))
}

