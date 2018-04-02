package sentiment

import (
	"testing"
	"fmt"
)

// TestSentiment 感情分析测试
func TestSentiment(t *testing.T) {
	dictPath := "./dict"
	mSentiment := NewSentiment(dictPath)
	// 初始化
	err := mSentiment.Init()
	if err != nil {
		t.Errorf("init err: %s", err.Error())
		return
	}

	text := "草泥马你就是个王八蛋，混账玩意!你们的手机真不好用！非常生气，我非常郁闷！！！！"
	//text := "我好开心啊，非常非常非常高兴！今天我得了一百分，我很兴奋开心，非常愉快，开心"
	//text := "公司马上要倒闭了！"
	//text := "这个手机太垃圾了！"
	//text := "今天公司股票涨停啦！！！"

	sentiType, scores := mSentiment.CalcSemti(text)
	fmt.Println("scores:", scores)
	if sentiType == S_NEGATIVE {
		fmt.Println("===>负面<===")
	} else if sentiType == S_POSITIVE {
		fmt.Println("===>正面<===")
	} else {
		fmt.Println("===>中立<===")
	}

	// 反初始化
	mSentiment.UnInit()
}
