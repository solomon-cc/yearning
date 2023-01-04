package lib

import (
	"C"
	"log"

	"Yearning-go/src/model"
	"github.com/xen0n/go-workwx"
)

var userIds = map[string]string{
	"jun.liu@aylaasia.com":        "LiuJun",
	"tao.chen@aylaasia.com":       "0b7468c8e0f045a490038a2b3ad356fb",
	"wei.fan@aylaasia.com":        "bd5ca88ea2384681af0e00dc6b157496",
	"xudong.fu@aylaasia.com":      "9301a5a2726846f58f8315dfa568bea4",
	"jing.lu@aylaasia.com":        "943e292081f842b8a3780ec5eca945ed",
	"jacky.song@aylaasia.com":     "9a2bf7dc19aa4ed78f0d11ca71432421",
	"tyler.tang@aylaasia.com":     "TangLei",
	"xiaodong.wang@aylaasia.com":  "88c04a58e20a434894154fdb0c23a3a7",
	"dinglin.wu@aylaasia.com":     "WuDingLin",
	"lin.yang@aylaasia.com":       "1cb9d8d2173944a49a00c9df100a2dae",
	"guoqiang.ye@aylaasia.com":    "YeGuoQiang",
	"jinghe.zhang@aylaasia.com":   "ZhangJingHe",
	"yinglong.zhang@aylaasia.com": "9436641919f243fa831fe9a5ca16b925",
	"jim.zhao@aylaasia.com":       "cff688373b3244039abaf127decaf0e9",
	"mingyang.zou@aylaasia.com":   "ed9a6ce5e0fa4fbea0c97ecb8cf0fcde",
	"cheng.chen@aylaasia.com":     "13027da7ccde410d95140ac7525c4e6e",
	"taidong.li@aylaasia.com":     "LiTaiDong",
	"feng.pi@aylaasia.com":        "5522380e65cf47dbad2db78ad3e7e252",
}

func SendWxMsg(msg model.Message, sv string) {
	client := workwx.New(model.C.General.CorpID)

	app := client.WithApp(model.C.General.CorpSecret, model.C.General.AgentID)

	uIDs := whoNeedToSend(msg)
	err := app.SendMarkdownMessage(
		&workwx.Recipient{
			UserIDs: uIDs,
		}, sv, false)

	if err != nil {
		log.Println(err.Error())
		return
	}

}

func whoNeedToSend(msg model.Message) []string {
	var uId []string

	u := userIds[msg.User]
	toU := userIds[msg.ToUser]

	uId = append(uId, u, toU)

	return uId
}
