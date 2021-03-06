package internal

import (
	"RedBlack-War/conf"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

func SendTgMessage(data string) {
	timeStr := time.Now().Format("2006-01-02 15:04:05")
	centerPort := conf.Server.CenterServerPort
	centreUrl := fmt.Sprintf("ws://" + conf.Server.CenterServer + ":" + centerPort)

	log.Println("连接服务地址:", centreUrl)

	var tgMessage string
	switch centreUrl {
	case "ws://172.16.100.2:9502":
		tgMessage = fmt.Sprintf("红黑大战游戏服务器" + "\n事件:" + data +
			"\n启动时间:" + timeStr + "\n环境：DEV")
		//SendToTelegram(tgMessage)
	case "ws://172.16.1.41:9502":
		tgMessage = fmt.Sprintf("红黑大战游戏服务器" + "\n事件:" + data +
			"\n启动时间:" + timeStr + "\n环境：PRE")
		SendToTelegram(tgMessage)
	default:
		tgMessage = fmt.Sprintf("红黑大战游戏服务器" + "\n事件:" + data +
			"\n启动时间:" + timeStr + "\n环境：OL")
		SendToTelegram(tgMessage)
	}
}

func SendToTelegram(data string) string {

	TelegramId := "-521977907"
	TelegramToken := "1726462670:AAEmwMgpIpxk0akDE3k-MuQCQ3rZm3NWGFU"
	var telegramApi = "https://api.telegram.org/bot" + TelegramToken + "/sendMessage"

	resp, err := http.PostForm(
		telegramApi,
		url.Values{
			"chat_id": {TelegramId},
			"text":    {data},
		})

	if err != nil {
		log.Println("postFrom err:", err.Error())
		return ""
	}
	defer resp.Body.Close()

	var body, errR = ioutil.ReadAll(resp.Body)
	if errR != nil {
		log.Println("readAll err:", errR.Error())
		return ""
	}

	return string(body)
}
