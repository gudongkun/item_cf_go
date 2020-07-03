package main

import "github.com/gudongkun/item_cf_go/cf_lib"

func main()  {
	cf := cf_lib.GetItemCF()
	cf.DoEvaluate()
}
