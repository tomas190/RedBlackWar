package internal

import (
	"RedBlack-War/base"
	"fmt"
	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/module"
)

var (
	skeleton = base.NewSkeleton()
	ChanRPC  = skeleton.ChanRPCServer

	c4c = &Conn4Center{}
)

type Module struct {
	*module.Skeleton
}

func (m *Module) OnInit() {
	m.Skeleton = skeleton

	var totalBetWin float64
	var totalBetLose float64
	var totalWinNum int
	var totalLoseNum int

	var winw int
	var winl int
	var losew int
	var losel int
	var lose int

	var sur1 int
	var sur2 int
	var sur3 int
	var randw int
	var randl int

	var surplusPool float64 = 1000

	for i := 0; i < 10000; i++ {
		fmt.Println(i)
		r := &Room{}
		p := &Player{}
		p.Init()
		r.PlayerList = append(r.PlayerList, p)
		p.IsAction = true

		r.Cards = new(CardData)

		var LuckWin = 0
		//p.DownBetMoneys.RedDownBet = 0
		//p.DownBetMoneys.RedDownBet = 50
		p.DownBetMoneys.LuckDownBet = 0
		p.DownBetMoneys.LuckDownBet = 10

		loseRate := 70

		percentageWin := 0
		countWin := 0
		percentageLose := 100
		countLose := 2

		settle := r.GetCardSettle()
		if settle >= 0 { // 玩家赢钱
			for {
				loseRateNum := RandInRange(1, 101)
				percentageWinNum := RandInRange(1, 101)
				if countWin > 0 {
					if percentageWinNum > int(percentageWin) { // 盈余池判定
						randw++
						if surplusPool > settle { // 盈余池足够
							sur1++
							break
						} else {                             // 盈余池不足
							if loseRateNum > int(loseRate) { // 30%玩家赢钱
								break
							} else { // 70%玩家输钱
								for {
									settle := r.GetCardSettle()
									if settle <= 0 {
										break
									}
								}
								break
							}
						}
					} else { // 又随机生成牌型
						settle := r.GetCardSettle()
						if settle > 0 { // 玩家赢
							countWin--
						} else {
							break
						}
					}
				} else {
					// 盈余池判定
					if surplusPool > settle { // 盈余池足够
						sur2++
						break
					} else {                             // 盈余池不足
						if loseRateNum > int(loseRate) { // 30%玩家赢钱
							for {
								settle := r.GetCardSettle()
								if settle >= 0 {
									winw++
									break
								}
							}
							break
						} else { // 70%玩家输钱
							for {
								log.Debug("进来了3")
								settle := r.GetCardSettle()
								if settle <= 0 {
									winl++
									break
								}
							}
							break
						}
					}
				}
			}
		} else { // 玩家输钱
			for {
				loseRateNum := RandInRange(1, 101)
				percentageLoseNum := RandInRange(1, 101)
				if countLose > 0 {
					if percentageLoseNum > int(percentageLose) {
						randl++
						break
					} else { // 又随机生成牌型
						settle := r.GetCardSettle()
						if settle >= 0 { // 玩家赢
							// 盈余池判定
							if surplusPool > settle { // 盈余池足够
								sur3++
								break
							} else {
								// 盈余池不足
								if loseRateNum > int(loseRate) { // 30%玩家赢钱
									for {
										settle := r.GetCardSettle()
										if settle >= 0 {
											losew++
											break
										}
									}
									break
								} else { // 70%玩家输钱
									for {
										settle := r.GetCardSettle()
										if settle <= 0 {
											losel++
											break
										}
									}
									break
								}
							}
						} else {
							countLose--
						}
					}
				} else { // 玩家输钱
					for {
						settle := r.GetCardSettle()
						if settle <= 0 {
							lose++
							break
						}
					}
					break
				}
			}
		}
		ag := dealer.GetGroup(aCard)
		bg := dealer.GetGroup(bCard)
		if ag.Weight > bg.Weight {
			if p.DownBetMoneys.LuckDownBet > 0 {
				if ag.IsThreeKind() {
					r.Cards.LuckType = CardsType(Leopard)
				}
				if ag.IsStraightFlush() {
					r.Cards.LuckType = CardsType(Shunjin)
				}
				if ag.IsFlush() {
					r.Cards.LuckType = CardsType(Golden)
				}
				if ag.IsStraight() {
					r.Cards.LuckType = CardsType(Straight)
				}
				if r.Cards.LuckType != CardsType(Leopard) {
					if (ag.Key.Pair() >> 8) >= 9 {
						r.Cards.LuckType = CardsType(Pair)
					}
				}
				if r.Cards.LuckType == Leopard {
					totalBetWin += float64(p.DownBetMoneys.LuckDownBet * WinLeopard)
					surplusPool -= float64(p.DownBetMoneys.LuckDownBet * WinLeopard)
					LuckWin = 1
				}
				if r.Cards.LuckType == Shunjin {
					totalBetWin += float64(p.DownBetMoneys.LuckDownBet * WinShunjin)
					surplusPool -= float64(p.DownBetMoneys.LuckDownBet * WinShunjin)
					LuckWin = 1
				}
				if r.Cards.LuckType == Golden {
					totalBetWin += float64(p.DownBetMoneys.LuckDownBet * WinGolden)
					surplusPool -= float64(p.DownBetMoneys.LuckDownBet * WinGolden)
					LuckWin = 1
				}
				if r.Cards.LuckType == Straight {
					totalBetWin += float64(p.DownBetMoneys.LuckDownBet * WinStraight)
					surplusPool -= float64(p.DownBetMoneys.LuckDownBet * WinStraight)
					LuckWin = 1
				}
				if r.Cards.LuckType == Pair {
					totalBetWin += float64(p.DownBetMoneys.LuckDownBet * WinBigPair)
					surplusPool -= float64(p.DownBetMoneys.LuckDownBet * WinBigPair)
					LuckWin = 1
				}
				if LuckWin == 1 {
					totalWinNum++
				} else {
					totalLoseNum++
					surplusPool += float64(p.DownBetMoneys.LuckDownBet)
					totalBetLose += float64(p.DownBetMoneys.LuckDownBet)
				}
			} else {
				if p.DownBetMoneys.RedDownBet > 0 {
					totalWinNum++
					totalBetWin += float64(p.DownBetMoneys.RedDownBet)
					surplusPool -= float64(p.DownBetMoneys.RedDownBet)
				} else {
					totalLoseNum++
					totalBetLose += float64(p.DownBetMoneys.RedDownBet)
					surplusPool += float64(p.DownBetMoneys.RedDownBet)
				}
			}
		} else {
			if p.DownBetMoneys.LuckDownBet > 0 {
				if bg.IsThreeKind() {
					r.Cards.LuckType = CardsType(Leopard)
				}
				if bg.IsStraightFlush() {
					r.Cards.LuckType = CardsType(Shunjin)
				}
				if bg.IsFlush() {
					r.Cards.LuckType = CardsType(Golden)
				}
				if bg.IsStraight() {
					r.Cards.LuckType = CardsType(Straight)
				}
				if r.Cards.LuckType != CardsType(Leopard) {
					if (bg.Key.Pair() >> 8) >= 9 {
						r.Cards.LuckType = CardsType(Pair)
					}
				}
				if r.Cards.LuckType == Leopard {
					totalBetWin += float64(p.DownBetMoneys.LuckDownBet * WinLeopard)
					surplusPool -= float64(p.DownBetMoneys.LuckDownBet * WinLeopard)
					LuckWin = 1
				}
				if r.Cards.LuckType == Shunjin {
					totalBetWin += float64(p.DownBetMoneys.LuckDownBet * WinShunjin)
					surplusPool -= float64(p.DownBetMoneys.LuckDownBet * WinShunjin)
					LuckWin = 1
				}
				if r.Cards.LuckType == Golden {
					totalBetWin += float64(p.DownBetMoneys.LuckDownBet * WinGolden)
					surplusPool -= float64(p.DownBetMoneys.LuckDownBet * WinGolden)
					LuckWin = 1
				}
				if r.Cards.LuckType == Straight {
					totalBetWin += float64(p.DownBetMoneys.LuckDownBet * WinStraight)
					surplusPool -= float64(p.DownBetMoneys.LuckDownBet * WinStraight)
					LuckWin = 1
				}
				if r.Cards.LuckType == Pair {
					totalBetWin += float64(p.DownBetMoneys.LuckDownBet * WinBigPair)
					surplusPool -= float64(p.DownBetMoneys.LuckDownBet * WinBigPair)
					LuckWin = 1
				}
				if LuckWin == 1 {
					totalWinNum++
				} else {
					totalLoseNum++
					totalBetLose += float64(p.DownBetMoneys.LuckDownBet)
					surplusPool += float64(p.DownBetMoneys.LuckDownBet)
				}
			} else {
				if p.DownBetMoneys.BlackDownBet > 0 {
					totalWinNum++
					totalBetWin += float64(p.DownBetMoneys.RedDownBet)
					surplusPool -= float64(p.DownBetMoneys.RedDownBet)
				} else {
					totalLoseNum++
					totalBetLose += float64(p.DownBetMoneys.RedDownBet)
					surplusPool += float64(p.DownBetMoneys.RedDownBet)
				}
			}
		}
	}
	taxWinMoney := totalBetWin - (totalBetWin * 0.05)
	fmt.Println("玩家总赢局:", totalWinNum)
	fmt.Println("玩家总输局:", totalLoseNum)
	fmt.Println("玩家总输:", int64(totalBetLose))
	fmt.Println("玩家总赢:", int64(totalBetWin))
	fmt.Println("玩家总赢(税后):", int64(taxWinMoney))
	fmt.Println("总流水:", int64(totalBetWin+totalBetLose))
	fmt.Println(winw, winl, losew, losel, lose)
	fmt.Println(sur1, sur2, sur3)
	fmt.Println(randw,randl)

	initMongoDB()

	gameHall.Init()
	InitMapPlayer()

	//机器人初始化并开始
	gRobotCenter.Init()
	gRobotCenter.Start()

	//中心服初始化,主动请求Token
	c4c.Init()
	c4c.CreatConnect()
	//c4c.ReqCenterToken()

	//中心服日志初始化
	cc.Init()

	go StartHttpServer()

}

func (m *Module) OnDestroy() {
	c4c.onDestroy()
}
