package server

import (
	"errors"
	"fmt"
	"github.com/bysir-zl/async-runner/core"
	"github.com/bysir-zl/bygo/log"
	"github.com/valyala/fasthttp"
	"strconv"
	"strings"
	"bytes"
	"encoding/binary"
)

type HttpServer struct {
	s *core.Scheduler
}

func NewServer() *HttpServer {
	return &HttpServer{
		s: core.NewScheduler(),
	}
}

func (p *HttpServer) Start() (err error) {
	go p.s.Work()
	log.Info("server_http", "start server success")
	err = fasthttp.ListenAndServe(":9989", func(ctx *fasthttp.RequestCtx) {
		p.handlerQuery(ctx)
		return
	})
	return
}

func (p *HttpServer) handlerQuery(ctx *fasthttp.RequestCtx) {
	req := ctx.Request
	uri := req.URI()
	path := string(uri.Path())
	pathS := strings.Split(path, "/")
	action := pathS[1]

	args := uri.QueryArgs()
	topic := string(args.Peek("topic"))
	callback := string(args.Peek("callback"))
	timeout := string(args.Peek("timeout"))

	form := req.PostArgs()
	data := form.Peek("data")

	timeoutInt, _ := strconv.ParseInt(timeout, 10, 64)

	if action == "push" {
		// 添加一个工作
		p.s.AddJob(timeoutInt, NewJobHttpClient(callback, topic, data))
	}
}

// job

type JobHttp struct {
	callback, topic string
	data            []byte
}

var defaultClient fasthttp.Client

func (p *JobHttp) String() string {
	return fmt.Sprintf("topic:%s", p.topic)
}

func (p *JobHttp) Unmarshal(bs []byte) error {
	err := binary.Read(bs, binary.LittleEndian, p)
	return err
}

func (p *JobHttp) Marshal() ([]byte, error) {
	bf := bytes.Buffer{}
	err := binary.Write(bf, binary.LittleEndian, p)
	if err != nil {
		return nil, err
	}
	return bf.Bytes(), nil
}
func (p *JobHttp) Unique() []byte {
	bf := bytes.Buffer{}
	bf.WriteString(p.callback)
	bf.WriteString(p.topic)
	bf.WriteString(p.data)
	return bf.Bytes()
}

func (p *JobHttp) Run() (err error) {
	args := fasthttp.Args{}
	// fasthttp有点奇葩, post只能是键值对
	args.SetBytesV("data", p.data)

	_, body, err := defaultClient.Post(nil, p.callback + "/do_job?topic=" + p.topic, &args)
	if err != nil {
		return
	}

	bodyString := string(body)
	if bodyString != "success" {
		err = errors.New("client response is not 'success', is " + bodyString)
		return
	}
	return nil
}

func NewJobHttpClient(callback, topic string, data []byte) core.Job {
	job := &JobHttp{
		callback: callback,
		topic:    topic,
		data:     data,
	}
	return job
}

func init() {
	defaultClient.MaxConnsPerHost = 60000
}