train
======

感情分析的训练代码

```
# 编译
go build train.go

# 运行
./trian -dir=/xx/xx -new=true
dir: 字典目录的根目录，默认为../dict
new: 是否新生成训练结果文件，true-会新生成结果文件 false-追加到结果文件后面，默认为true

{$dir}/senti/words_senti.dat : 关键词
{$dir}/senti/train.dat       : 待训练数据
{$dir}/senti/train_res.dat   : 训练的结果文件
```
