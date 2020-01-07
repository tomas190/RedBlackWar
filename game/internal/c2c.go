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

		switch baseData.Event {
		case msgServerLogin:
			c4c.onServerLogin(baseData.Data)
			break
		case msgUserLogin:
			c4c.onUserLogin(baseData.Data)
			break
		case msgUserLogout:
			c4c.onUserLogout(baseData.Data)
			break
		case msgUserWinScore:
			c4c.onUserWinScore(baseData.Data)
			break
		case msgUserLoseScore:
			c4c.onUserLoseScore(baseData.Data)
			break
		case msgWinMoreThanNotice:
			c4c.onWinMoreThanNotice(baseData.Data)
			break
		case msgLockSettlement:
			c4c.onLockSettlement(baseData.Data)
			break
		case msgUnlockSettlement:
			c4c.onUnlockSettlement(baseData.Data)
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
		code, err := data["code"].(json.Number).Int64()
		if err != nil {
			log.Fatal(err.Error())
		}

		fmt.Println(code, reflect.TypeOf(code))
		if data["status"] == "SUCCESS" && code == 200 {
			log.Debug("<-------- serverLogin SUCCESS~!!! -------->")

			c4c.LoginStat = true
		}
	}
}

//onUserLogin 收到中心服的用户登录回应
func (c4c *Conn4Center) onUserLogin(msgBody interface{}) {
	data, ok := msgBody.(map[string]interface{})
	if !ok {
		log.Debug("onUserLogout Error")
	}

	code, err := data["code"].(json.Number).Int64()
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
		log.Debug("data:%v,ok:%v", data, ok)

		userInfo, ok := data["msg"].(map[string]interface{})
		var strId string
		var userData *UserCallback
		if ok {
			gameUser, uok := userInfo["game_user"].(map[string]interface{})
			if uok {
				nick := gameUser["game_nick"]
				headImg := gameUser["game_img"]
				userId := gameUser["id"]

				intID, err := userId.(json.Number).Int64()
				if err != nil {
					log.Fatal(err.Error())
				}
				strId = strconv.Itoa(int(intID))

				//找到等待登录玩家
				userData, ok = c4c.waitUser[strId]
				if ok {
					userData.Data.HeadImg = headImg.(string)
					userData.Data.NickName = nick.(string)
				}
			}
			gameAccount, okA := userInfo["game_account"].(map[string]interface{})

			if okA {
				balance := gameAccount["balance"]
				floatBalance, err := balance.(json.Number).Float64()
				if err != nil {
					log.Error(err.Error())
				}

				userData.Data.Account = floatBalance

				//调用玩家绑定回调函数
				if userData.Callback != nil {
					userData.Callback(&userData.Data)
				}
			}
		}
	}
}

func (c4c *Conn4Center) onUserLogout(msgBody interface{}) {
	data, ok := msgBody.(map[string]interface{})
	if !ok {
		log.Debug("onUserLogout Error")
	}

	code, err := data["code"].(json.Number).Int64()
	if err != nil {
		log.Error(err.Error())
	}

	if data["status"] == "SUCCESS" && code == 200 {
		log.Debug("<-------- UserLogout SUCCESS~ -------->")
		log.Debug("data:%v,ok:%v", data, ok)

		userInfo, ok := data["msg"].(map[string]interface{})
		var strId string
		var userData *UserCallback
		if ok {
			gameUser, uok := userInfo["game_user"].(map[string]interface{})
			if uok {
				nick := gameUser["game_nick"]
				headImg := gameUser["game_img"]
				userId := gameUser["id"]

				intID, err := userId.(json.Number).Int64()
				if err != nil {
					log.Fatal(err.Error())
				}
				strId = strconv.Itoa(int(intID))
				//找到等待登录玩家
				userData, ok = c4c.waitUser[strId]
				if ok {
					userData.Data.HeadImg = headImg.(string)
					userData.Data.NickName = nick.(string)
				}
			}
		}
	}
}

func (c4c *Conn4Center) onUserWinScore(msgBody interface{}) {
	data, ok := msgBody.(map[string]interface{})
	if !ok {
		log.Debug("onUserWinScore Error")
	}

	code, err := data["code"].(json.Number).Int64()
	if err != nil {
		log.Error(err.Error())
	}

	if code != 200 {
		cc.error("同步中心服赢钱失败", data)
		return
	}

	if data["status"] == "SUCCESS" && code == 200 {
		log.Debug("<-------- UserWinScore SUCCESS~ -------->")
		log.Debug("data:%v,ok:%v", data, ok)

		//将Win数据插入数据
		InsertWinMoney(msgBody)

		userInfo, ok := data["msg"].(map[string]interface{})
		if ok {
			jsonScore := userInfo["final_pay"]
			score, err := jsonScore.(json.Number).Float64()

			cc.log("同步中心服赢钱成功", score)

			if err != nil {
				log.Error(err.Error())
				return
			}
		}
	}
}

func (c4c *Conn4Center) onUserLoseScore(msgBody interface{}) {
	data, ok := msgBody.(map[string]interface{})
	if !ok {
		log.Debug("onUserLoseScore Error")
	}

	code, err := data["code"].(json.Number).Int64()
	if err != nil {
		log.Error(err.Error())
	}
	if code != 200 {
		cc.error("同步中心服输钱失败", data)
		return
	}

	if data["status"] == "SUCCESS" && code == 200 {
		log.Debug("<-------- UserLoseScore SUCCESS~ -------->")
		log.Debug("data:%v,ok:%v", data, ok)

		//将Lose数据插入数据
		InsertLoseMoney(msgBody)

		userInfo, ok := data["msg"].(map[string]interface{})
		if ok {
			jsonScore := userInfo["final_pay"]
			score, err := jsonScore.(json.Number).Float64()

			cc.log("同步中心服输钱成功", score)

			if err != nil {
				log.Error(err.Error())
				return
			}
		}
	}
}

//onWinMoreThanNotice 加锁金额
func (c4c *Conn4Center) onLockSettlement(msgBody interface{}) {
	data, ok := msgBody.(map[string]interface{})
	if ok {
		code, err := data["code"].(json.Number).Int64()
		if err != nil {
			log.Fatal(err.Error())
		}

		fmt.Println(code, reflect.TypeOf(code))
		if data["status"] == "SUCCESS" && code == 200 {
			log.Debug("<-------- onLockSettlement SUCCESS~!!! -------->")
		}
	}
}

//onWinMoreThanNotice 解锁金额
func (c4c *Conn4Center) onUnlockSettlement(msgBody interface{}) {
	data, ok := msgBody.(map[string]interface{})
	if ok {
		code, err := data["code"].(json.Number).Int64()
		if err != nil {
			log.Fatal(err.Error())
		}

		fmt.Println(code, reflect.TypeOf(code))
		if data["status"] == "SUCCESS" && code == 200 {
			log.Debug("<-------- onUnlockSettlement SUCCESS~!!! -------->")
		}
	}
}

//onWinMoreThanNotice 服务器登录
func (c4c *Conn4Center) onWinMoreThanNotice(msgBody interface{}) {
	data, ok := msgBody.(map[string]interface{})
	if ok {
		code, err := data["code"].(json.Number).Int64()
		if err != nil {
			log.Fatal(err.Error())
		}

		fmt.Println(code, reflect.TypeOf(code))
		if data["status"] == "SUCCESS" && code == 200 {
			log.Debug("<-------- onWinMoreThanNotice SUCCESS~!!! -------->")
		}
	}
}

//ServerLoginCenter 服务器登录Center
func (c4c *Conn4Center) ServerLoginCenter() {
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
func (c4c *Conn4Center) UserLoginCenter(userId string, password string, token string, callback func(data *Player)) {
	if !c4c.LoginStat {
		log.Debug("<-------- RedBlack-War not ready~!!! -------->")
		return
	}
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
	c4c.waitUser[userId].Data.Id = userId
	c4c.waitUser[userId].Callback = callback
}

//UserLogoutCenter 用户登出
func (c4c *Conn4Center) UserLogoutCenter(userId string, password string, token string) {
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
	log.Debug("Msg to Send Center:%v", string(codeData))

	err2 := c4c.conn.WriteMessage(websocket.TextMessage, []byte(codeData))
	if err2 != nil {
		log.Fatal(err2.Error())
	}
}

//UserSyncWinScore 同步赢分
func (c4c *Conn4Center) UserSyncWinScore(p *Player, timeUnix int64, timeStr, reason string) {
	winOrder := p.Id + "_" + timeStr + "_win"
	baseData := &BaseMessage{}
	baseData.Event = msgUserWinScore
	userWin := &UserChangeScore{}
	userWin.Auth.DevName = conf.Server.DevName
	userWin.Auth.DevKey = c4c.DevKey
	userWin.Info.CreateTime = timeUnix
	userWin.Info.GameId = c4c.GameId
	userWin.Info.ID = p.Id
	userWin.Info.LockMoney = 0
	userWin.Info.Money = p.WinResultMoney
	userWin.Info.Order = winOrder
	userWin.Info.PayReason = reason
	userWin.Info.PreMoney = 0
	userWin.Info.RoundId = p.room.RoomId
	baseData.Data = userWin
	c4c.SendMsg2Center(baseData)
}

//UserSyncWinScore 同步输分
func (c4c *Conn4Center) UserSyncLoseScore(p *Player, timeUnix int64, timeStr, reason string) {
	loseOrder := p.Id + "_" + timeStr + "_lose"
	baseData := &BaseMessage{}
	baseData.Event = msgUserLoseScore
	userLose := &UserChangeScore{}
	userLose.Auth.DevName = conf.Server.DevName
	userLose.Auth.DevKey = c4c.DevKey
	userLose.Info.CreateTime = timeUnix
	userLose.Info.GameId = c4c.GameId
	userLose.Info.ID = p.Id
	userLose.Info.LockMoney = 0
	userLose.Info.Money = p.LoseResultMoney
	userLose.Info.Order = loseOrder
	userLose.Info.PayReason = reason
	userLose.Info.PreMoney = 0
	userLose.Info.RoundId = p.room.RoomId
	baseData.Data = userLose
	c4c.SendMsg2Center(baseData)
}

//锁钱
func (c4c *Conn4Center) LockSettlement(p *Player) {
	timeStr := time.Now().Format("2006-01-02_15:04:05")
	loseOrder := p.Id + "_" + timeStr + "_LockMoney"

	baseData := &BaseMessage{}
	baseData.Event = msgLockSettlement
	lockMoney := &UserChangeScore{}
	lockMoney.Auth.DevName = conf.Server.DevName
	lockMoney.Auth.DevKey = c4c.DevKey
	lockMoney.Info.CreateTime = time.Now().Unix()
	lockMoney.Info.GameId = c4c.GameId
	lockMoney.Info.ID = p.Id
	lockMoney.Info.LockMoney = p.Account
	lockMoney.Info.Money = 0
	lockMoney.Info.Order = loseOrder
	lockMoney.Info.PayReason = "lockMoney"
	lockMoney.Info.PreMoney = 0
	lockMoney.Info.RoundId = p.room.RoomId
	baseData.Data = lockMoney
	c4c.SendMsg2Center(baseData)
}

//解锁
func (c4c *Conn4Center) UnlockSettlement(p *Player) {
	timeStr := time.Now().Format("2006-01-02_15:04:05")
	loseOrder := p.Id + "_" + timeStr + "_UnlockMoney"

	baseData := &BaseMessage{}
	baseData.Event = msgUnlockSettlement
	lockMoney := &UserChangeScore{}
	lockMoney.Auth.DevName = conf.Server.DevName
	lockMoney.Auth.DevKey = c4c.DevKey
	lockMoney.Info.CreateTime = time.Now().Unix()
	lockMoney.Info.GameId = c4c.GameId
	lockMoney.Info.ID = p.Id
	lockMoney.Info.LockMoney = p.Account
	lockMoney.Info.Money = 0
	lockMoney.Info.Order = loseOrder
	lockMoney.Info.PayReason = "UnlockMoney"
	lockMoney.Info.PreMoney = 0
	lockMoney.Info.RoundId = p.room.RoomId
	baseData.Data = lockMoney
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

//GetRandNumber 获取中心服随机数值
func GetRandNumber() ([]uint8, error) {
	res, err := http.Get(conf.Server.RandNum)
	if err != nil {
		log.Debug("再次获取随机数值失败: %v", err)
		return nil, err
	}

	result, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		log.Error("解析随机数值失败: %v", err)
	}

	//fmt.Printf("读取的奖源池数据: %s", result)

	var users interface{}
	err2 := json.Unmarshal(result, &users)
	if err2 != nil {
		panic(err2)
	}

	var RandNum []uint8

	data, ok := users.(map[string]interface{})
	if ok {
		msg := data["msg"]
		codeString := msg.(string)
		codeSlice := strings.Split(codeString, " ")

		//log.Debug("获取的随机数值: %v", codeSlice)

		for _, val := range codeSlice {
			switch val {
			case "1":
				RandNum = append(RandNum, 14)
			case "2":
				RandNum = append(RandNum, 2)
			case "3":
				RandNum = append(RandNum, 3)
			case "4":
				RandNum = append(RandNum, 4)
			case "5":
				RandNum = append(RandNum, 5)
			case "6":
				RandNum = append(RandNum, 6)
			case "7":
				RandNum = append(RandNum, 7)
			case "8":
				RandNum = append(RandNum, 8)
			case "9":
				RandNum = append(RandNum, 9)
			case "10":
				RandNum = append(RandNum, 10)
			case "11":
				RandNum = append(RandNum, 11)
			case "12":
				RandNum = append(RandNum, 12)
			case "13":
				RandNum = append(RandNum, 13)
			case "14":
				RandNum = append(RandNum, 30)
			case "15":
				RandNum = append(RandNum, 18)
			case "16":
				RandNum = append(RandNum, 19)
			case "17":
				RandNum = append(RandNum, 20)
			case "18":
				RandNum = append(RandNum, 21)
			case "19":
				RandNum = append(RandNum, 22)
			case "20":
				RandNum = append(RandNum, 23)
			case "21":
				RandNum = append(RandNum, 24)
			case "22":
				RandNum = append(RandNum, 25)
			case "23":
				RandNum = append(RandNum, 26)
			case "24":
				RandNum = append(RandNum, 27)
			case "25":
				RandNum = append(RandNum, 28)
			case "26":
				RandNum = append(RandNum, 29)
			case "27":
				RandNum = append(RandNum, 46)
			case "28":
				RandNum = append(RandNum, 34)
			case "29":
				RandNum = append(RandNum, 35)
			case "30":
				RandNum = append(RandNum, 36)
			case "31":
				RandNum = append(RandNum, 37)
			case "32":
				RandNum = append(RandNum, 38)
			case "33":
				RandNum = append(RandNum, 39)
			case "34":
				RandNum = append(RandNum, 40)
			case "35":
				RandNum = append(RandNum, 41)
			case "36":
				RandNum = append(RandNum, 42)
			case "37":
				RandNum = append(RandNum, 43)
			case "38":
				RandNum = append(RandNum, 44)
			case "39":
				RandNum = append(RandNum, 45)
			case "40":
				RandNum = append(RandNum, 62)
			case "41":
				RandNum = append(RandNum, 50)
			case "42":
				RandNum = append(RandNum, 51)
			case "43":
				RandNum = append(RandNum, 52)
			case "44":
				RandNum = append(RandNum, 53)
			case "45":
				RandNum = append(RandNum, 54)
			case "46":
				RandNum = append(RandNum, 55)
			case "47":
				RandNum = append(RandNum, 56)
			case "48":
				RandNum = append(RandNum, 57)
			case "49":
				RandNum = append(RandNum, 58)
			case "50":
				RandNum = append(RandNum, 59)
			case "51":
				RandNum = append(RandNum, 60)
			case "52":
				RandNum = append(RandNum, 61)
			}
		}
	}
	return RandNum, nil
}
