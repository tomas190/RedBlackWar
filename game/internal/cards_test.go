package internal

import (
	"fmt"
	"testing"
)

func Test_Dealer(t *testing.T) {
	//d := &Cards{}
	//
	//d.Shuffle()
	//
	//var ca Cards
	//ca = Cards{d.Take(), d.Take(), d.Take()}
	//fmt.Println("牌型类型1 :", ca)
	//fmt.Println("牌型类型2 :", ca.Hex())
	//fmt.Println("牌型类型3 :", ca.HexInt())
	//
	//fmt.Println("牌型剩余Take数量 ~ :", len(*d))

	this := &RBdzDealer{}

	// 检查剩余牌数量
	offset := this.Offset
	if offset >= len(this.Poker)/2 {
		this.Poker = NewPoker(1, false, true)
		offset = 0
	}
	aaa := Hex(this.Poker)
	fmt.Println("12:", aaa)
	// 红黑各取3张牌
	a := this.Poker[offset : offset+3]
	b := this.Poker[offset+3 : offset+6]

	note := PokerArrayString(a) + "|" + PokerArrayString(b)
	fmt.Println("note:::", note)

	hexa := HexInt(a)
	str1 := fmt.Sprintf("%#v", a)
	fmt.Println("offfff1 ::", str1)
	fmt.Println("111:", hexa)
	hexb := HexInt(b)
	str2 := fmt.Sprintf("%#v", b)
	fmt.Println("offfff2 ::", str2)
	fmt.Println("222:", hexb)

	//RBdzPk()
}

var betItemCount int
var taxRate = []int64{50, 50, 50}

// 预算输赢(prize:扣税前总返奖，tax:总税收，bet:总下注)
//func Balance(group []int64, odds []int32) (prize, tax, bet int64) {
//	for i := 0; i < betItemCount; i++ {
//		// 下注金币大于0
//		if b := group[i]; b > 0 {
//			bet += b
//			if odd := int64(odds[i]); odd != 0 {
//				w := b * odd / internal.Radix
//				//有钱回收,包含输1半
//				if w > b {
//					// 赢钱了收税，税率按千分比配置，需除以1000
//					tax += (w - b) * taxRate[i] / 1000
//				}
//				prize += w
//			}
//		}
//	}
//	return
//}
