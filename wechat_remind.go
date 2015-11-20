package utils

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

var is_login = false
var cookieStr string

type WebWeChat struct {
	token   string
	cookies []*http.Cookie
}

type Msg struct {
	Base      Base_resp `json:"base"`
	Redirect_url string `json:"redirect_url"`
}

type Base_resp struct {
	Ret int `json:"ret"`
	Err_msg string `json:"err_msg"`
}


func NewWebWeChat() *WebWeChat {
	w := new(WebWeChat)
	return w
}

const (
	LOGIN_URL = "https://mp.weixin.qq.com/cgi-bin/login?lang=zh_CN"
	EMAIL = "**@qq.com"
	PASSWORD = "**"
	
	//头文件
	REFERER_H = "Referer";
	HOST = "https://mp.weixin.qq.com/"
	
	USER_AGENT_H = "User-Agent";
	USER_AGENT = "Mozilla/5.0 (Windows NT 6.2; WOW64) AppleWebKit/537.22 (KHTML, like Gecko) Chrome/25.0.1364.172 Safari/537.22";
	
	//向某一个人发送消息，48小时之内必须有互动,不然不能主动发送消息
	SEND_URL = "https://mp.weixin.qq.com/cgi-bin/singlesend"
	REFER_URL = "https://mp.weixin.qq.com/cgi-bin/singlesendpage?t=message/send&action=index&tofakeid=%s&token=%s&lang=zh_CN"
	

)

//登入
func (w *WebWeChat) login() bool {
	h := md5.New()
	h.Write([]byte(PASSWORD))
	password := hex.EncodeToString(h.Sum(nil))
	//请求的key-value对
	post_arg := url.Values{"username": {EMAIL}, "pwd": {password}, "imagecode": {""}, "f": {"json"}}

	fmt.Println(strings.NewReader(post_arg.Encode()))
	
	req, err := http.NewRequest("POST", LOGIN_URL, strings.NewReader(post_arg.Encode()))
	
	req.Header.Set(REFERER_H, HOST)
	req.Header.Set(USER_AGENT_H,USER_AGENT)

	if err != nil {
		return is_login
		log.Fatal(err)
	}

	client := new(http.Client)
	resp, _ := client.Do(req)
	data, _ := ioutil.ReadAll(resp.Body)

	s := string(data)
	
	decode := json.NewDecoder(strings.NewReader(s))

	var m Msg
	
	if  err := decode.Decode(&m);err!=nil {
		return is_login
		log.Fatal("data decode error ",err.Error())
	}else {	
	    is_login = true
		if m.Base.Ret == 0 || m.Base.Ret == 65201 || m.Base.Ret == 65202 {
		var url_correct = false 
		var token string
		var url string
		var url_strings = strings.Split(m.Redirect_url, "?")
		if len(url_strings) == 2 {
		   url = url_strings[1]
		   urls := strings.Split(url,"&")
		   if len(urls) ==3 {
		    	token = strings.Split(urls[2],"=")[1]
				url_correct = true 
		}
		}else if len(url_strings) == 1{
			url = url_strings[0]
			urls := strings.Split(url,"&")
			 if len(urls) ==3 {
		    	token = strings.Split(urls[2],"=")[1]
				url_correct = true 
		}
		}

        w.token = token
		w.cookies = resp.Cookies()
		
		return is_login&&url_correct
}else {
	switch m.Base.Ret {
	case -1:
		fmt.Println("系统错误，请稍候再试。")
	case -2:
		fmt.Println("帐号或密码错误。")
	case -3:
		fmt.Println("您输入的帐号或者密码不正确，请重新输入。")
	case -4:
		fmt.Println("不存在该帐户。")
	case -5:
		fmt.Println("您目前处于访问受限状态。")
	case -6:
		fmt.Println("请输入图中的验证码")
	case -7:
		fmt.Println("此帐号已绑定私人微信号，不可用于公众平台登录。")
	case -8:
		fmt.Println("邮箱已存在。")
	case -32:
		fmt.Println("您输入的验证码不正确，请重新输入。")
	case -200:
		fmt.Println("因频繁提交虚假资料，该帐号被拒绝登录。")
	case -94:
		fmt.Println("请使用邮箱登陆。")
	case 10:
		fmt.Println("该公众会议号已经过期，无法再登录使用。")
	case -100:
		fmt.Println("海外帐号请在公众平台海外版登录,<a href=\"http://admin.wechat.com/\">点击登录</a>")
	default:
		fmt.Println("未知的返回。")
	}
	return false	
}
 	   
}
return false
}

func (w *WebWeChat) SendTextMsg(fakeid string, content string) bool {

	post_arg := url.Values{
		"tofakeid": {fakeid},
		"type":     {"1"},
		"content":  {content},
		"lang":     {"zh_CN"},
		"token":    {w.token},
		"t":        {"ajax-response"},
		"random":   {"8.1"},
		"imgcode":   {""},
	}
//推送消息

	req, _ := http.NewRequest("POST", SEND_URL, strings.NewReader(post_arg.Encode()))

    req.Header.Set(USER_AGENT_H,USER_AGENT)
	req.Header.Set(REFERER_H,fmt.Sprintf(REFER_URL, fakeid, w.token))
  
//两种方式都可以 
//    var ck string
//    for _,cookie :=range w.cookies {
//		ck += cookie.Name +"="+cookie.Value+";"
//	}
//	req.Header.Set("Cookie",ck)

    for _,cookie := range w.cookies {
		req.AddCookie(cookie)
	}
	
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")

	req.Header.Set("Accept-Encoding", "gzip, deflate");
	req.Header.Set("Accept-Language", "zh-cn,zh;q=0.8,en-us;q=0.5,en;q=0.3");
	req.Header.Set("Cache-Control", "no-cache");
	req.Header.Set("Connection", "keep-alive");
	// req.Header.Set("Content-Length", "100");

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8");
	req.Header.Set("Host", "mp.weixin.qq.com");
	req.Header.Set("Pragma", "no-cache");
	req.Header.Set("X-Requested-With", "XMLHttpRequest");


	client := new(http.Client)
	resp, _ := client.Do(req)
	if resp.StatusCode == 200 {
		fmt.Println("send weChat success!!!")
	} else {
		fmt.Println("send weChat failure!!!")
	}
	return false
}

func main() {
	wechat := NewWebWeChat()

	if wechat.login() == true {
		//玩家的fakeId
		tofakeid := "*****" 
		wechat.SendTextMsg(tofakeid, "Hello world")
	} else {
		fmt.Println("wechat login failed.")
	}
}

