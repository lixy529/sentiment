// 工具函数测试
//   变更历史
//     2018-03-23  lixiaoya  新建
package sentiment

import (
	"testing"
	"fmt"
)

// TestHtml2Text Html2Text函数测试
func TestHtml2Text(t *testing.T) {
	strHtml := `
<html>
<head>
  <title>This is a test</title>
  <script>aaa.js</script>
</head>
<body>
  <script>bbb.js</script>
  <div>div test 001</div>
  <div>div test 002</div>
  <div><p>pppp</p?</div>
  <br/>
  <br />
  <a href="http://xxx.com">aaa</a>
  <img src="http://xxx.com" /">
</body
</html>`
	str, err := Html2Text(&strHtml)
	fmt.Println(str, err)
}

// TestSelStrVal SelStrVal函数测试
func TestSelStrVal(t *testing.T) {
	opt1 := "aaa"
	opt2 := "bbb"
	opt := SelStrVal(true, opt1, opt2)
	if opt != opt1 {
		t.Errorf("SelStrVal err, Got:%s expected:%s", opt, opt1)
	}

	opt = SelStrVal(false, opt1, opt2)
	if opt != opt2 {
		t.Errorf("SelStrVal err, Got:%s expected:%s", opt, opt2)
	}
}

// TestSelIntVal SelIntVal函数测试
func TestSelIntVal(t *testing.T) {
	opt1 := 111
	opt2 := 222
	opt := SelIntVal(true, opt1, opt2)
	if opt != opt1 {
		t.Errorf("SelIntVal err, Got:%d expected:%d", opt, opt1)
	}

	opt = SelIntVal(false, opt1, opt2)
	if opt != opt2 {
		t.Errorf("SelIntVal err, Got:%d expected:%d", opt, opt2)
	}
}
