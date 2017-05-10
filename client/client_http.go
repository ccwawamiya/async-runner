package client

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"strings"
)
type Listener func(body []byte) error

type HttpReceiver struct {
	listeners  map[string]Listener
	listenAddr string
}

type HttpPusher struct {
	callbackAddr string
	serverAddr   string
}

// 在startServer后会监听listenAddr地址
func NewHttpReceiver(listenAddr string) *HttpReceiver {
	return &HttpReceiver{
		listeners:  map[string]Listener{},
		listenAddr: listenAddr,
	}
}

// callbackAddr 服务器地址，callbackAddr 回调地址 同时会监听回调地址以实现Listener功能
func NewHttpPusher(serverAddr, callbackAddr string) *HttpPusher {
	return &HttpPusher{
		callbackAddr: callbackAddr,
		serverAddr:   serverAddr,
	}
}

// 开启阻塞服务
func (p *HttpReceiver) StartServer() (err error) {
	err = fasthttp.ListenAndServe(p.listenAddr, func(ctx *fasthttp.RequestCtx) {
		uri := ctx.Request.URI()
		path := string(uri.Path())
		pathS := strings.Split(path, "/")
		args := uri.QueryArgs()
		topic := string(args.Peek("topic"))

		data := ctx.Request.PostArgs().Peek("data")

		if pathS[1] == "do_job" {
			fun, ok := p.listeners[topic]
			if !ok {
				ctx.WriteString("no topic named :" + topic)
				return
			}
			err := fun(data)
			if err != nil {
				ctx.WriteString(err.Error())
				return
			}
		}

		ctx.WriteString("success")
		return
	})

	return
}

func (p *HttpReceiver) AddListener(topic string, listener Listener) {
	p.listeners[topic] = listener
}

func (p *HttpPusher) Push(topic string, timeout int64, data []byte) (err error) {
	arg := fasthttp.Args{}
	arg.SetBytesV("data", data)
	_, _, err = fasthttp.Post(nil, fmt.Sprintf("%s/push?topic=%s&timeout=%d&callback=%s", p.serverAddr, topic, timeout, p.callbackAddr), &arg)
	if err != nil {
		return
	}
	return
}

func init() {

}
