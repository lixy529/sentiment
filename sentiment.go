// 情感分析模型训练
//   变更历史
//     2018-03-24  lixiaoya  新建
package sentiment

import (
	"github.com/yanyiwu/gojieba"
	"github.com/jbrukh/bayesian"
	"path"
	"strings"
	"os"
	"io"
	"bufio"
	"strconv"
)

var (
	// 情感类型
	S_POSITIVE = 1 // 正面
	S_NEGATIVE = 2 // 反面
	S_NEUTRAL  = 3 // 中立

	// 文本正负面
	GOOD bayesian.Class = "Good"
	BAD  bayesian.Class = "Bad"

	MIN = 0.7 // 正负面比较值
)

// stWarn 预警级别
type stWarn struct {
	level int      // 预警级别
	words []string // 预警词库
}

// Sentiment 情感分析构体
type Sentiment struct {
	dictPath   string               // 字典目录
	jieba      *gojieba.Jieba       // 语义分析对象
	classifier *bayesian.Classifier // 情感分析对象
	sentiWords []string             // 情感词库
}

// NewSentiment 返回对象
//   参数
//     dictPath: 字典目录
//   返回
//     Sentiment对象
func NewSentiment(dictPath string) *Sentiment {
	s := &Sentiment{}
	s.dictPath = dictPath

	return s
}

// Init 初始化相关信息
//   参数
//     void
//   返回
//     错误信息
func (s *Sentiment) Init() error {
	// 读取情感词库
	var err error
	fullFile := path.Join(s.dictPath, "senti", "words_senti.dat")
	s.sentiWords, err = ReadFile(fullFile)
	if err != nil {
		return err
	}

	// 初始化情感分析对象
	if s.classifier == nil {
		s.classifier = bayesian.NewClassifier(GOOD, BAD)
		s.learn()
	}

	// 初始化分词对象
	if s.jieba == nil {
		dictFile := path.Join(s.dictPath, "jieba", "jieba.dict.utf8")
		hmmFile := path.Join(s.dictPath, "jieba", "hmm_model.utf8")
		userFile := path.Join(s.dictPath, "jieba", "user.dict.utf8")
		idfFile := path.Join(s.dictPath, "jieba", "idf.utf8")
		stopFile := path.Join(s.dictPath, "jieba", "stop_words.utf8")
		s.jieba = gojieba.NewJieba(dictFile, hmmFile, userFile, idfFile, stopFile)
	}

	return nil
}

// UnInit 反初始化相关信息
//   参数
//     void
//   返回
//     void
func (s *Sentiment) UnInit() {
	defer s.jieba.Free()
}

// learn 学习关键词
//   参数
//     void
//   返回
//     void
func (s *Sentiment) learn() {
	// 打开学习结果文件
	trainResFile := path.Join(s.dictPath, "senti", "train_res.dat")
	fi, err := os.Open(trainResFile)
	if err != nil {
		return
	} else if fi == nil {
		return
	}
	defer fi.Close()

	// 读取文件
	br := bufio.NewReader(fi)
	for {
		line, _, err := br.ReadLine()
		if err == io.EOF {
			break
		}

		trainText := string(line)
		if len(trainText) < 3 {
			continue
		}

		// 学习
		sentiType := trainText[0:1]
		tmp, _ := strconv.Atoi(sentiType)
		var which bayesian.Class
		if tmp == S_POSITIVE {
			which = GOOD
		} else if tmp == S_NEGATIVE {
			which = BAD
		}
		transRes := strings.Split(trainText[2:], "|")
		if len(transRes) != len(s.sentiWords) {
			return
		}
		words := []string{}
		for k, v := range transRes {
			if v == "1" {
				words = append(words, s.sentiWords[k])
			}
		}
		s.classifier.Learn(words, which)
	}

	return
}

// replace 替换符号
// 将文本里的全角符号转成半角符号
//   参数
//     text: 文本内容
//   返回
//     返回替换后的文本
func (s *Sentiment) replace(text *string) string {
	// 全角符号替换为半角
	symList := map[string]string{
		"，": ",",
		"。": ".",
		"？": "?",
		"！": "!",
		"：": ":",
	}
	sepLen := 3
	textLen := len(*text)
	start := 0
	tmpText := ""
	for i := 0; i+sepLen <= textLen; i++ {
		tmp := (*text)[i:i+sepLen]
		if sym, ok := symList[tmp]; ok {
			tmpText += (*text)[start:i] + sym
			start = i + sepLen
			i += sepLen - 1
		}
	}
	if (*text)[start:] != "" {
		tmpText += (*text)[start:]
	}

	return tmpText
}

// CalcSemti 计算文本情感值和预警级别
//   参数
//     text: 文本内容
//   返回
//     情感类型 1-正面 2-负面 3-中立
//     正、负面得分
func (s *Sentiment) CalcSemti(text string) (int, []float64) {
	if text == "" {
		return S_NEUTRAL, []float64{0, 0}
	}

	text = s.replace(&text)
	dataSet := s.jieba.Cut(text, true)
	scores, _, _ := s.classifier.LogScores(dataSet) // 计算情感得分
	diffScore := scores[0] - scores[1]

	if diffScore > MIN {
		return S_POSITIVE, scores
	} else if -diffScore > MIN {
		return S_NEGATIVE, scores
	}

	return S_NEUTRAL, scores
}
