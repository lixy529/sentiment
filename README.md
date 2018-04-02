sentiment
======

一个golang的编写的情感分析小程序，以学习为目的，仅供大家学习。

首先通过train进行训练，然后通过训练的结果进行对文本的是正负面概率的判断

install
-------

go get https://github.com/lixy529/sentiment


示例
-------
dictPath := "./dict"
mSentiment := NewSentiment(dictPath)
// 初始化
err := mSentiment.Init()
if err != nil {
    t.Errorf("init err: %s", err.Error())
    return
}

text := "这个公司要倒闭"

sentiType, scores := mSentiment.CalcSemti(text)
if sentiType == S_NEGATIVE {
    fmt.Println("===>负面<===")
} else if sentiType == S_POSITIVE {
    fmt.Println("===>正面<===")
} else {
    fmt.Println("===>中立<===")
}

// 反初始化
mSentiment.UnInit()