package global

import (
	"custom-go/generated"
	"custom-go/pkg/base"
	"custom-go/pkg/plugins"
	"custom-go/pkg/utils"
	"net/http"
	"strings"
	"sync"
	"time"
)

type (
	createLogI generated.System__Log__CreateLogInput
	createLogO generated.System__Log__CreateLogResponseData

	getIsOpenI generated.System__Log__GetIsOpenInput
	getIsOpenO generated.System__Log__GetIsOpenResponseData
)

var Mq = make(chan string)
var Wg = new(sync.WaitGroup)

func BeforeOriginRequest(hook *base.HttpTransportHookRequest, body *plugins.HttpTransportBody) (*base.ClientRequest, error) {
	hook.Logger().Info()
	result, _ := plugins.ExecuteInternalRequestQueries[getIsOpenI, getIsOpenO](hook.InternalClient, generated.System__Log__GetIsOpen, getIsOpenI{})

	if result.Main_findFirstdic.IsOpen {
		go func() {
			Wg.Add(1)

			//上一个客户端连接的地址
			defer Wg.Done()
			code := <-Mq

			//获取请求的是哪个api
			api := body.Request.RequestURI
			path := strings.Split(api, "?")[0]
			bearerToken := body.Request.Headers["Authorization"]
			token := strings.Split(bearerToken, " ")[1]
			//解析token
			parseToken, err := utils.ParseToken(token)
			var userId string
			if parseToken.User.UserId == "" {
				userId = " "
			}
			_, err = plugins.ExecuteInternalRequestMutations[createLogI, createLogO](hook.InternalClient, generated.System__Log__CreateLog, createLogI{
				Code:      code,
				UpdatedAt: GetCurrentTime(),
				Ip:        body.Request.Headers["X-Real-Ip"],
				Method:    body.Request.Method,
				Path:      path,
				Ua:        body.Request.Headers["User-Agent"],
				UserName:  parseToken.User.Name,
				UserId:    userId,
			})

			if err != nil {
				hook.Logger().Errorf(err.Error())
			}
		}()

	}

	statusCode := http.StatusOK

	body.Response = &base.ClientResponse{}
	body.Response.StatusCode = statusCode

	return body.Request, nil
}

func GetCurrentTime() string {
	timestamp := time.Now().Unix()
	tm := time.Unix(timestamp, 0)
	return tm.Format(time.RFC3339)
}
