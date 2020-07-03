package main

import "github.com/gudongkun/item_cf_go/cf_lib"

func main()  {
	cf := cf_lib.GetItemCF()
	cf.DoCalculate()
	//训练完成后直接显示评估结果
	cf.EvaluateData()
}

