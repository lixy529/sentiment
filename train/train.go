// 感情分析模型训练
//   变更历史
//     2018-03-24  lixiaoya  新建
package main

import (
	"github.com/yanyiwu/gojieba"
	"github.com/lixy529/sentiment"
	"os"
	"io"
	"log"
	"bufio"
	"flag"
	"path"
	"strconv"
	"strings"
	"errors"
	"fmt"
)

// stTriain 训练结构体
type stTrain struct {
	dictPath string         // 字典目录
	jieba    *gojieba.Jieba // 语义分析对象
}

// NewStTrain 返回对象
//   参数
//     void
//   返回
//     stTrain对象
func NewStTrain(dictPath string) *stTrain {
	t := &stTrain{}
	t.dictPath = dictPath

	return t
}

// 读取字典文件内容
func (s *stTrain) init() error {
	return nil
}

// train 训练
//   参数
//     isNew: 是否新生成结果文件，true-是 false-否
//   返回
//     错误信息
func (t *stTrain) train(isNew bool) error {
	wordsFile := path.Join(t.dictPath, "senti", "words_senti.dat")
	trainFile := path.Join(t.dictPath, "senti", "train.dat")
	trainResFile := path.Join(t.dictPath, "senti", "train_res.dat")
	log.Println("======train begin======")
	log.Println("wordsFile:", wordsFile)
	log.Println("trainFile:", trainFile)
	log.Println("trainResFile:", trainResFile)

	// 读取关键词文件
	words, err := sentiment.ReadFile(wordsFile)
	if err != nil {
		log.Fatalf("ReadFile %s err: %s", wordsFile, err.Error())
		return err
	}

	// 初始化分析对象
	if t.jieba == nil {
		dictFile := path.Join(t.dictPath, "jieba", "jieba.dict.utf8")
		hmmFile := path.Join(t.dictPath, "jieba", "hmm_model.utf8")
		userFile := path.Join(t.dictPath, "jieba", "user.dict.utf8")
		idfFile := path.Join(t.dictPath, "jieba", "idf.utf8")
		stopFile := path.Join(t.dictPath, "jieba", "stop_words.utf8")
		t.jieba = gojieba.NewJieba(dictFile, hmmFile, userFile, idfFile, stopFile)
	}
	defer t.jieba.Free()

	// 打开结果文件
	var resFd *os.File
	openFlag := os.O_RDWR | os.O_CREATE
	if !isNew {
		// 追加
		log.Println("train append file")
		openFlag |= os.O_APPEND
	} else {
		os.Remove(trainResFile)
		log.Println("train new file")
	}
	resFd, err = os.OpenFile(trainResFile, openFlag, 0644)
	if err != nil {
		return err
	} else if resFd == nil {
		return errors.New("train result file is nul")
	}
	defer resFd.Close()

	// 待训练的数据文件
	fi, err := os.Open(trainFile)
	if err != nil {
		log.Fatalf("Open %s err: %s", trainFile, err.Error())
		return err
	}
	defer fi.Close()

	// 读取文件
	br := bufio.NewReader(fi)
	for {
		line, _, err := br.ReadLine()
		if err == io.EOF {
			break
		}

		if string(line) == "" {
			continue
		}

		// 写训练数据
		trainText := t.trainOne(string(line), words)
		resFd.WriteString(trainText + "\n")
	}

	log.Println("======train end======")

	return nil
}

// trainOne 训练一条数据
//   参数
//     trainText: 一条要训练的数据
//     words:     关键词
//   返回
//     错误信息
func (s *stTrain) trainOne(trainText string, words []string) string {
	if len(trainText) < 5 || len(words) == 0 {
		return ""
	}

	// 获取感情类型
	sentiType := trainText[0:1]
	tmp, _ := strconv.Atoi(sentiType)
	if tmp != sentiment.S_NEGATIVE && tmp != sentiment.S_POSITIVE {
		return ""
	}
	trainType := trainText[2:3]
	//fmt.Println("trainType:", trainType)
	trainText = trainText[4:]
	if trainType == "u" {
		urlAddr := trainText
		res, code, err := sentiment.Curl(urlAddr, "", "GET", 3)
		if err != nil {
			log.Fatalf("curl fail, url:%s code:%s, err:%s", urlAddr, code, err.Error())
			return ""
		} else if code >= 400 {
			log.Fatalf("curl fail, code err, url:%s code:%s", urlAddr, code)
			return ""
		}

		// 提取文件
		trainText, err = sentiment.Html2Text(&res)
		if err != nil {
			log.Fatalf("html to text fail, url:%s err:%s", urlAddr, err.Error())
			return ""
		}
	}
	fmt.Println("trainText:", trainText)

	// 分词处理
	if s.jieba == nil {
		return ""
	}
	textSet := s.jieba.Cut(trainText, true)
	trainData := []string{}
	for _, w := range words {
		if sentiment.InStrSlice(w, textSet) {
			trainData = append(trainData, "1")
		} else {
			trainData = append(trainData, "0")
		}
	}

	return sentiType + ":" + strings.Join(trainData, "|")
}

// main 训练入口
// 训练生成模型
// 执行命令:
// ./train -dir=关键词目录 -new=是否新生成结果文件[true|false]
func main() {
	dictPath := flag.String("dir", "../dict", "字典文件目录")
	isNew := flag.Bool("new", true, "是否新生成结果文件")
	flag.Parse()
	t := NewStTrain(*dictPath)
	err := t.train(*isNew)
	if err != nil {
		log.Fatalf("train err: %s", err.Error())
	}

	return
}
