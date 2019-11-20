package internal

import (
	"RedBlack-War/conf"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/name5566/leaf/log"
	"io/ioutil"
	"net/http"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"
)

//CGTokenRsp 接受Token结构体
type CGTokenRsp struct {
	Token string
}

//CGCenterRsp 中心返回消息结构体
type CGCenterRsp struct {
	Status string
	Code   int
	Msg    *CGTokenRsp
}

//Conn4Center 连接到Center(中心服务器)的网络协议处理器
type Conn4Center struct {
	GameId    string
	centerUrl string
	token     string
	DevKey    string
	conn      *websocket.Conn

	//除于登录成功状态
	LoginStat bool

	closebreathchan  chan bool
	closereceivechan chan bool

	//待处理的用户登录请求
	waitUser map[string]*UserCallback
}

//Init 初始化
func (c4c *Conn4Center) Init() {
	c4c.GameId = conf.Server.GameID
	c4c.DevKey = conf.Server.DevKey
	c4c.LoginStat = false

	c4c.waitUser = make(map[string]*UserCallback)
	//go changeToken()
}

var gt CGCenterRsp

//func changeToken() {
//	for {
//		time.Sleep(time.Second * 7000)
//		getToken()
//		c4c.token = gt.Msg.Token
//	}
//}

func getToken() {
	// 拼接center Url
	url4Center := fmt.Sprintf("%s?dev_key=%s&dev_name=%s", conf.Server.TokenServer, c4c.DevKey, conf.Server.DevName)

	log.Debug("<--- TokenServer Url --->: %v ", conf.Server.TokenServer)
	log.Debug("<--- Center access Url --->: %v ", url4Center)

	resp, err1 := http.Get(url4Center)
	if err1 != nil {
		panic(err1.Error())
	}
	log.Debug("<--- resp --->: %v ", resp)

	defer resp.Body.Close()

	if err1 == nil && resp.StatusCode == 200 {
		body, err2 := ioutil.ReadAll(resp.Body)
		if err2 != nil {
			panic(err2.Error())
		}
		//log.Debug("<----- resp.StatusCode ----->: %v", resp.StatusCode)
		log.Debug("<--- body --->: %v ,<--- err2 --->: %v", string(body), err2)

		err3 := json.Unmarshal(body, &gt)
		log.Debug("<--- err3 --->: %v <--- Results --->: %v", err3, gt)
	}
}

//onDestroy 销毁用户
func (c4c *Conn4Center) onDestroy() {
	log.Debug("Conn4Center onDestroy ~")
	//c4c.UserLogoutCenter("991738698","123456") //测试用户 和 密码
}

//ReqCenterToken 向中心服务器请求token
func (c4c *Conn4Center) ReqCenterToken() {
	// 拼接center Url
	url4Center := fmt.Sprintf("%s?dev_key=%s&dev_name=%s", conf.Server.TokenServer, c4c.DevKey, conf.Server.DevName)

	//log.Debug("<--- TokenServer Url --->: %v ", conf.Server.TokenServer)
	log.Debug("<--- Center access Url --->: %v ", url4Center)

	resp, err1 := http.Get(url4Center)
	if err1 != nil {
		panic(err1.Error())
	}
	log.Debug("<--- resp --->: %v ", resp)

	defer resp.Body.Close()

	if err1 == nil && resp.StatusCode == 200 {
		body, err2 := ioutil.ReadAll(resp.Body)
		if err2 != nil {
			panic(err2.Error())
		}
		//log.Debug("<----- resp.StatusCode ----->: %v", resp.StatusCode)
		log.Debug("<--- body --->: %v ,<--- err2 --->: %v", string(body), err2)

		var t CGCenterRsp
		err3 := json.Unmarshal(body, &t)
		log.Debug("<--- err3 --->: %v <--- Results --->: %v", err3, t)

		if t.Status == "SUCCESS" && t.Code == 200 {
			c4c.token = conf.Server.DevName
			c4c.CreatConnect()
		} else {
			log.Fatal("<--- Request Token Fail~ --->")
		}
	}
}

//CreatConnect 和Center建立链接
func (c4c *Conn4Center) CreatConnect() {
	// TODO
	c4c.centerUrl = conf.Server.CenterUrl
	//c4c.centerUrl = "ws://172.16.1.41:9502/" //Pre
	//c4c.centerUrl = "ws://172.16.100.2:9502/" //上线
	//c4c.centerUrl = "ws" + strings.TrimPrefix(conf.Server.CenterServer, "http") //域名生成使用

	log.Debug("--- dial: --- : %v", c4c.centerUrl)
	conn, rsp, err := websocket.DefaultDialer.Dial(c4c.centerUrl, nil)
	c4c.conn = conn
	log.Debug("<--- Dial rsp --->: %v", rsp)

	if err != nil {
		log.Fatal(err.Error())
	} else {
		c4c.Run()
	}
}

func (c4c *Conn4Center) ReConnect() {
	go func() {
		for {
			c4c.closebreathchan <- true
			c4c.closereceivechan <- true
			if c4c.LoginStat == true {
				return
			}
			time.Sleep(time.Second * 5)
			c4c.CreatConnect()
		}
	}()
}

//Run 开始运行,监听中心服务器的返回
func (c4c *Conn4Center) Run() {
	ticker := time.NewTicker(time.Second * 5)
	go func() {
		for { //循环
			select {
			case <-ticker.C:
				c4c.onBreath()
				break
			case <-c4c.closebreathchan:
				return
			}
		}
	}()

	go func() {
		for {
			select {
			case <-c4c.closereceivechan:
				return
			default:
				typeId, message, err := c4c.conn.ReadMessage()
				if err != nil {
					log.Debug("Here is error by ReadMessage~ %v", err)
					log.Error(err.Error())
				}
				//log.Debug("Receive a message from Center~")
				//log.Debug("typeId: %v", typeId)
				//log.Debug("message: %v", string(message))
				if typeId == -1 {
					log.Debug("中心服异常消息~")
					c4c.LoginStat = false
					c4c.ReConnect()
					return
				} else {
					c4c.onReceive(typeId, message)
				}
				break
			}
		}
	}()

	c4c.ServerLoginCenter()
}

//onBreath 中心服心跳
func (c4c *Conn4Center) onBreath() {
	err := c4c.conn.WriteMessage(websocket.TextMessage, []byte(""))
	if err != nil {
		log.Error(err.Error())
	}
}

//onReceive 接收消息
func (c4c *Conn4Center) onReceive(messType int, messBody []byte) {
	if messType == websocket.TextMessage {
		baseData := &BaseMessage{}

		decoder := json.NewDecoder(strings.NewReader(string(messBody)))
		decoder.UseNumber()

		err := decoder.Decode(&baseData)
		if err != nil {
			log.Error(err.Error())
		}
		//log.Debug("<-------- baseData -------->: %v", baseData)

		switch baseData.Event {
		case msgServerLogin:
			c4c.onServerLogin(baseData.Data)
			log.Debug("<-------- baseData onServerLogin -------->")
			break
		case msgUserLogin:
			c4c.onUserLogin(baseData.Data)
			log.Debug("<-------- baseData msgUserLogin -------->")
			break
		case msgUserLogout:
			c4c.onUserLogout(baseData.Data)
			log.Debug("<-------- baseData onUserLogout -------->")
			break
		case msgUserWinScore:
			c4c.onUserWinScore(baseData.Data)
			log.Debug("<-------- baseData msgUserWinScore -------->")
			break
		case msgUserLoseScore:
			c4c.onUserLoseScore(baseData.Data)
			log.Debug("<-------- baseData msgUserLoseScore -------->")
			break
		case msgWinMoreThanNotice:
			c4c.onWinMoreThanNotice(baseData.Data)
			log.Debug("<-------- baseData onWinMoreThanNotice -------->")
			break
		default:
			log.Error("Receive a message but don't identify~")
		}
	}
}

//onServerLogin 服务器登录
func (c4c *Conn4Center) onServerLogin(msgBody interface{}) {
	log.Debug("<-------- onServerLogin -------->: %v", msgBody)
	data, ok := msgBody.(map[string]interface{})
	if ok {
		//fmt.Println(data["status"], reflect.TypeOf(data["status"]))
		//fmt.Println(data["code"], reflect.TypeOf(data["code"]))
		//fmt.Println(data["msg"], reflect.TypeOf(data["msg"]))
		code, err := data["code"].(json.Number).Int64()
		//fmt.Println("code,err", code, err)
		if err != nil {
			log.Fatal(err.Error())
		}

		fmt.Println(code, reflect.TypeOf(code))
		if data["status"] == "SUCCESS" && code == 200 {
			log.Debug("<-------- serverLogin success~!!! -------->")

			c4c.LoginStat = true
		}
	}
}

//onUserLogin 收到中心服的用户登录回应
func (c4c *Conn4Center) onUserLogin(msgBody interface{}) {
	log.Debug("<-------- onUserLogin -------->: %v", msgBody)
	data, ok := msgBody.(map[string]interface{})
	log.Debug("data:%v, ok:%v", data, ok)

	//fmt.Println(data["status"], reflect.TypeOf(data["status"]))
	//fmt.Println(data["code"], reflect.TypeOf(data["code"]))
	//fmt.Println(data["msg"], reflect.TypeOf(data["msg"]))
	code, err := data["code"].(json.Number).Int64()
	//fmt.Println("code,err", code, err)
	if err != nil {
		log.Error(err.Error())
		return
	}

	if code != 200 {
		cc.error("同步中心服用户登录失败", data)
		return
	}

	if data["status"] == "SUCCESS" && code == 200 {
		log.Debug("<-------- UserLogin SUCCESS~ -------->")
		userInfo, ok := data["msg"].(map[string]interface{})
		var strId string
		var userData *UserCallback
		if ok {
			log.Debug("userInfo: %v", userInfo)
			gameUser, uok := userInfo["game_user"].(map[string]interface{})
			if uok {
				log.Debug("gameUser: %v", gameUser)
				nick := gameUser["game_nick"]
				headImg := gameUser["game_img"]
				userId := gameUser["id"]
				//log.Debug("nick: %v", nick)
				//log.Debug("headImg: %v", headImg)
				//log.Debug("userId: %v %v", userId, reflect.TypeOf(userId))

				intID, err := userId.(json.Number).Int64()
				if err != nil {
					log.Fatal(err.Error())
				}
				strId = strconv.Itoa(int(intID))
				log.Debug("strId: %v %v", strId, reflect.TypeOf(strId))

				//找到等待登录玩家
				userData, ok = c4c.waitUser[strId]
				if ok {
					userData.Data.HeadImg = headImg.(string)
					userData.Data.Nick = nick.(string)
				}
			}
			gameAccount, okA := userInfo["game_account"].(map[string]interface{})

			if okA {
				log.Debug("<-------- gameAccount -------->: %v", gameAccount)
				balance := gameAccount["balance"]
				//log.Debug("<-------- balance -------->: %v %v", balance, reflect.TypeOf(balance))
				floatBalance, err := balance.(json.Number).Float64()
				if err != nil {
					log.Error(err.Error())
				}

				userData.Data.Score = floatBalance

				//调用玩家绑定回调函数
				if userData.Callback != nil {
					userData.Callback(&userData.Data)
				}
			}
		}
	}
}

func (c4c *Conn4Center) onUserLogout(msgBody interface{}) {
	log.Debug("<-------- onUserLogout -------->: %v", msgBody)

	data, ok := msgBody.(map[string]interface{})
	log.Debug("data:%v, ok:%v", data, ok)

	code, err := data["code"].(json.Number).Int64()
	if err != nil {
		log.Error(err.Error())
	}

	if data["status"] == "SUCCESS" && code == 200 {
		log.Debug("<-------- UserLogin SUCCESS~ -------->")
		userInfo, ok := data["msg"].(map[string]interface{})
		var strId string
		var userData *UserCallback
		if ok {
			log.Debug("userInfo: %v", userInfo)
			gameUser, uok := userInfo["game_user"].(map[string]interface{})
			if uok {
				log.Debug("gameUser: %v", gameUser)
				nick := gameUser["game_nick"]
				headImg := gameUser["game_img"]
				userId := gameUser["id"]

				intID, err := userId.(json.Number).Int64()
				if err != nil {
					log.Fatal(err.Error())
				}
				strId = strconv.Itoa(int(intID))
				log.Debug("strId: %v %v", strId, reflect.TypeOf(strId))

				//找到等待登录玩家
				userData, ok = c4c.waitUser[strId]
				if ok {
					userData.Data.HeadImg = headImg.(string)
					userData.Data.Nick = nick.(string)
				}
			}
			gameAccount, okA := userInfo["game_account"].(map[string]interface{})
			if okA {
				log.Debug("<-------- gameAccount -------->: %v", gameAccount)
			}
		}
	}
}

func (c4c *Conn4Center) onUserWinScore(msgBody interface{}) {
	log.Debug("<-------- onUserWinScore -------->: %v", msgBody)
	//将Win数据插入数据
	InsertWinMoney(msgBody)

	data, ok := msgBody.(map[string]interface{})

	log.Debug("<-------- data -------->:%v, <-------- ok -------->:%v", data, ok)

	//fmt.Println(data["status"], reflect.TypeOf(data["status"]))
	//fmt.Println(data["code"], reflect.TypeOf(data["code"]))
	//fmt.Println(data["msg"], reflect.TypeOf(data["msg"]))
	code, err := data["code"].(json.Number).Int64()
	//fmt.Println("code,err", code, err)
	if err != nil {
		log.Error(err.Error())
	}

	if code != 200 {
		cc.error("同步中心服赢钱失败", data)
		return
	}

	log.Debug("data:%v, ok:%v", data, ok)
	if data["status"] == "SUCCESS" && code == 200 {
		log.Debug("<-------- UserWinScore SUCCESS~ -------->")
		userInfo, ok := data["msg"].(map[string]interface{})
		if ok {
			userId := userInfo["id"]
			log.Debug("userId: %v, %v", userId, reflect.TypeOf(userId))

			intID, err := userId.(json.Number).Int64()
			if err != nil {
				log.Error(err.Error())
				return
			}
			strID := strconv.Itoa(int(intID))
			log.Debug("<-------- strID -------->: %v, %v", strID, reflect.TypeOf(strID))

			jsonScore := userInfo["final_pay"]
			score, err := jsonScore.(json.Number).Float64()

			log.Debug("<--------- final win score: %v", score)

			cc.log("同步中心服赢钱成功", score)

			if err != nil {
				log.Error(err.Error())
				return
			}
		}
	}
}

func (c4c *Conn4Center) onUserLoseScore(msgBody interface{}) {
	log.Debug("<-------- onUserLoseScore -------->: %v", msgBody)
	//将Lose数据插入数据
	InsertLoseMoney(msgBody)

	data, ok := msgBody.(map[string]interface{})

	log.Debug("data:%v, ok:%v", data, ok)

	//fmt.Println(data["status"], reflect.TypeOf(data["status"]))
	//fmt.Println(data["code"], reflect.TypeOf(data["code"]))
	//fmt.Println(data["msg"], reflect.TypeOf(data["msg"]))
	code, err := data["code"].(json.Number).Int64()
	//fmt.Println("code,err", code, err)
	if err != nil {
		log.Error(err.Error())
	}

	if code != 200 {
		cc.error("同步中心服输钱失败", data)
		return
	}

	fmt.Println(code, err)
	if data["status"] == "SUCCESS" && code == 200 {
		fmt.Println("UserWinScore SUCCESS~")
		userInfo, ok := data["msg"].(map[string]interface{})
		if ok {
			userId := userInfo["id"]
			log.Debug("userId: %v, %v", userId, reflect.TypeOf(userId))

			intID, err := userId.(json.Number).Int64()
			if err != nil {
				log.Error(err.Error())
				return
			}
			strID := strconv.Itoa(int(intID))
			log.Debug("<-------- strID -------->: %v, %v", strID, reflect.TypeOf(strID))

			jsonScore := userInfo["final_pay"]
			score, err := jsonScore.(json.Number).Float64()

			log.Debug("<-------- final lose score -------->: %v", score)

			cc.log("同步中心服输钱成功", score)

			if err != nil {
				log.Error(err.Error())
				return
			}
		}
	}
}

//onWinMoreThanNotice 服务器登录
func (c4c *Conn4Center) onWinMoreThanNotice(msgBody interface{}) {
	log.Debug("<-------- onWinMoreThanNotice -------->: %v", msgBody)
	data, ok := msgBody.(map[string]interface{})
	if ok {
		code, err := data["code"].(json.Number).Int64()
		if err != nil {
			log.Fatal(err.Error())
		}

		fmt.Println(code, reflect.TypeOf(code))
		if data["status"] == "SUCCESS" && code == 200 {
			log.Debug("<-------- onWinMoreThanNotice success~!!! -------->")
		}
	}
}

//ServerLoginCenter 服务器登录Center
func (c4c *Conn4Center) ServerLoginCenter() {
	log.Debug("<-------- ServerLoginCenter c4c.token -------->: %v", c4c.token)
	baseData := &BaseMessage{}
	baseData.Event = msgServerLogin
	baseData.Data = ServerLogin{
		Host:    conf.Server.CenterServer,
		Port:    conf.Server.CenterServerPort,
		GameId:  c4c.GameId,
		DevName: conf.Server.DevName,
		DevKey:  c4c.DevKey,
	}
	// 发送消息到中心服
	c4c.SendMsg2Center(baseData)
}

//UserLoginCenter 用户登录
func (c4c *Conn4Center) UserLoginCenter(userId string, password string, token string, callback func(data *UserInfo)) {
	if !c4c.LoginStat {
		log.Debug("<-------- RedBlack-War not ready~!!! -------->")
		return
	}

	log.Debug("<-------- UserLoginCenter c4c.token -------->: %v", c4c.token)
	baseData := &BaseMessage{}
	baseData.Event = msgUserLogin
	if password != "" {
		baseData.Data = &UserReq{
			ID:       userId,
			PassWord: password,
			GameId:   c4c.GameId,
			DevName:  conf.Server.DevName,
			DevKey:   c4c.DevKey}
	} else {
		baseData.Data = &UserReq{
			ID:      userId,
			Token:   token,
			GameId:  c4c.GameId,
			DevName: conf.Server.DevName,
			DevKey:  c4c.DevKey}
	}

	c4c.SendMsg2Center(baseData)

	//加入待处理map，等待处理
	c4c.waitUser[userId] = &UserCallback{}
	c4c.waitUser[userId].Data.ID = userId
	c4c.waitUser[userId].Callback = callback
}

//UserLogoutCenter 用户登出
func (c4c *Conn4Center) UserLogoutCenter(userId string, password string, token string) {
	log.Debug("<-------- UserLogoutCenter c4c.token -------->: %v", c4c.token)
	base := &BaseMessage{}
	base.Event = msgUserLogout
	if password != "" {
		base.Data = &UserReq{
			ID:       userId,
			PassWord: password,
			GameId:   c4c.GameId,
			DevName:  conf.Server.DevName,
			DevKey:   c4c.DevKey}
	} else {
		base.Data = &UserReq{
			ID:      userId,
			Token:   token,
			GameId:  c4c.GameId,
			DevName: conf.Server.DevName,
			DevKey:  c4c.DevKey}
	}

	// 发送消息到中心服
	c4c.SendMsg2Center(base)
}

//SendMsg2Center 发送消息到中心服
func (c4c *Conn4Center) SendMsg2Center(data interface{}) {
	// Json序列化
	codeData, err1 := json.Marshal(data)
	if err1 != nil {
		log.Error(err1.Error())
	}
	log.Debug("<-------- 发送消息中心服 -------->: %v", string(codeData))

	err2 := c4c.conn.WriteMessage(websocket.TextMessage, []byte(codeData))
	if err2 != nil {
		log.Fatal(err2.Error())
	}
	log.Debug("<==========================================================>")
}

//UserSyncWinScore 同步赢分
func (c4c *Conn4Center) UserSyncWinScore(p *Player, timeUnix int64, timeStr, reason string) {
	winOrder := p.Id + "_" + timeStr + "_win"
	log.Debug("<-------- GenWinOrder -------->: %v", winOrder)
	baseData := &BaseMessage{}
	baseData.Event = msgUserWinScore
	userWin := &UserChangeScore{}
	userWin.Auth.DevName = conf.Server.DevName
	userWin.Auth.DevKey = c4c.DevKey
	userWin.Info.CreateTime = timeUnix
	userWin.Info.GameId = c4c.GameId
	userWin.Info.ID = p.Id
	userWin.Info.LockMoney = p.Account
	userWin.Info.Money = p.WinResultMoney
	userWin.Info.Order = winOrder
	userWin.Info.PayReason = reason
	userWin.Info.PreMoney = 0
	userWin.Info.RoundId = p.room.RoomId
	baseData.Data = userWin
	log.Debug("<<===== UserSyncWinScore: %v =====>>", baseData)
	c4c.SendMsg2Center(baseData)
}

//UserSyncWinScore 同步输分
func (c4c *Conn4Center) UserSyncLoseScore(p *Player, timeUnix int64, timeStr, reason string) {
	loseOrder := p.Id + "_" + timeStr + "_lose"
	log.Debug("<-------- GenLoseOrder -------->: %v", loseOrder)

	baseData := &BaseMessage{}
	baseData.Event = msgUserLoseScore
	userLose := &UserChangeScore{}
	userLose.Auth.DevName = conf.Server.DevName
	userLose.Auth.DevKey = c4c.DevKey
	userLose.Info.CreateTime = timeUnix
	userLose.Info.GameId = c4c.GameId
	userLose.Info.ID = p.Id
	userLose.Info.LockMoney = p.Account
	userLose.Info.Money = p.LoseResultMoney
	userLose.Info.Order = loseOrder
	userLose.Info.PayReason = reason
	userLose.Info.PreMoney = 0
	userLose.Info.RoundId = p.room.RoomId
	baseData.Data = userLose
	log.Debug("<<===== UserSyncLoseScore: %v =====>>", baseData)
	c4c.SendMsg2Center(baseData)
}

//UserSyncScoreChange 同步尚未同步过的输赢分
func (c4c *Conn4Center) UserSyncScoreChange(p *Player, reason string) {
	timeStr := time.Now().Format("2006-01-02_15:04:05")
	nowTime := time.Now().Unix()

	//同时同步赢分和输分
	c4c.UserSyncWinScore(p, nowTime, timeStr, reason)
	c4c.UserSyncLoseScore(p, nowTime, timeStr, reason)
}

func (c4c *Conn4Center) NoticeWinMoreThan(playerId, playerName string, winGold float64) {
	log.Debug("<-------- NoticeWinMoreThan  -------->")
	msg := fmt.Sprintf("<size=20><color=YELLOW>恭喜!</color><color=orange>%s</color><color=YELLOW>在</color></><color=WHITE><b><size=25>红黑大战</color></b></><color=YELLOW><size=20>中一把赢了</color></><color=YELLOW><b><size=35>%.2f</color></b></><color=YELLOW><size=20>金币！</color></>", playerName, winGold)

	base := &BaseMessage{}
	base.Event = msgWinMoreThanNotice
	base.Data = &Notice{
		DevName: conf.Server.DevName,
		DevKey:  conf.Server.DevKey,
		ID:      playerId,
		GameId:  c4c.GameId,
		Type:    2000,
		Message: msg,
		Topic:   "系统提示",
	}
	c4c.SendMsg2Center(base)
}

//Init 初始化
func (cc *mylog) Init() {

}
func (cc *mylog) log(v ...interface{}) {
	senddata := logmsg{
		Type:     "LOG",
		From:     "RedBlack-War",
		GameName: "红黑大战",
		Id:       conf.Server.GameID,
		Host:     "",
		Time:     time.Now().Unix(),
	}

	_, file, line, ok := runtime.Caller(2)
	if ok {
		senddata.File = file
		senddata.Line = line
	}
	Msg := fmt.Sprintln(v...)
	senddata.Msg = Msg
	cc.sendMsg(senddata)
}

func (cc *mylog) debug(v ...interface{}) {
	senddata := logmsg{
		Type:     "DEG",
		From:     "RedBlack-War",
		GameName: "红黑大战",
		Id:       conf.Server.GameID,
		Host:     "",
		Time:     time.Now().Unix(),
	}

	_, file, line, ok := runtime.Caller(2)
	if ok {
		senddata.File = file
		senddata.Line = line
	}
	Msg := fmt.Sprintln(v...)
	senddata.Msg = Msg
	cc.sendMsg(senddata)
}

func (cc *mylog) error(v ...interface{}) {
	senddata := logmsg{
		Type:     "ERR",
		From:     "RedBlack-War",
		GameName: "RedBlack-War",
		Id:       conf.Server.GameID,
		Host:     "",
		Time:     time.Now().Unix(),
	}

	_, file, line, ok := runtime.Caller(2)
	if ok {
		senddata.File = file
		senddata.Line = line
	}
	Msg := fmt.Sprintln(v...)
	senddata.Msg = Msg
	cc.sendMsg(senddata)
}

func (cc *mylog) sendMsg(senddata logmsg) {
	bodyJson, err1 := json.Marshal(senddata)
	if err1 != nil {
		log.Error(err1.Error())
	}
	req, err2 := http.NewRequest(http.MethodPost, conf.Server.LogAddr, bytes.NewBuffer(bodyJson))
	if err2 != nil {
		log.Error(err1.Error())
	}
	if req != nil {
		req.Header.Add("content-type", "application/json")
		err3 := req.Body.Close()
		if err3 != nil {
			log.Error(err1.Error())
		}
	}
}

var cc mylog
