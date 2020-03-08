package modbusserver

import (
	"context"
	"io"
	"net"
	"strings"
	"sync"
	"time"
	"xlib/log"
	. "xmodbustcp/framer"
)

var (
	cancel context.CancelFunc
)

type Server struct {
	listeners   []net.Listener                                    //Tcp
	requestChan chan *Request                                     //请求报文
	function    [256](func(*Server, Framer) ([]byte, *Exception)) //对应功能码
	wg          *sync.WaitGroup
}

type Request struct {
	conn  io.ReadWriteCloser
	frame Framer
}

func NewServer() *Server {
	s := &Server{
		wg:          new(sync.WaitGroup),
		requestChan: make(chan *Request),
	}

	svr_status_mutex.Lock()
	svr_status = true
	svr_status_mutex.Unlock()

	go s.handler()
	return s
}



func (s *Server) RegisterFunctionHandler(funcCode uint8, function func(*Server, Framer) ([]byte, *Exception)) {
	s.function[funcCode] = function
}

func (s *Server) handle(request *Request) Framer {
	//获取数据
	var exception *Exception
	var bytes []byte

	response := request.frame.Copy()

	//获取功能码
	dealdata := request.frame.GetFunction()

	//执行处理操作函数
	if s.function[dealdata] != nil {
		bytes, exception = s.function[dealdata](s, request.frame)
		log.Info("byte", bytes)
		response.SetData(bytes)
	} else {
		exception = &IllegalFunction
	}

	if exception != &Success {
		response.SetException(exception)
	}

	return response
}

func (s *Server) Close(){
	for _,listener := range s.listeners{
		listener.Close()
	}
}


func (s *Server) handler() {
	s.wg.Add(1)
	defer s.wg.Done()
	for {
		request, ok := <-s.requestChan
		if !ok {
			log.Debug("管道已经关闭")
			break
		}
		response := s.handle(request)
		request.conn.Write(response.Bytes())
	}
}

func (s *Server) accept(listen net.Listener) error {
	s.wg.Add(1)
	defer s.wg.Done()
	defer close(s.requestChan)
	for {
			conn, err := listen.Accept()
			if err != nil {
				if strings.Contains(err.Error(), "use of closed network connection") {
					return nil
				}

				log.Errorf("不能accept: %#v\n", err)
				return err
			}
			log.Info("客户端 %v", conn.RemoteAddr(), " 连接了")

			go s.readandresponse(conn)
	}
	return nil
}

func (s *Server) readandresponse(conn net.Conn) {
	s.wg.Add(1)
	defer s.wg.Done()
	defer conn.Close()
	packet := make([]byte, 512,1024)

	for {

		if deadline := conn.SetDeadline(time.Now().Add(2 * time.Minute)); deadline != nil {
			log.Error("连接超时设置错误", deadline)
			return
		}

		bytesRead, err := conn.Read(packet)
		if err != nil {
			if err == io.EOF {
				log.Error("客户端与服务器断开连接 \n", err, "Addr", conn.RemoteAddr())
			} else if e, _ := err.(*net.OpError); e.Timeout() {
				log.Error("Read超时", err)
			} else {
				log.Error("TCPError", err)
			}
			return
		}

		if !checkstatus(){
			log.Info("状态变更，服务退出")
			s.Close()
			return
		}

		packet = packet[:bytesRead]
		frame, err := NewTCPFrame(packet)
		if err != nil {
			log.Error("Frame错误 %v\n", err)
			return
		}

		request := &Request{conn, frame}
		s.requestChan <- request
	}
}

// 监听ModbusTcp连接
func (s *Server) ListenTCP(addressPort string) (err error) {
	log.Info("开始监听：" + addressPort)
	listen, err := net.Listen("tcp", addressPort)
	if err != nil {
		log.Error("监听失败: %v\n", err)
		return err
	}

	s.listeners = append(s.listeners, listen)
	go s.accept(listen)
	s.wg.Wait()
	return err
}

//停止服务
func StopSvr() {
	svr_status_mutex.Lock()
	svr_status = false
	svr_status_mutex.Unlock()
	log.Info("收到信号，停止服务")
}


var(
	svr_status = false
	svr_status_mutex = new(sync.RWMutex)
)

//检查服务状态
func checkstatus()(status bool){
	svr_status_mutex.RLock()
	status = svr_status
	svr_status_mutex.RUnlock()
	return
}
