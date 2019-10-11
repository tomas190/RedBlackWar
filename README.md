
### **一、部署文件**
##### 1、克隆代码到本地
git clone http://git.0717996.com/Tomas/RedBlack-War.git

##### 2、进入RedBlack-War文件夹
cd RedBlack-War

##### 3、编译
go build -o server ./main.go  `要权限`

##### 4、后台运行
nohup ./server >load.log 2>&1 &  `要权限`

##### 5、查看是否运行成功
cat load.log

###### 如果看到日志文件输出以下数据代表成功启动~
1. 2019/10/08 14:37:25 [debug  ] msg init~
1. 2019/10/08 14:37:25 [release] Leaf 1.1.3 starting up
1. 2019/10/08 14:37:25 [debug  ] Connect DataBase 数据库连接成功~
1. 2019/10/08 14:37:25 [debug  ] GameHall Init~!!! This gameHall can hold 5000 player running ~


### **二、項目所需套件支持**
###### 1、Go语言配置环境   `go version go1.13 linux/amd64`
###### 2、Mongo数据库     `MongoDB server version: 4.0.12`


### **三、配置文件位置及文件名稱**
##### 1、文件名称: `server/conf/server.json`
##### 2、日志文件：`load.log`  `路径为：编译好可执行文件同级`
##### 3、服务配置信息：
```
{
  "LogLevel": "debug",
  "LogPath": "",
  "WSAddr": "0.0.0.0:1214",     
  "MaxConnNum": 20000,

  "MongoDBAddr": "172.16.100.5:27017",   Mongo数据库连接地址
  "MongoDBAuth": "",                     Mongo认证(可不填默认admin)
  "MongoDBUser": "rbdz",                 Mongo连接用户名
  "MongoDBPwd": "123456",                Mongo连接密码

  "TokenServer": "http://172.16.100.2:9502/Token/getToken", 中心服Token
  "CenterServer": "172.16.100.2",                           中心服地址 (易动)           
  "CenterServerPort": "9502",                               中心服端口 (易动)
  "DevKey": "new_game_20",                                  devKey
  "DevName": "新游戏开发"                                    devName
  "GameID": "5b1f3a3cb76a591e7f251719"                      gameID
}
```

