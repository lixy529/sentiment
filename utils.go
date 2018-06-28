// 工具函数
//   变更历史
//     2018-03-23  lixiaoya  新建
package sentiment

import (
	"strings"
	"regexp"
	"github.com/djimenez/iconv-go"
	"os"
	"io"
	"bufio"
	"io/ioutil"
	"crypto/tls"
	"crypto/x509"
	"time"
	"net/http"
	"net/url"
)

const (
	OPT_PROXY = iota
	OPT_HTTPHEADER
	OPT_SSLCERT
)
// InSlice 保存字符串在切片里是否存在
//   参数
//     str: 要匹配的字符串
//     arr: 切片
//     i:   是否忽略大小写，默认为 true
//   返回
//     true-存在 false-不存在
func InStrSlice(str string, arr []string, i ...bool) bool {
	i = append(i, true)
	if i[0] {
		str = strings.ToLower(str)
	}
	for _, v := range arr {
		if i[0] {
			v = strings.ToLower(v)
		}
		if str == v {
			return true
		}
	}

	return false
}

// Html2Text html格式转文本
//   参数
//     strHtml:  html格式串
//   返回
//     文本、错误信息
func Html2Text(strHtml *string) (string, error) {
	// 判断页面是否是gb2312
	isGb2312 := false
	src := *strHtml
	re, _ := regexp.Compile(`(?i)<meta[\S\s]+?charset=["|']?gB2312[\S\s]*?>`)
	if re.MatchString(*strHtml) {
		isGb2312 = true
	}

	// 删除<head>xxxx</head>
	re, _ = regexp.Compile(`(?i)<head[\S\s]+?</head>`)
	src = re.ReplaceAllString(*strHtml, "")

	// 删除<script>xxxx</script>
	re, _ = regexp.Compile(`(?i)<script[\S\s]+?</script>`)
	src = re.ReplaceAllString(src, "")

	// 删除<style>xxxx</style>
	re, _ = regexp.Compile(`(?i)<style[\S\s]+?</style>`)
	src = re.ReplaceAllString(src, "")

	// 删除<a>xxxx</a>
	re, _ = regexp.Compile(`(?i)<a[\S\s]+?</a>`)
	src = re.ReplaceAllString(src, "")

	// 删除<xxx></xxx>
	re, _ = regexp.Compile(`(?i)<[\S\s]+?>`)
	src = re.ReplaceAllString(src, "")

	// 删除多个换行符
	re, _ = regexp.Compile(`\s{2,}`)
	src = re.ReplaceAllString(src, "\n")

	// 是否需要转码
	if isGb2312 {
		out := make([]byte, len([]byte(src))*2)
		_, _, err := iconv.Convert([]byte(src), out, "gb2312", "utf-8")
		if err != nil {
			return "", err
		}
		src = string(out)
	}

	return src, nil
}

// SelStrVal 根据条件返回相应选项
//   参数
//     con:  条件
//     opt1: 选项1
//     opt2: 选项2
//   返回
//     如果条件为true，返回选项1，否则返回选项2
func SelStrVal(con bool, opt1, opt2 string) string {
	if con {
		return opt1
	}

	return opt2
}

// SelIntVal 根据条件返回相应选项
//   参数
//     con:  条件
//     opt1: 选项1
//     opt2: 选项2
//   返回
//     如果条件为true，返回选项1，否则返回选项2
func SelIntVal(con bool, opt1, opt2 int) int {
	if con {
		return opt1
	}

	return opt2
}

// ReadFile 按行读取文件
//   参数
//      fullFile: 文件全路径
//   返回
//     文件内容、错误信息
func ReadFile(fullFile string) ([]string, error) {
	words := []string{}

	// 打开文件
	fi, err := os.Open(fullFile)
	if err != nil {
		return words, err
	}
	defer fi.Close()

	// 读取文件
	br := bufio.NewReader(fi)
	for {
		line, _, err := br.ReadLine()
		if err == io.EOF {
			break
		}

		if string(line) != "" {
			words = append(words, string(line))
		}
	}

	return words, nil
}

// Curl Get或Post数据
//   参数
//     url:     访问的url地址
//     data:    提交的数据，json或http query格式
//     method:  方法标识，取值： POST | GET，默认为GET
//     timeout: 超时时间，单位秒，默认为5秒
//     params:  其它参数
//       目前支持如下参数：
//         OPT_PROXY：     代理，如:http://10.12.34.53:2443
//         OPT_SSLCERT:    https证书，传map[string]string型，certFile（cert证书）、keyFile（key证书，为空时使用cert证书）、caFile（根ca证书，可为空）
//         OPT_HTTPHEADER: http请求头，传map[string]string型
//   返回
//     结果串、http状态、错误内容
func Curl(urlAddr, data, method string, timeout time.Duration, params ...map[int]interface{}) (string, int, error) {
	if timeout <= 0 {
		timeout = 5
	}

	if strings.ToUpper(method) == "POST" {
		method = "POST"
	} else {
		method = "GET"
	}

	// 设置Content-Type
	headers := make(map[string]string)
	if data != "" && data[0] == '{' { // json: {"a":"1", "b":"2", "c":"3"}
		headers["Content-Type"] = "application/json; charset=utf-8"
	} else { // http query: a=1&b=2&c=3
		headers["Content-Type"] = "application/x-www-form-urlencoded"
	}

	tpFlag := false
	tp := http.Transport{}
	if len(params) > 0 {
		param := params[0]

		// 设置代理
		if v, ok := param[OPT_PROXY]; ok {
			if proxyAddr, ok := v.(string); ok {
				proxy := func(_ *http.Request) (*url.URL, error) {
					return url.Parse(proxyAddr)
				}
				tp.Proxy = proxy
				tpFlag = true
			}
		}

		// 设置证书
		if v, ok := param[OPT_SSLCERT]; ok {
			if t, ok := v.(map[string]string); ok {
				if certFile, ok := t["certFile"]; ok && certFile != "" {
					keyFile := ""
					if keyFile, ok = t["keyFile"]; !ok || keyFile == "" {
						keyFile = certFile
					}
					caFile, _ := t["caFile"]

					tlsCfg, err := parseTLSConfig(certFile, keyFile, caFile)
					if err == nil {
						tp.TLSClientConfig = tlsCfg
						tpFlag = true
					}
				}
			}
		}

		// 设置HEADER
		if v, ok := param[OPT_HTTPHEADER]; ok {
			if t, ok := v.(map[string]string); ok {
				for key,val := range t {
					headers[key] = val
				}
			}
		}
	}

	req, err := http.NewRequest(method, urlAddr, strings.NewReader(data))
	if err != nil {
		return "", -1, err
	}

	// 设置HEADER
	for key, val := range headers {
		if strings.ToLower(key) == "host" {
			req.Host = val
		} else {
			req.Header.Set(key, val)
		}
	}

	client := &http.Client{
		Timeout: timeout * time.Second,
	}
	// 设置Transport
	if tpFlag {
		client.Transport = &tp
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", -1, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", -1, err
	}

	return string(body), resp.StatusCode, nil
}

// parseTLSConfig 解析证书文件
//   参数
//     certFile: cert证书
//     keyFile:  key证书
//     caFile:   根ca证书，可为空
//   返回
//     解析结果、错误信息
func parseTLSConfig(certFile, keyFile, caFile string) (*tls.Config, error) {
	// load cert
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	tlsCfg := tls.Config{
		InsecureSkipVerify: true,
		Certificates: []tls.Certificate{cert},
	}

	// load root ca
	if caFile != "" {
		caData, err := ioutil.ReadFile(caFile)
		if err != nil {
			return nil, err
		}
		pool := x509.NewCertPool()
		pool.AppendCertsFromPEM(caData)
		tlsCfg.RootCAs = pool
	}

	return &tlsCfg, nil
}

