package cf_lib

import (
	"os"
	"testing"
)

var con_cf *ItemCf

func TestGetItemCF(t *testing.T) {
	t.Skip()
	cf := GetItemCF()
	if !IsExist(cf.DataPath) {
		t.Errorf("数据文件:%s 不存在", cf.DataPath)
	}
	if cf.DataPath == "" {
		t.Errorf("数据文件路径未设置")
	}
	if cf.SaveTestPath == "" {
		t.Errorf("测试数据集存储路径未设置")
	}
	if cf.SaveTrainPath == "" {
		t.Errorf("训练数据集存储路径未设置")
	}
	if cf.SaveSMXPath == "" {
		t.Errorf("相似度矩阵存储路径未设定")
	}
}

func TestItemCFTrain(t *testing.T) {
	t.Skip()
	con_cf = GetItemCF()
	t.Run("loadData", testLoadData)        //读取文件分成训练集合，和测试集合
	t.Run("calculate", testGenerateMarix)  //生成相似度矩阵
	t.Run("saveData", testSaveTrainedData) //生成相似度矩阵
}

//读取文件分成训练集合，和测试集合
func testLoadData(t *testing.T) {
	con_cf.LoadData()
	if con_cf.MaxUserId != 610 {
		t.Errorf("最大uid读取错误")
	}
	if len(con_cf.TestSet) == 0 {
		t.Errorf("测试数据集生成错误")
	}
	if len(con_cf.TrainSet) == 0 {
		t.Errorf("训练数据集生成错误")
	}
}

//生成相似度矩阵
func testGenerateMarix(t *testing.T) {
	con_cf.GenerateMarix()
	if len(con_cf.SaveSMXPath) == 0 {
		t.Errorf("相似度矩阵生成错误，无数据生成")
	}
}

// 持久化计算结果
func testSaveTrainedData(t *testing.T) {
	os.Remove(con_cf.SaveSMXPath)
	os.Remove(con_cf.SaveTestPath)
	os.Remove(con_cf.SaveTrainPath)
	con_cf.SaveTrainedData()
	if !IsExist(con_cf.SaveSMXPath) {
		t.Errorf("相似度矩阵文件:%s 未生成", con_cf.SaveSMXPath)
	}
	if !IsExist(con_cf.SaveTestPath) {
		t.Errorf("测试数据集文件:%s 未生成", con_cf.SaveSMXPath)
	}
	if !IsExist(con_cf.SaveSMXPath) {
		t.Errorf("训练数据集文件文件:%s 未生成", con_cf.SaveSMXPath)
	}

}

//测试训练效果
func TestItemCFTest(t *testing.T) {
	con_cf = GetItemCF()
	t.Run("loadTrainedData", testLoadTrainedData)
	t.Run("Recomened", testRecomened)
	t.Run("EvaluateData", testEvaluateData)
}

func testEvaluateData(t *testing.T) {
	con_cf.EvaluateData()
}

//测试推荐函数
func testRecomened(t *testing.T) {
	var recList []MapSortBean
	for uid, _ := range con_cf.TestSet {
		recList = con_cf.Recommend(uid)
		break
	}
	if len(recList) == 0 {
		t.Errorf("推荐函数未产生电影，推荐失败")
	}

}

//测试加载数据
func testLoadTrainedData(t *testing.T) {
	con_cf.LoadTrainedData() //加载训练数据
	if len(con_cf.SaveSMXPath) == 0 {
		t.Errorf("相似度矩加载失败")
	}
	if len(con_cf.TrainSetRec) == 0 {
		t.Errorf("训练集数据加载失败")
	}
	if len(con_cf.TestSet) == 0 {
		t.Errorf("测试集数据加载失败")
	}
}

//判断文件是否存在
func IsExist(f string) bool {
	_, err := os.Stat(f)
	return err == nil || os.IsExist(err)
}
