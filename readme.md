# 推荐系统基于项目

> 实现一个电影推荐系统，采用协同过滤算法，，相似度算法为余弦相似度，基于用户和基于项目中选择基于项目数据集为movielens数据集

## 一、项目说明

项目名称：item_cf_go

语言：golang

项目地址：github.com/gudongkun/item_cf_go

目录结构：
1. calculate 计算相似度入口
2. cf_lib    业务主逻辑类
3. evaluete  计算后，不想计算只想再显示一次本次的准确率等信息可以执行此程序
4. runtime   运行calculate时自动生成，保存计算结果

## 二、如何使用
1.执行相似度计算计算-主要方法
```shell
# cd  {项目目录}/calculate
# go run main.go
```
![](.\doc\calculate_res.png)
2.重新显示测试信息
```shell
# cd {项目目录}/evaluete
# go run main.go
```
![](.\doc\evaluate_res.png)
### 三、版本更新记录
1. tag: v1 使用余弦相似度算，准确率保持在约 26.95%
### 四、项目参与人
gudongkun

952142073@qq.com

如有疑问可邮件联系
