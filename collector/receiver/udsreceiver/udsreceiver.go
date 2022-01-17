package udsreceiver

import (
	analyzerpackage "github.com/dxsup/kindling-collector/analyzer"
	"github.com/dxsup/kindling-collector/analyzer/network"
	"github.com/dxsup/kindling-collector/analyzer/tcpmetricanalyzer"
	"github.com/dxsup/kindling-collector/analyzer/uprobeanalyzer"
	"github.com/dxsup/kindling-collector/model"
	"github.com/dxsup/kindling-collector/model/constnames"
	"github.com/dxsup/kindling-collector/receiver"
	"github.com/golang/protobuf/proto"
	zmq "github.com/pebbe/zmq4"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"sync"
	"time"
)

const (
	Uds = "udsreceiver"
)

type Socket struct {
	*zmq.Socket
}

type UdsReceiver struct {
	cfg             *Config
	analyzerManager analyzerpackage.Manager
	zmqPullSocket   Socket
	zmqReqSocket    Socket
	shutdownWG      sync.WaitGroup
	shutdwonState   bool
	logger          *zap.Logger
}

type Config struct {
	ZEROMQPULL *ZeroMqPullSettings `mapstructure:"zeromqpull"`
	ZEROMQREQ  *ZeroMqReqSettings  `mapstructure:"zeromqreq"`
}

type ZeroMqPullSettings struct {
	// Endpoint configures the listening address for the server.
	Endpoint string `mapstructure:"endpoint"`
	hwm      int    `mapstructure:"hwm"`
}

type SubEvent struct {
	Category string `mapstructure:"category"`
	Name     string `mapstructure:"name"`
}

type ZeroMqReqSettings struct {
	// Endpoint configures the listening address for the server.
	Endpoint     string     `mapstructure:"endpoint"`
	hwm          int        `mapstructure:"hwm"`
	SubcribeInfo []SubEvent `mapstructure:"subscribe"`
}

func (r *UdsReceiver) newPullSocket(zss *ZeroMqPullSettings) Socket {

	zmqContextServer, _ := zmq.NewContext()
	ServerClient, _ := zmqContextServer.NewSocket(zmq.PULL)
	if zss.hwm != 0 {
		ServerClient.SetSndhwm(zss.hwm)
	}
	return Socket{ServerClient}
}

func (r *UdsReceiver) newReqSocket(zss *ZeroMqReqSettings) Socket {
	zmqContextServer, _ := zmq.NewContext()
	ServerClient, _ := zmqContextServer.NewSocket(zmq.REQ)
	if zss.hwm != 0 {
		ServerClient.SetSndhwm(zss.hwm)
	}
	return Socket{ServerClient}
}

func (soc Socket) connect(endpoint string) error {
	var failedCount = 0
	var err error
	for {
		err = soc.Connect(endpoint)
		if err != nil {
			failedCount++
			if failedCount > 3 {
				break
			}
			time.Sleep(time.Second)
		} else {
			err = nil
			break
		}
	}
	return err
}

func NewUdsReceiver(config interface{}, logger *zap.Logger, analyzerManager analyzerpackage.Manager) receiver.Receiver {
	cfg, ok := config.(*Config)
	if !ok {
		logger.Sugar().Panicf("Cannot convert [%s] config", Uds)
	}
	return &UdsReceiver{cfg: cfg, analyzerManager: analyzerManager, logger: logger}
}

func (r *UdsReceiver) startZeroMqPull() error {
	cfg := r.cfg.ZEROMQPULL
	r.logger.Info("Starting ZeroMq Pull connect on endpoint", zap.String("endpoint", cfg.Endpoint))
	pullSocket := r.zmqPullSocket
	err := pullSocket.connect(cfg.Endpoint)
	if err != nil {
		r.logger.Panic("Connecting ZeroMq Pull failed on endpoint", zap.String("endpoint", cfg.Endpoint))
	}
	r.shutdownWG.Add(1)
	go func() {
		defer r.shutdownWG.Done()
		for {
			if r.shutdwonState == true {
				err := pullSocket.Close()
				if err != nil {
					return
				}
				r.shutdwonState = false
				break
			}
			req, _ := pullSocket.RecvMessage(0)
			if len(req) >= 1 {
				events := &model.KindlingEventList{}
				err = proto.Unmarshal([]byte(req[0]), events)
				if err != nil {
					r.logger.Error("Error unmarshalling event: %v", zap.Error(err))
					continue
				}
				err = r.SendToNextConsumer(events)
				if err != nil {
					r.logger.Error("Error sending event to next consumer: %v", zap.Error(err))
					continue
				}
				//r.logger.Info("name"+data.HcmineEvent[0].GetName())
			}

		}
	}()
	r.shutdownWG.Wait()
	return nil
}

func (r *UdsReceiver) startZeroMqReq() error {
	cfg := r.cfg.ZEROMQREQ
	cfgPull := r.cfg.ZEROMQPULL
	reqSocket := r.zmqReqSocket
	err := reqSocket.connect(cfg.Endpoint)
	if err != nil {
		r.logger.Error("Connecting ZeroMq Req failed on endpoint " + cfg.Endpoint)
		return err
	}
	labels := make([]*model.Label, len(cfg.SubcribeInfo))
	for i, e := range cfg.SubcribeInfo {
		labels[i] = &model.Label{Name: e.Name, Category: e.Category}
	}
	pid := os.Getpid()
	subEvent := &model.SubEvent{
		Address: []byte(cfgPull.Endpoint),
		Pid:     uint32(pid),
		Labels:  labels,
	}
	p, err := proto.Marshal(subEvent)
	if err != nil {
		r.logger.Error("subscribe events marshal failed.")
	}
	_, err = reqSocket.SendMessage(p)
	if err != nil {
		r.logger.Error("subscribe request failed")
	}
	rep, err := reqSocket.RecvMessage(8)
	if ce := r.logger.Check(zapcore.DebugLevel, "Receiver Message"); ce != nil {
		ce.Write(
			zap.String("rep", rep[0]),
		)
	}
	return err
}

func (r *UdsReceiver) Start() error {
	var err error
	if r.cfg.ZEROMQREQ != nil {
		r.zmqReqSocket = r.newReqSocket(r.cfg.ZEROMQREQ)
	}
	if r.cfg.ZEROMQPULL != nil {
		r.zmqPullSocket = r.newPullSocket(r.cfg.ZEROMQPULL)
	}
	r.logger.Info("startZeroMqReq")
	err = r.startZeroMqReq()
	if err != nil {
		return err
	}
	r.logger.Info("startZeroMqPull")
	err = r.startZeroMqPull()
	if err != nil {
		return err
	}
	return err
}

func (r *UdsReceiver) Shutdown() error {
	var err error
	if r.zmqPullSocket.Socket != nil {
		r.shutdwonState = true
		time.Sleep(1 * time.Second)
		//err = r.zmqPullSocket.Close()
		//r.shutdownWG.Done()
	}

	r.shutdownWG.Wait()
	return err
}

func (r *UdsReceiver) SendToNextConsumer(events *model.KindlingEventList) error {
	// TODO: Decouple dispatching logic from receiver and conduct it at analyzerManager via configuration
	for _, evt := range events.KindlingEventList {
		var analyzer analyzerpackage.Analyzer
		var isFound bool
		switch evt.Name {
		case constnames.GrpcUprobeEvent:
			analyzer, isFound = r.analyzerManager.GetAnalyzer(uprobeanalyzer.UprobeType)
		case constnames.TcpCloseEvent:
			fallthrough
		case constnames.TcpRcvEstablishedEvent:
			fallthrough
		case constnames.TcpDropEvent:
			fallthrough
		case constnames.TcpRetransmitSkbEvent:
			analyzer, isFound = r.analyzerManager.GetAnalyzer(tcpmetricanalyzer.TcpMetric)
		default:
			analyzer, isFound = r.analyzerManager.GetAnalyzer(network.Network)
		}
		if !isFound {
			r.logger.Info("analyzer not found for event %s", zap.String("eventName", evt.Name))
			continue
		}
		if ce := r.logger.Check(zapcore.DebugLevel, "Receive Event"); ce != nil {
			ce.Write(
				zap.String("event", evt.String()),
			)
		}
		err := analyzer.ConsumeEvent(evt)
		if err != nil {
			return err
		}
	}
	return nil
}
