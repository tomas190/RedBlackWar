package internal

import (
	"RedBlack-War/conf"
	pb_msg "RedBlack-War/msg/Protocal"
	"encoding/json"
	"fmt"
	"github.com/name5566/leaf/log"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"strconv"
	"time"
)

type GameDataReq struct {
	Id        string `form:"id" json:"id"`
	GameId    string `form:"game_id" json:"game_id"`
	RoomId    string `form:"room_id" json:"room_id"`
	RoundId   string `form:"round_id" json:"round_id"`
	StartTime string `form:"start_time" json:"start_time"`
	EndTime   string `form:"end_time" json:"end_time"`
	Page      string `form:"page" json:"page"`
	Limit     string `form:"limit" json:"limit"`
}

type ApiResp struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type GameData struct {
	Time            int64       `json:"time"`
	TimeFmt         string      `json:"time_fmt"`
	StartTime       int64       `json:"start_time"`
	EndTime         int64       `json:"end_time"`
	PlayerId        string      `json:"player_id"`
	RoundId         string      `json:"round_id"`
	RoomId          string      `json:"room_id"`
	TaxRate         float64     `json:"tax_rate"`
	Card            interface{} `json:"card"`             // 开牌信息
	BetInfo         interface{} `json:"bet_info"`         // 玩家下注信息
	SettlementFunds interface{} `json:"settlement_funds"` // 结算信息 输赢结果
	SpareCash       interface{} `json:"spare_cash"`       // 剩余金额
	CreatedAt       int64       `json:"created_at"`
}

type pageData struct {
	Total int         `json:"total"`
	List  interface{} `json:"list"`
}

type GetSurPool struct {
	PlayerTotalLose                float64 `json:"player_total_lose" bson:"player_total_lose"`
	PlayerTotalWin                 float64 `json:"player_total_win" bson:"player_total_win"`
	PercentageToTotalWin           float64 `json:"percentage_to_total_win" bson:"percentage_to_total_win"`
	TotalPlayer                    int32   `json:"total_player" bson:"total_player"`
	CoefficientToTotalPlayer       int32   `json:"coefficient_to_total_player" bson:"coefficient_to_total_player"`
	FinalPercentage                float64 `json:"final_percentage" bson:"final_percentage"`
	PlayerTotalLoseWin             float64 `json:"player_total_lose_win" bson:"player_total_lose_win" `
	SurplusPool                    float64 `json:"surplus_pool" bson:"surplus_pool"`
	PlayerLoseRateAfterSurplusPool float64 `json:"player_lose_rate_after_surplus_pool" bson:"player_lose_rate_after_surplus_pool"`
	DataCorrection                 float64 `json:"data_correction" bson:"data_correction"`
	PlayerWinRate                  float64 `json:"player_win_rate" bson:"player_win_rate"`
	RandomCountAfterWin            float64 `json:"random_count_after_win" bson:"random_count_after_win"`
	RandomCountAfterLose           float64 `json:"random_count_after_lose" bson:"random_count_after_lose"`
	RandomPercentageAfterWin       float64 `json:"random_percentage_after_win" bson:"random_percentage_after_win"`
	RandomPercentageAfterLose      float64 `json:"random_percentage_after_lose" bson:"random_percentage_after_lose"`
}

type UpSurPool struct {
	PlayerLoseRateAfterSurplusPool float64 `json:"player_lose_rate_after_surplus_pool" bson:"player_lose_rate_after_surplus_pool"`
	PercentageToTotalWin           float64 `json:"percentage_to_total_win" bson:"percentage_to_total_win"`
	CoefficientToTotalPlayer       int32   `json:"coefficient_to_total_player" bson:"coefficient_to_total_player"`
	FinalPercentage                float64 `json:"final_percentage" bson:"final_percentage"`
	DataCorrection                 float64 `json:"data_correction" bson:"data_correction"`
	PlayerWinRate                  float64 `json:"player_win_rate" bson:"player_win_rate"`
	RandomCountAfterWin            float64 `json:"random_count_after_win" bson:"random_count_after_win"`
	RandomCountAfterLose           float64 `json:"random_count_after_lose" bson:"random_count_after_lose"`
	RandomPercentageAfterWin       float64 `json:"random_percentage_after_win" bson:"random_percentage_after_win"`
	RandomPercentageAfterLose      float64 `json:"random_percentage_after_lose" bson:"random_percentage_after_lose"`
}

type GRobotData struct {
	RoomId   string       `json:"room_id" bson:"room_id"`
	RoomTime int64        `json:"room_time" bson:"room_time"`
	RobotNum int          `json:"robot_num" bson:"robot_num"`
	RedPot   *ChipDownBet `json:"red_pot" bson:"red_pot"`
	BlackPot *ChipDownBet `json:"black_pot" bson:"black_pot"`
	LuckPot  *ChipDownBet `json:"luck_pot" bson:"luck_pot"`
}

type StatementReq struct {
	Id        string `form:"id" json:"id"`
	StartTime string `form:"start_time" json:"start_time"`
	EndTime   string `form:"end_time" json:"end_time"`
	PackageId string `form:"package_id" json:"package_id"`
}

type StatementResp struct {
	GameId             string  `json:"game_id" bson:"game_id"`
	GameName           string  `json:"game_name" bson:"game_name"`
	WinStatementTotal  float64 `json:"win_statement_total" bson:"win_statement_total"`
	LoseStatementTotal float64 `json:"lose_statement_total" bson:"lose_statement_total"`
	BetMoney           float64 `json:"bet_money" bson:"bet_money"`
	Count              []int   `json:"count" json:"count"`
}

type OnlineTotal struct {
	GameId   string          `json:"game_id" bson:"game_id"`
	GameName string          `json:"game_name" bson:"game_name"`
	GameData []*OnlinePlayer `json:"game_data" bson:"game_data"`
}

type OnlinePlayer struct {
	PackageId uint16 `json:"packageID" bson:"packageID"`
	UserData  []int  `json:"userData" bson:"userData"`
}

const (
	SuccCode = 0
	ErrCode  = -1
)

// HTTP端口监听
func StartHttpServer() {
	// 运营后台数据接口
	http.HandleFunc("/api/accessData", getAccessData)
	// 获取游戏数据接口
	http.HandleFunc("/api/getGameData", getAccessData)
	// 请求玩家退出
	http.HandleFunc("/api/reqPlayerLeave", reqPlayerLeave)
	// 解锁金额
	http.HandleFunc("/api/unLockUserMoney", unLockUserMoney)
	// 查询子游戏盈余池数据
	http.HandleFunc("/api/getSurplusOne", getSurplusOne)
	// 修改盈余池数据
	http.HandleFunc("/api/uptSurplusConf", uptSurplusOne)
	// 获取机器人数据
	http.HandleFunc("/api/getRobotData", getRobotData)
	// 获取游戏统计数据接口
	http.HandleFunc("/api/getStatementTotal", getStatementTotal)
	// 获取实时在线人数
	http.HandleFunc("/api/getOnlineTotal", getOnlineTotal)
	// 设定特殊品牌
	http.HandleFunc("/api/SetSpecialPackageId", SetSpecialPackageId)

	err := http.ListenAndServe(":"+conf.Server.HTTPPort, nil)
	if err != nil {
		log.Error("Http server启动异常:", err.Error())
		panic(err)
	}
}

func getAccessData(w http.ResponseWriter, r *http.Request) {
	var req GameDataReq

	req.Id = r.FormValue("id")
	req.GameId = r.FormValue("game_id")
	req.RoomId = r.FormValue("room_id")
	req.RoundId = r.FormValue("round_id")
	req.StartTime = r.FormValue("start_time")
	req.EndTime = r.FormValue("end_time")
	req.Page = r.FormValue("page")
	req.Limit = r.FormValue("limit")
	log.Debug("获取分页数据:%v", req.Page)

	selector := bson.M{}

	if req.Id != "" {
		selector["id"] = req.Id
	}

	if req.GameId != "" {
		selector["game_id"] = req.GameId
	}

	if req.RoomId != "" {
		selector["room_id"] = req.RoomId
	}

	if req.RoundId != "" {
		selector["round_id"] = req.RoundId
	}

	sTime, _ := strconv.Atoi(req.StartTime)

	eTime, _ := strconv.Atoi(req.EndTime)

	if sTime != 0 && eTime != 0 {
		selector["down_bet_time"] = bson.M{"$gte": sTime, "$lte": eTime}
	}

	if sTime != 0 && eTime == 0 {
		selector["start_time"] = bson.M{"$gte": sTime}
	}

	if eTime != 0 && sTime == 0 {
		selector["end_time"] = bson.M{"$lte": eTime}
	}

	page, _ := strconv.Atoi(req.Page)

	limits, _ := strconv.Atoi(req.Limit)
	//if limits != 0 {
	//	selector["limit"] = limits
	//}

	recodes, count, err := GetDownRecodeList(page, limits, selector, "-down_bet_time")
	if err != nil {
		return
	}

	var gameData []GameData
	for i := 0; i < len(recodes); i++ {
		var gd GameData
		pr := recodes[i]
		gd.Time = pr.DownBetTime
		gd.TimeFmt = FormatTime(pr.DownBetTime, "2006-01-02 15:04:05")
		gd.StartTime = pr.StartTime
		gd.EndTime = pr.EndTime
		gd.PlayerId = pr.Id
		gd.RoomId = pr.RoomId
		gd.RoundId = pr.RoundId
		gd.BetInfo = *pr.DownBetInfo
		gd.Card = *pr.CardResult
		gd.SettlementFunds = pr.SettlementFunds
		gd.SpareCash = pr.SpareCash
		gd.TaxRate = pr.TaxRate
		gd.CreatedAt = pr.DownBetTime
		gameData = append(gameData, gd)
	}

	var result pageData
	result.Total = count
	result.List = gameData

	//fmt.Fprintf(w, "%+v", ApiResp{Code: SuccCode, Msg: "", Data: result})
	js, err := json.Marshal(NewResp(SuccCode, "", result))
	if err != nil {
		fmt.Fprintf(w, "%+v", ApiResp{Code: ErrCode, Msg: "", Data: nil})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func FormatTime(timeUnix int64, layout string) string {
	if timeUnix == 0 {
		return ""
	}
	format := time.Unix(timeUnix, 0).Format(layout)
	return format
}

func NewResp(code int, msg string, data interface{}) ApiResp {
	return ApiResp{Code: code, Msg: msg, Data: data}
}

func reqPlayerLeave(w http.ResponseWriter, r *http.Request) {
	Id := r.FormValue("id")
	user, _ := gameHall.UserRecord.Load(Id)
	if user != nil {
		p := user.(*Player)
		c4c.UserLogoutCenter(p.Id, p.PassWord, p.Token) //, p.PassWord
		p.IsOnline = false
		p.IsAction = false
		p.PlayerReqExit()
		gameHall.UserRecord.Delete(p.Id)
		leaveHall := &pb_msg.PlayerLeaveHall_S2C{}
		p.SendMsg(leaveHall)
		p.ConnAgent.Close()

		js, err := json.Marshal(NewResp(SuccCode, "", "已成功T出房间!"))
		if err != nil {
			fmt.Fprintf(w, "%+v", ApiResp{Code: ErrCode, Msg: "", Data: nil})
			return
		}
		w.Write(js)
	} else {
		c4c.UserLogoutCenter(Id, "", "") //, p.PassWord
	}
}

func unLockUserMoney(w http.ResponseWriter, r *http.Request) {
	Id := r.FormValue("id")
	lockMoney := r.FormValue("lock_money")
	log.Debug("unLockUserMoney 解锁金额:%v,%v", Id, lockMoney)
	money, _ := strconv.ParseFloat(lockMoney, 64)
	c4c.UnlockSettlement(Id, money)

	js, err := json.Marshal(NewResp(SuccCode, "", "操作解锁玩家金额!"))
	if err != nil {
		fmt.Fprintf(w, "%+v", ApiResp{Code: ErrCode, Msg: "", Data: nil})
		return
	}
	w.Write(js)
}

// 查询子游戏盈余池数据
func getSurplusOne(w http.ResponseWriter, r *http.Request) {
	var req GameDataReq
	req.GameId = r.FormValue("game_id")
	log.Debug("game_id :%v", req.GameId)

	selector := bson.M{}
	if req.GameId != "" {
		selector["game_id"] = req.GameId
	}

	result, err := GetSurPoolData(selector)
	if err != nil {
		return
	}

	var getSur GetSurPool
	getSur.PlayerTotalLose = result.PlayerTotalLose
	getSur.PlayerTotalWin = result.PlayerTotalWin
	getSur.PercentageToTotalWin = result.PercentageToTotalWin
	getSur.TotalPlayer = result.TotalPlayer
	getSur.CoefficientToTotalPlayer = result.CoefficientToTotalPlayer
	getSur.FinalPercentage = result.FinalPercentage
	getSur.PlayerTotalLoseWin = result.PlayerTotalLoseWin
	getSur.SurplusPool = result.SurplusPool
	getSur.PlayerLoseRateAfterSurplusPool = result.PlayerLoseRateAfterSurplusPool
	getSur.DataCorrection = result.DataCorrection
	getSur.PlayerWinRate = result.PlayerWinRate
	getSur.RandomCountAfterWin = result.RandomCountAfterWin
	getSur.RandomCountAfterLose = result.RandomCountAfterLose
	getSur.RandomPercentageAfterWin = result.RandomPercentageAfterWin
	getSur.RandomPercentageAfterLose = result.RandomPercentageAfterLose

	js, err := json.Marshal(NewResp(SuccCode, "", getSur))
	if err != nil {
		fmt.Fprintf(w, "%+v", ApiResp{Code: ErrCode, Msg: "", Data: nil})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func uptSurplusOne(w http.ResponseWriter, r *http.Request) {
	rateSur := r.PostFormValue("player_lose_rate_after_surplus_pool")
	percentage := r.PostFormValue("percentage_to_total_win")
	coefficient := r.PostFormValue("coefficient_to_total_player")
	final := r.PostFormValue("final_percentage")
	correction := r.PostFormValue("data_correction")
	winRate := r.PostFormValue("player_win_rate")
	countWin := r.PostFormValue("random_count_after_win")
	countLose := r.PostFormValue("random_count_after_lose")
	percentageWin := r.PostFormValue("random_percentage_after_win")
	percentageLose := r.PostFormValue("random_percentage_after_lose")

	s, c := connect(dbName, surPool)
	defer s.Close()

	sur := &SurPool{}
	err := c.Find(nil).One(sur)
	if err != nil {
		log.Debug("uptSurplusOne 盈余池赋值失败~")
	}

	var upt UpSurPool
	upt.PlayerLoseRateAfterSurplusPool = sur.PlayerLoseRateAfterSurplusPool
	upt.PercentageToTotalWin = sur.PercentageToTotalWin
	upt.CoefficientToTotalPlayer = sur.CoefficientToTotalPlayer
	upt.FinalPercentage = sur.FinalPercentage
	upt.DataCorrection = sur.DataCorrection
	upt.PlayerWinRate = sur.PlayerWinRate
	upt.RandomCountAfterWin = sur.RandomCountAfterWin
	upt.RandomCountAfterLose = sur.RandomCountAfterLose
	upt.RandomPercentageAfterWin = sur.RandomPercentageAfterWin
	upt.RandomPercentageAfterLose = sur.RandomPercentageAfterLose

	if rateSur != "" {
		upt.PlayerLoseRateAfterSurplusPool, _ = strconv.ParseFloat(rateSur, 64)
		sur.PlayerLoseRateAfterSurplusPool = upt.PlayerLoseRateAfterSurplusPool
	}
	if percentage != "" {
		upt.PercentageToTotalWin, _ = strconv.ParseFloat(percentage, 64)
		sur.PercentageToTotalWin = upt.PercentageToTotalWin
	}
	if coefficient != "" {
		data, _ := strconv.ParseInt(coefficient, 10, 32)
		upt.CoefficientToTotalPlayer = int32(data)
		sur.CoefficientToTotalPlayer = upt.CoefficientToTotalPlayer
	}
	if final != "" {
		upt.FinalPercentage, _ = strconv.ParseFloat(final, 64)
		sur.FinalPercentage = upt.FinalPercentage
	}
	if correction != "" {
		upt.DataCorrection, _ = strconv.ParseFloat(correction, 64)
		sur.DataCorrection = upt.DataCorrection
	}
	if winRate != "" {
		upt.PlayerWinRate, _ = strconv.ParseFloat(winRate, 64)
		sur.PlayerWinRate = upt.PlayerWinRate
	}
	if countWin != "" {
		upt.RandomCountAfterWin, _ = strconv.ParseFloat(countWin, 64)
		sur.RandomCountAfterWin = upt.RandomCountAfterWin
	}
	if countLose != "" {
		upt.RandomCountAfterLose, _ = strconv.ParseFloat(countLose, 64)
		sur.RandomCountAfterLose = upt.RandomCountAfterLose
	}
	if percentageWin != "" {
		upt.RandomPercentageAfterWin, _ = strconv.ParseFloat(percentageWin, 64)
		sur.RandomPercentageAfterWin = upt.RandomPercentageAfterWin
	}
	if percentageLose != "" {
		upt.RandomPercentageAfterLose, _ = strconv.ParseFloat(percentageLose, 64)
		sur.RandomPercentageAfterLose = upt.RandomPercentageAfterLose
	}

	sur.SurplusPool = (sur.PlayerTotalLose - (sur.PlayerTotalWin * sur.PercentageToTotalWin) - float64(sur.TotalPlayer*sur.CoefficientToTotalPlayer) + sur.DataCorrection) * sur.FinalPercentage
	// 更新盈余池数据
	UpdateSurPool(sur)

	js, err := json.Marshal(NewResp(SuccCode, "", upt))
	if err != nil {
		fmt.Fprintf(w, "%+v", ApiResp{Code: ErrCode, Msg: "", Data: nil})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func getRobotData(w http.ResponseWriter, r *http.Request) {
	recodes, err := GetRobotData()
	if err != nil {
		return
	}

	var rData []GRobotData
	for i := 0; i < len(recodes); i++ {
		var rd GRobotData
		rd.RedPot = new(ChipDownBet)
		rd.BlackPot = new(ChipDownBet)
		rd.LuckPot = new(ChipDownBet)
		pr := recodes[i]
		log.Debug("获取机器数据:%v", pr)
		rd.RoomId = pr.RoomId
		rd.RoomTime = pr.RoomTime
		rd.RobotNum = pr.RobotNum
		rd.RedPot = pr.RedPot
		rd.BlackPot = pr.BlackPot
		rd.LuckPot = pr.LuckPot
		rData = append(rData, rd)
	}

	js, err := json.Marshal(NewResp(SuccCode, "", rData))
	if err != nil {
		fmt.Fprintf(w, "%+v", ApiResp{Code: ErrCode, Msg: "", Data: nil})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func getStatementTotal(w http.ResponseWriter, r *http.Request) {
	var req StatementReq

	req.Id = r.FormValue("id")
	req.PackageId = r.FormValue("package_id")
	req.StartTime = r.FormValue("start_time")
	req.EndTime = r.FormValue("end_time")

	log.Debug("获取游戏数据:%v", req)

	selector := bson.M{}
	if req.Id != "" {
		selector["id"] = req.Id
	}
	packId, _ := strconv.Atoi(req.PackageId)
	if req.PackageId != "" {
		selector["package_id"] = uint16(packId)
	}

	sTime, _ := strconv.Atoi(req.StartTime)
	eTime, _ := strconv.Atoi(req.EndTime)

	if sTime != 0 && eTime != 0 {
		selector["down_bet_time"] = bson.M{"$gte": sTime, "$lte": eTime}
	}
	if sTime != 0 && eTime == 0 {
		selector["start_time"] = bson.M{"$gte": sTime}
	}
	if eTime != 0 && sTime == 0 {
		selector["end_time"] = bson.M{"$lte": eTime}
	}

	recodes, _ := GetStatementList(selector)
	data := &StatementResp{}
	for _, v := range recodes {
		data.GameId = v.GameId
		data.GameName = v.GameName
		data.WinStatementTotal += v.WinStatementTotal
		data.LoseStatementTotal += v.LoseStatementTotal
		data.BetMoney += float64(v.BetMoney)
		id, _ := strconv.Atoi(v.Id)
		data.Count = append(data.Count, id)
	}
	data.Count = uniqueArr(data.Count)

	js, err := json.Marshal(NewResp(SuccCode, "", data))
	if err != nil {
		fmt.Fprintf(w, "%+v", ApiResp{Code: ErrCode, Msg: "", Data: nil})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func getOnlineTotal(w http.ResponseWriter, r *http.Request) {
	packageId := r.FormValue("package_id")
	log.Debug("获取getOnlineTotal接口 packageId:%v", packageId)

	total := &OnlineTotal{}
	total.GameId = conf.Server.GameID
	total.GameName = "红黑大战"
	total.GameData = []*OnlinePlayer{}
	if packageId == "" {
		packageIds := make([]uint16, 0)
		gameHall.UserRecord.Range(func(key, value interface{}) bool {
			user := value.(*Player)
			packageIds = append(packageIds, user.PackageId)
			return true
		})
		packageIds = removeDuplicateElement(packageIds)

		for _, v := range packageIds {
			data := &OnlinePlayer{}
			data.PackageId = v
			gameHall.UserRecord.Range(func(key, value interface{}) bool {
				user := value.(*Player)
				if v == user.PackageId {
					id, _ := strconv.Atoi(user.Id)
					data.UserData = append(data.UserData, id)
				}
				return true
			})
			total.GameData = append(total.GameData, data)
		}
	} else {
		packId, _ := strconv.Atoi(packageId)
		data := &OnlinePlayer{}
		data.PackageId = uint16(packId)
		gameHall.UserRecord.Range(func(key, value interface{}) bool {
			user := value.(*Player)
			if user.PackageId == uint16(packId) {
				id, _ := strconv.Atoi(user.Id)
				data.UserData = append(data.UserData, id)
				log.Debug("获取玩家信息:%v", user)
			}
			return true
		})
		if data.UserData != nil {
			total.GameData = append(total.GameData, data)
		}
	}

	js, err := json.Marshal(NewResp(SuccCode, "", total))
	if err != nil {
		fmt.Fprintf(w, "%+v", ApiResp{Code: ErrCode, Msg: "", Data: nil})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func SetSpecialPackageId(w http.ResponseWriter, r *http.Request) {
	packageId := r.FormValue("package_id")
	special := r.FormValue("special")
	log.Debug("SetSpecialPackageId 设定特殊品牌 packageId:%v,special:%v", packageId, special)

	if packageId != "" && special != "" {
		pack, _ := strconv.Atoi(packageId)
		gameHall.RoomRecord.Range(func(key, value interface{}) bool {
			room := value.(*Room)
			if room.PackageId == uint16(pack) {
				if special == "true" {
					room.IsSpecial = true
				} else if special == "false" {
					room.IsSpecial = false
				}
			}
			return true
		})

		js, err := json.Marshal(NewResp(SuccCode, "", "设定特殊品牌"))
		if err != nil {
			fmt.Fprintf(w, "%+v", ApiResp{Code: ErrCode, Msg: "", Data: nil})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	}
}

func removeDuplicateElement(languages []uint16) []uint16 {
	result := make([]uint16, 0, len(languages))
	temp := map[uint16]struct{}{}
	for _, item := range languages {
		if _, ok := temp[item]; !ok {
			temp[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

func uniqueArr(m []int) []int {
	d := make([]int, 0)
	tempMap := make(map[int]bool, len(m))
	for _, v := range m { // 以值作为键名
		if tempMap[v] == false {
			tempMap[v] = true
			d = append(d, v)
		}
	}
	return d
}
