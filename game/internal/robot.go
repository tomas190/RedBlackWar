package internal

import (
	pb_msg "RedBlack-War/msg/Protocal"
	"fmt"
	"github.com/name5566/leaf/log"
	"math/rand"
	"time"
)

//机器人问题:
//1、机器人没钱怎么充值,不能再房间就直接充值,不然可以被其他用户看见
//2、机器人怎么下注，如果在桌面6个位置上，是否设置机器的下注速度和选择注池
//3、机器人选择注池的输赢,都要进行计算，只是不和盈余池牵扯，主要是前端做展示
//4、如果机器人金额如果小于50或不能参加游戏,则踢出房间删除机器人，在生成新的机器人加入该房间。

//机器人下标
var RobotIndex uint32

//Init 初始机器人控制中心
func (rc *RobotsCenter) Init() {
	log.Debug("-------------- RobotsCenter Init~! ---------------")
	rc.mapRobotList = make(map[uint32]*Player)
}

//CreateRobot 创建一个机器人
func (rc *RobotsCenter) CreateRobot() *Player {
	r := &Player{}
	r.Init()

	r.IsRobot = true
	//生成随机ID
	r.Id = RandomID()
	//生成随机头像IMG
	r.HeadImg = RandomIMG()
	//生成随机机器人NickName
	r.NickName = RandomName()
	//生成机器人金币随机数
	r.Account = RandomAccount()

	r.Index = RobotIndex
	//fmt.Println("robot Index :", r.Index)
	RobotIndex++
	//log.Debug("创建机器人~ : %v", r.Id)
	return r
}

//RobotsDownBet 机器人进行下注
func (r *Room) RobotsDownBet() {
	var robotSlice []*Player
	for _, v := range r.PlayerList {
		if v != nil && v.IsRobot == true {
			robotSlice = append(robotSlice, v)
		}
	}
	// 线程下注
	go func() {
		time.Sleep(time.Second)
		rData := &RobotDATA{}
		rData.RoomId = r.RoomId
		rData.RoomTime = time.Now().Unix()
		rData.RobotNum = r.RobotLength()
		rData.RedPot = new(ChipDownBet)
		rData.BlackPot = new(ChipDownBet)
		rData.LuckPot = new(ChipDownBet)
		for {
			for _, v := range r.PlayerList {
				if v != nil && v.IsRobot == true {

					timerSlice := []int32{50, 150, 20, 300, 30, 500}
					rand.Seed(time.Now().UnixNano())
					num2 := rand.Intn(len(timerSlice))
					time.Sleep(time.Millisecond * time.Duration(timerSlice[num2]))

					var bet1 int32
					if r.GameStat == DownBet {
						pot1 := RobotRandPot(v, r)
						if v.Id == r.GodGambleName {
							if pot1 != 3 {
								slice := make([]int32, 0)
								if v.DownBetMoneys.RedDownBet != 0 {
									slice = append(slice, 1)
									rand.Seed(time.Now().UnixNano())
									n := rand.Intn(len(slice))
									pot1 = slice[n]
								}
								if v.DownBetMoneys.BlackDownBet != 0 {
									slice = append(slice, 2)
									rand.Seed(time.Now().UnixNano())
									n := rand.Intn(len(slice))
									pot1 = slice[n]
								}
							}
						}

						if v.DownBetMoneys.RedDownBet+v.DownBetMoneys.BlackDownBet+v.DownBetMoneys.LuckDownBet >= 1000 {
							continue
						}
						bet1 = RobotRandBet()

						v.IsAction = true

						if bet1 < 1000 {
							randNum := RandInRange(1, 10)
							for i := 0; i < randNum; i++ {
								time.Sleep(time.Millisecond * 10)
								if v.Account < float64(bet1) {
									//log.Debug("机器人:%v 下注金额小于身上筹码,下注失败~", v.Id)
									continue
								}

								//判断玩家下注金额是否限红1-20000
								if pot1 == int32(pb_msg.PotType_RedPot) {
									if (r.PotMoneyCount.RedMoneyCount+bet1)+(r.PotMoneyCount.LuckMoneyCount*10)-r.PotMoneyCount.BlackMoneyCount > 20000 {
										continue
									}
								}
								if pot1 == int32(pb_msg.PotType_BlackPot) {
									if (r.PotMoneyCount.BlackMoneyCount+bet1)+(r.PotMoneyCount.LuckMoneyCount*10)-r.PotMoneyCount.RedMoneyCount > 20000 {
										continue
									}
								}
								if pot1 == int32(pb_msg.PotType_LuckPot) {
									if r.PotMoneyCount.RedMoneyCount+((r.PotMoneyCount.LuckMoneyCount+bet1)*10)-r.PotMoneyCount.BlackMoneyCount > 20000 {
										continue
									}
									if r.PotMoneyCount.BlackMoneyCount+((r.PotMoneyCount.LuckMoneyCount+bet1)*10)-r.PotMoneyCount.RedMoneyCount > 20000 {
										continue
									}
								}

								//记录玩家在该房间总下注 和 房间注池的总金额
								if pb_msg.PotType(pot1) == pb_msg.PotType_RedPot {
									v.Account -= float64(bet1)
									v.DownBetMoneys.RedDownBet += bet1
									v.TotalAmountBet += bet1
									r.PotMoneyCount.RedMoneyCount += bet1
									if bet1 == 1 {
										rData.RedPot.Chip1 += 1
									} else if bet1 == 10 {
										rData.RedPot.Chip10 += 1
									} else if bet1 == 50 {
										rData.RedPot.Chip50 += 1
									} else if bet1 == 100 {
										rData.RedPot.Chip100 += 1
									} else if bet1 == 1000 {
										rData.RedPot.Chip1000 += 1
									}
								}
								if pb_msg.PotType(pot1) == pb_msg.PotType_BlackPot {
									v.Account -= float64(bet1)
									v.DownBetMoneys.BlackDownBet += bet1
									v.TotalAmountBet += bet1
									r.PotMoneyCount.BlackMoneyCount += bet1
									if bet1 == 1 {
										rData.BlackPot.Chip1 += 1
									} else if bet1 == 10 {
										rData.BlackPot.Chip10 += 1
									} else if bet1 == 50 {
										rData.BlackPot.Chip50 += 1
									} else if bet1 == 100 {
										rData.BlackPot.Chip100 += 1
									} else if bet1 == 1000 {
										rData.BlackPot.Chip1000 += 1
									}
								}
								if pb_msg.PotType(pot1) == pb_msg.PotType_LuckPot {
									v.Account -= float64(bet1)
									v.DownBetMoneys.LuckDownBet += bet1
									v.TotalAmountBet += bet1
									r.PotMoneyCount.LuckMoneyCount += bet1
									if bet1 == 1 {
										rData.LuckPot.Chip1 += 1
									} else if bet1 == 10 {
										rData.LuckPot.Chip10 += 1
									} else if bet1 == 50 {
										rData.LuckPot.Chip50 += 1
									} else if bet1 == 100 {
										rData.LuckPot.Chip100 += 1
									} else if bet1 == 1000 {
										rData.LuckPot.Chip1000 += 1
									}
								}
								//返回前端玩家行动,更新玩家最新金额
								action := &pb_msg.PlayerAction_S2C{}
								action.Id = v.Id
								action.DownBet = bet1
								action.DownPot = pb_msg.PotType(pot1)
								action.IsAction = v.IsAction
								action.Account = v.Account
								r.BroadCastMsg(action)

								//广播玩家注池金额
								pot := &pb_msg.PotTotalMoney_S2C{}
								pot.PotMoneyCount = new(pb_msg.PotMoneyCount)
								pot.PotMoneyCount.RedMoneyCount = r.PotMoneyCount.RedMoneyCount
								pot.PotMoneyCount.BlackMoneyCount = r.PotMoneyCount.BlackMoneyCount
								pot.PotMoneyCount.LuckMoneyCount = r.PotMoneyCount.LuckMoneyCount
								r.BroadCastMsg(pot)
							}
						} else {
							if v.Account < float64(bet1) {
								//log.Debug("机器人:%v 下注金额小于身上筹码,下注失败~", v.Id)
								continue
							}

							//判断玩家下注金额是否限红1-20000
							if pot1 == int32(pb_msg.PotType_RedPot) {
								if (r.PotMoneyCount.RedMoneyCount+bet1)+(r.PotMoneyCount.LuckMoneyCount*10)-r.PotMoneyCount.BlackMoneyCount > 20000 {
									continue
								}
							}
							if pot1 == int32(pb_msg.PotType_BlackPot) {
								if (r.PotMoneyCount.BlackMoneyCount+bet1)+(r.PotMoneyCount.LuckMoneyCount*10)-r.PotMoneyCount.RedMoneyCount > 20000 {
									continue
								}
							}
							if pot1 == int32(pb_msg.PotType_LuckPot) {
								if r.PotMoneyCount.RedMoneyCount+((r.PotMoneyCount.LuckMoneyCount+bet1)*10)-r.PotMoneyCount.BlackMoneyCount > 20000 {
									continue
								}
								if r.PotMoneyCount.BlackMoneyCount+((r.PotMoneyCount.LuckMoneyCount+bet1)*10)-r.PotMoneyCount.RedMoneyCount > 20000 {
									continue
								}
							}

							//记录玩家在该房间总下注 和 房间注池的总金额
							if pb_msg.PotType(pot1) == pb_msg.PotType_RedPot {
								v.Account -= float64(bet1)
								v.DownBetMoneys.RedDownBet += bet1
								v.TotalAmountBet += bet1
								r.PotMoneyCount.RedMoneyCount += bet1
								if bet1 == 1 {
									rData.RedPot.Chip1 += 1
								} else if bet1 == 10 {
									rData.RedPot.Chip10 += 1
								} else if bet1 == 50 {
									rData.RedPot.Chip50 += 1
								} else if bet1 == 100 {
									rData.RedPot.Chip100 += 1
								} else if bet1 == 1000 {
									rData.RedPot.Chip1000 += 1
								}
							}
							if pb_msg.PotType(pot1) == pb_msg.PotType_BlackPot {
								v.Account -= float64(bet1)
								v.DownBetMoneys.BlackDownBet += bet1
								v.TotalAmountBet += bet1
								r.PotMoneyCount.BlackMoneyCount += bet1
								if bet1 == 1 {
									rData.BlackPot.Chip1 += 1
								} else if bet1 == 10 {
									rData.BlackPot.Chip10 += 1
								} else if bet1 == 50 {
									rData.BlackPot.Chip50 += 1
								} else if bet1 == 100 {
									rData.BlackPot.Chip100 += 1
								} else if bet1 == 1000 {
									rData.BlackPot.Chip1000 += 1
								}
							}
							if pb_msg.PotType(pot1) == pb_msg.PotType_LuckPot {
								v.Account -= float64(bet1)
								v.DownBetMoneys.LuckDownBet += bet1
								v.TotalAmountBet += bet1
								r.PotMoneyCount.LuckMoneyCount += bet1
								if bet1 == 1 {
									rData.LuckPot.Chip1 += 1
								} else if bet1 == 10 {
									rData.LuckPot.Chip10 += 1
								} else if bet1 == 50 {
									rData.LuckPot.Chip50 += 1
								} else if bet1 == 100 {
									rData.LuckPot.Chip100 += 1
								} else if bet1 == 1000 {
									rData.LuckPot.Chip1000 += 1
								}
							}
							//返回前端玩家行动,更新玩家最新金额
							action := &pb_msg.PlayerAction_S2C{}
							action.Id = v.Id
							action.DownBet = bet1
							action.DownPot = pb_msg.PotType(pot1)
							action.IsAction = v.IsAction
							action.Account = v.Account
							r.BroadCastMsg(action)

							//广播玩家注池金额
							pot := &pb_msg.PotTotalMoney_S2C{}
							pot.PotMoneyCount = new(pb_msg.PotMoneyCount)
							pot.PotMoneyCount.RedMoneyCount = r.PotMoneyCount.RedMoneyCount
							pot.PotMoneyCount.BlackMoneyCount = r.PotMoneyCount.BlackMoneyCount
							pot.PotMoneyCount.LuckMoneyCount = r.PotMoneyCount.LuckMoneyCount
							r.BroadCastMsg(pot)
						}
						//fmt.Println("玩家:", v.Id, "行动 红、黑、Luck下注: ", v.DownBetMoneys, "玩家总下注金额: ", v.TotalAmountBet)
					} else {
						InsertRobotData(rData)
						return
					}
				}
			}
		}
	}()
}

//RandNumber 随机机器下注金额
func RobotRandBet() int32 {
	//slice := []int32{1, 10, 50, 100}
	var slice []int32
	slice = []int32{1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
		1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
		1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
		1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
		1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
		1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
		1, 1, 1, 10, 10, 10, 10, 10, 10, 10,
		10, 10, 10, 10, 10, 10, 10, 10, 10, 10,
		10, 10, 10, 50, 50, 50, 50, 50, 50, 50,
		50, 50, 100, 100, 100, 100, 100, 100, 1000, 1000,
	}

	rand.Seed(time.Now().UnixNano())
	num := rand.Intn(len(slice))
	return slice[num]
}

//RandNumber 随机机器下注金额
func RobotRandPot(p *Player, r *Room) int32 {
	//设置赌神随机只能下 红、Luck 或者 黑、Luck池
	if p.Id == r.GodGambleName {
		slice := make([]int32, 0)
		if p.DownBetMoneys.RedDownBet != 0 {
			slice = []int32{1, 3}
			rand.Seed(time.Now().UnixNano())
			n := rand.Intn(len(slice))
			return slice[n]
		}
		if p.DownBetMoneys.BlackDownBet != 0 {
			slice = []int32{2, 3}
			rand.Seed(time.Now().UnixNano())
			n := rand.Intn(len(slice))
			return slice[n]
		}
		randSlice := []int32{1, 2, 3}
		slice = append(slice, randSlice...)
		rand.Seed(time.Now().UnixNano())
		n2 := rand.Intn(len(randSlice))
		return slice[n2]
	}
	slice2 := []int32{1, 1, 1, 1, 2, 2, 2, 2, 3, 3} //1, 2, 1, 2, 1, 2, 3, 1, 2, 1, 2
	rand.Seed(time.Now().UnixNano())
	n3 := rand.Intn(len(slice2))
	return slice2[n3]
}

//Start 机器人开工~！
func (rc *RobotsCenter) Start() {
	rand.Seed(time.Now().UnixNano())
	num := RandInRange(15, 25)
	gameHall.LoadHallRobots(num)
}

//生成随机机器人ID
func RandomID() string {
	for {
		RobotId := fmt.Sprintf("%09v", rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(800000000))
		if RobotId[0:1] != "0" {
			return RobotId
		}
	}
}

//生成随机机器人头像IMG
func RandomIMG() string {
	slice := []string{
		"1.png", "2.png", "3.png", "4.png", "5.png", "6.png", "7.png", "8.png", "9.png", "10.png",
		"11.png", "12.png", "13.png", "14.png", "15.png", "16.png", "17.png", "18.png", "19.png", "20.png",
	}
	rand.Seed(int64(time.Now().UnixNano()))
	num := rand.Intn(len(slice))

	return slice[num]
}

func RandomAccount() float64 {
	rand.Intn(int(time.Now().Unix()))
	money := RandInRange(200, 5000)
	return float64(money)
}

//生成随机机器人NickName
func RandomName() string {
	for {
		randNum := fmt.Sprintf("%09v", rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(800000000))
		if randNum[0:1] != "0" {
			return randNum
		}
	}
}

func RandInRange(min int, max int) int {
	rand.Seed(time.Now().UnixNano())
	time.Sleep(1 * time.Nanosecond)
	return rand.Intn(max-min) + min
}
