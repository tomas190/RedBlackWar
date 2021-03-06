package internal

import (
	"RedBlack-War/conf"
	pb_msg "RedBlack-War/msg/Protocal"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/name5566/leaf/log"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"net/http"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"
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

// 添加互斥锁，防止websocket写并发
var writeMutex sync.Mutex

//Init 初始化
func (c4c *Conn4Center) Init() {
	c4c.GameId = conf.Server.GameID
	c4c.DevKey = conf.Server.DevKey
	c4c.LoginStat = false

	c4c.waitUser = make(map[string]*UserCallback)
}

//CreatConnect 和Center建立链接
func (c4c *Conn4Center) CreatConnect() {
	c4c.centerUrl = conf.Server.CenterUrl

	log.Debug("--- dial: --- : %v", c4c.centerUrl)
	for {
		conn, rsp, err := websocket.DefaultDialer.Dial(c4c.centerUrl, nil)
		log.Debug("<--- Dial rsp --->: %v", rsp)
		if err == nil {
			c4c.conn = conn
			break
		}
		time.Sleep(time.Second * 5)
	}

	c4c.ServerLoginCenter()

	c4c.Run()
}

func (c4c *Conn4Center) ReConnect() {
	if c4c.LoginStat == true {
		return
	}
	time.Sleep(time.Second * 5)

	c4c.centerUrl = conf.Server.CenterUrl

	log.Debug("--- dial: --- : %v", c4c.centerUrl)
	for {
		conn, rsp, err := websocket.DefaultDialer.Dial(c4c.centerUrl, nil)
		log.Debug("<--- Dial rsp --->: %v", rsp)
		if err == nil {
			c4c.conn = conn
			break
		}
		time.Sleep(time.Second * 5)
	}

	c4c.ServerLoginCenter()
}

//Run 开始运行,监听中心服务器的返回
func (c4c *Conn4Center) Run() {
	ticker := time.NewTicker(time.Second * 3)
	log.Debug("发送心跳!")
	go func() {
		for { //循环
			select {
			case <-ticker.C:
				c4c.onBreath()
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
				}
				if typeId == -1 {
					log.Debug("中心服异常消息~")
					c4c.LoginStat = false
					c4c.ReConnect()
				} else {
					c4c.onReceive(typeId, message)
				}
			}
		}
	}()
}

//onBreath 中心服心跳
func (c4c *Conn4Center) onBreath() {
	writeMutex.Lock()
	defer writeMutex.Unlock()
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

		log.Debug("code:%v, %v", code, reflect.TypeOf(code))
		if data["status"] == "SUCCESS" && code == 200 {
			log.Debug("<-------- serverLogin SUCCESS~!!! -------->")

			c4c.LoginStat = true

			SendTgMessage("启动成功")

			msginfo := data["msg"].(map[string]interface{})
			log.Debug("globals:%v, %v", msginfo["globals"], reflect.TypeOf(msginfo["globals"]))

			globals := msginfo["globals"].([]interface{})
			for _, v := range globals {
				info := v.(map[string]interface{})
				log.Debug("package_id:%v", info["package_id"])

				var nPackage uint16
				var nTax float64

				jsonPackageId, err := info["package_id"].(json.Number).Int64()
				if err != nil {
					log.Fatal(err.Error())
				} else {
					nPackage = uint16(jsonPackageId)
				}
				jsonTax, err := info["platform_tax_percent"].(json.Number).Float64()

				if err != nil {
					log.Fatal(err.Error())
				} else {
					log.Debug("tax:%v", jsonTax)
					nTax = jsonTax
				}

				SetPackageTaxM(nPackage, nTax)

				log.Debug("packageId:%v,tax:%v", nPackage, nTax)
			}
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
		log.Debug("同步中心服用户登录失败:%v", data)
		return
	}

	if data["status"] == "SUCCESS" && code == 200 {
		log.Debug("<-------- UserLogin SUCCESS~ -------->")
		log.Debug("data:%v,ok:%v", data, ok)

		userInfo, ok := data["msg"].(map[string]interface{})
		var strId string
		var userData *UserCallback
		if ok {
			var lockMoney float64
			gameAccount, aok := userInfo["game_account"].(map[string]interface{})
			if aok {
				jsonMoney := gameAccount["lock_balance"]
				money, err := jsonMoney.(json.Number).Float64()
				if err != nil {
					log.Error(err.Error())
				}
				log.Debug("玩家登入锁金额:%v", money)
				lockMoney = money
			}
			gameUser, uok := userInfo["game_user"].(map[string]interface{})
			if uok {
				nick := gameUser["game_nick"]
				headImg := gameUser["game_img"]
				userId := gameUser["id"]
				packageId := gameUser["package_id"]

				intID, err := userId.(json.Number).Int64()
				if err != nil {
					log.Fatal(err.Error())
				}
				strId = strconv.Itoa(int(intID))

				// 登入存在锁钱将解锁金额
				user, _ := gameHall.UserRecord.Load(strId)
				if user != nil {
					u := user.(*Player)
					if u.IsAction == false {
						if lockMoney > 0 {
							c4c.UnlockSettlement(strId, lockMoney)
						}
					}
				} else {
					if lockMoney > 0 {
						c4c.UnlockSettlement(strId, lockMoney)
					}
				}

				pckId, err2 := packageId.(json.Number).Int64()
				if err2 != nil {
					log.Fatal(err2.Error())
				}

				//找到等待登录玩家
				userData, ok = c4c.waitUser[strId]
				if ok {
					userData.Data.HeadImg = headImg.(string)
					userData.Data.NickName = nick.(string)
					userData.Data.PackageId = uint16(pckId)
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
	log.Debug("data:%v , ok:%v", data, ok)

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
		log.Debug("同步中心服赢钱失败:%v", data)
		return
	}

	if data["status"] == "SUCCESS" && code == 200 {
		log.Debug("<-------- UserWinScore SUCCESS~ -------->")
		log.Debug("data:%+v, ok:%+v", data, ok)

		//将Win数据插入数据
		InsertWinMoney(msgBody)

		userInfo, ok := data["msg"].(map[string]interface{})
		if ok {
			jsonScore := userInfo["final_pay"]
			score, err := jsonScore.(json.Number).Float64()

			log.Debug("同步中心服赢钱成功:%v", score)

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
	msg, ok := data["msg"].(map[string]interface{})
	if ok {
		order := msg["order"]
		if code != 200 {
			log.Debug("同步中心服输钱失败:%v", data)
			v, ok := gameHall.OrderIDRecord.Load(order)
			if ok {
				p := v.(*Player)
				message := fmt.Sprintf("玩家" + p.Id + "输钱失败并登出")
				SendTgMessage(message)
				c4c.UserLogoutCenter(p.Id, p.PassWord, p.Token) //, p.PassWord
				p.IsOnline = false
				p.IsAction = false
				p.PlayerReqExit()
				gameHall.UserRecord.Delete(p.Id)
				leaveHall := &pb_msg.PlayerLeaveHall_S2C{}
				p.SendMsg(leaveHall)
				p.ConnAgent.Close()
				gameHall.OrderIDRecord.Delete(order)
			}
			return
		}
		if data["status"] == "SUCCESS" && code == 200 {
			log.Debug("<-------- UserLoseScore SUCCESS~ -------->")
			log.Debug("data:%+v, ok:%v", data, ok)

			gameHall.OrderIDRecord.Delete(order)

			//将Lose数据插入数据
			InsertLoseMoney(msgBody)

			userInfo, ok := data["msg"].(map[string]interface{})
			if ok {
				jsonScore := userInfo["final_pay"]
				score, err := jsonScore.(json.Number).Float64()

				log.Debug("同步中心服输钱成功:%v", score)

				if err != nil {
					log.Error(err.Error())
					return
				}
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

		msg, ok := data["msg"].(map[string]interface{})
		if ok {
			order := msg["order"]
			if code != 200 {
				log.Debug("锁钱失败:%v", data)
				v, ok := gameHall.OrderIDRecord.Load(order)
				if ok {
					p := v.(*Player)
					p.LockChan <- false
					gameHall.OrderIDRecord.Delete(order)
				}
				return
			}
			if data["status"] == "SUCCESS" && code == 200 {
				log.Debug("<-------- onLockSettlement SUCCESS~!!! -------->")
				v, ok := gameHall.OrderIDRecord.Load(order)
				if ok {
					p := v.(*Player)
					money, _ := msg["lock_money"].(json.Number).Float64()
					p.LockMoney += money
					p.LockChan <- true
					gameHall.OrderIDRecord.Delete(order)
				}
				return
			}
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

		log.Debug("code:%v, %v", code, reflect.TypeOf(code))
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

		log.Debug("code:%v, %v", code, reflect.TypeOf(code))
		if data["status"] == "SUCCESS" && code == 200 {
			log.Debug("<-------- onWinMoreThanNotice SUCCESS~!!! -------->")
		}
	}
}

//ServerLoginCenter 服务器登录Center
func (c4c *Conn4Center) ServerLoginCenter() {
	baseData := &BaseMessage{}
	baseData.Event = msgServerLogin
	port, _ := strconv.Atoi(conf.Server.CenterServerPort)
	baseData.Data = ServerLogin{
		Host:    conf.Server.CenterServer,
		Port:    port,
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
	id, _ := strconv.Atoi(userId)
	baseData := &BaseMessage{}
	baseData.Event = msgUserLogin
	if password != "" {
		baseData.Data = &UserReq{
			ID:       id,
			PassWord: password,
			GameId:   c4c.GameId,
			DevName:  conf.Server.DevName,
			DevKey:   c4c.DevKey}
	} else {
		baseData.Data = &UserReq{
			ID:      id,
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
	id, _ := strconv.Atoi(userId)
	if password != "" {
		base.Data = &UserReq{
			ID:       id,
			PassWord: password,
			GameId:   c4c.GameId,
			DevName:  conf.Server.DevName,
			DevKey:   c4c.DevKey}
	} else {
		base.Data = &UserReq{
			ID:      id,
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

	writeMutex.Lock()
	defer writeMutex.Unlock()
	err2 := c4c.conn.WriteMessage(websocket.TextMessage, []byte(codeData))
	if err2 != nil {
		log.Fatal(err2.Error())
	}
}

//UserSyncWinScore 同步赢分
func (c4c *Conn4Center) UserSyncWinScore(p *Player, timeUnix int64, roundId, reason string, betMoney float64) {
	baseData := &BaseMessage{}
	baseData.Event = msgUserWinScore
	id, _ := strconv.Atoi(p.Id)
	userWin := &UserChangeScore{}
	userWin.Auth.DevName = conf.Server.DevName
	userWin.Auth.DevKey = c4c.DevKey
	userWin.Info.CreateTime = timeUnix
	userWin.Info.GameId = c4c.GameId
	userWin.Info.ID = id
	userWin.Info.LockMoney = 0
	userWin.Info.Money = p.WinResultMoney
	userWin.Info.BetMoney = betMoney
	userWin.Info.Order = bson.NewObjectId().Hex()

	userWin.Info.PayReason = reason
	userWin.Info.PreMoney = 0
	userWin.Info.RoundId = roundId
	baseData.Data = userWin
	c4c.SendMsg2Center(baseData)
}

//UserSyncWinScore 同步输分
func (c4c *Conn4Center) UserSyncLoseScore(p *Player, timeUnix int64, roundId, reason string, betMoney float64) {
	baseData := &BaseMessage{}
	baseData.Event = msgUserLoseScore
	id, _ := strconv.Atoi(p.Id)
	userLose := &UserChangeScore{}
	userLose.Auth.DevName = conf.Server.DevName
	userLose.Auth.DevKey = c4c.DevKey
	userLose.Info.CreateTime = timeUnix
	userLose.Info.GameId = c4c.GameId
	userLose.Info.ID = id
	userLose.Info.LockMoney = 0
	userLose.Info.Money = p.LoseResultMoney
	userLose.Info.BetMoney = betMoney
	userLose.Info.Order = bson.NewObjectId().Hex()
	userLose.Info.PayReason = reason
	userLose.Info.PreMoney = 0
	userLose.Info.RoundId = roundId
	baseData.Data = userLose
	c4c.SendMsg2Center(baseData)
	gameHall.OrderIDRecord.Store(userLose.Info.Order, p)
}

//锁钱
func (c4c *Conn4Center) LockSettlement(p *Player, lockAccount float64) {
	id, _ := strconv.Atoi(p.Id)
	roundId := fmt.Sprintf("%+v-%+v", time.Now().Unix(), p.Id)
	baseData := &BaseMessage{}
	baseData.Event = msgLockSettlement
	lockMoney := &UserChangeScore{}
	lockMoney.Auth.DevName = conf.Server.DevName
	lockMoney.Auth.DevKey = c4c.DevKey
	lockMoney.Info.CreateTime = time.Now().Unix()
	lockMoney.Info.GameId = c4c.GameId
	lockMoney.Info.ID = id
	lockMoney.Info.LockMoney = lockAccount
	lockMoney.Info.Money = 0
	lockMoney.Info.Order = bson.NewObjectId().Hex()
	lockMoney.Info.PayReason = "加锁投注资金"
	lockMoney.Info.PreMoney = 0
	lockMoney.Info.RoundId = roundId
	baseData.Data = lockMoney
	c4c.SendMsg2Center(baseData)
	gameHall.OrderIDRecord.Store(lockMoney.Info.Order, p)
}

//解锁
func (c4c *Conn4Center) UnlockSettlement(Id string, LockMoney float64) {
	id, _ := strconv.Atoi(Id)
	roundId := fmt.Sprintf("%+v-%+v", time.Now().Unix(), Id)
	baseData := &BaseMessage{}
	baseData.Event = msgUnlockSettlement
	lockMoney := &UserChangeScore{}
	lockMoney.Auth.DevName = conf.Server.DevName
	lockMoney.Auth.DevKey = c4c.DevKey
	lockMoney.Info.CreateTime = time.Now().Unix()
	lockMoney.Info.GameId = c4c.GameId
	lockMoney.Info.ID = id
	lockMoney.Info.LockMoney = LockMoney
	lockMoney.Info.Money = 0
	lockMoney.Info.Order = bson.NewObjectId().Hex()
	lockMoney.Info.PayReason = "解锁投注资金"
	lockMoney.Info.PreMoney = 0
	lockMoney.Info.RoundId = roundId
	baseData.Data = lockMoney
	c4c.SendMsg2Center(baseData)
}

func (c4c *Conn4Center) NoticeWinMoreThan(playerId, playerName string, winGold float64) {
	log.Debug("<-------- NoticeWinMoreThan  -------->")
	msg := fmt.Sprintf("<size=20><color=yellow>恭喜!</color><color=orange>%v</color><color=yellow>在</color></><color=orange><size=25>红黑大战</color></><color=yellow><size=20>中一把赢了</color></><color=yellow><size=30>%.2f</color></><color=yellow><size=25>金币！</color></>", playerName, winGold)

	base := &BaseMessage{}
	base.Event = msgWinMoreThanNotice
	id, _ := strconv.Atoi(playerId)
	base.Data = &Notice{
		DevName: conf.Server.DevName,
		DevKey:  conf.Server.DevKey,
		ID:      id,
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
		return nil, err
	}

	//fmt.Printf("读取的奖源池数据: %s", result)

	var users interface{}
	err2 := json.Unmarshal(result, &users)
	if err2 != nil {
		log.Error("解码随机数值失败: %v", err)
		return nil, err
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
