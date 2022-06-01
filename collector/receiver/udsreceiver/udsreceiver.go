package udsreceiver

import (
	"os"
	"sync"
	"time"

	analyzerpackage "github.com/Kindling-project/kindling/collector/analyzer"
	"github.com/Kindling-project/kindling/collector/component"
	"github.com/Kindling-project/kindling/collector/model"
	"github.com/Kindling-project/kindling/collector/receiver"
	"github.com/gogo/protobuf/proto"
	zmq "github.com/pebbe/zmq4"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	Uds = "udsreceiver"
)

type Socket struct {
	*zmq.Socket
}

type UdsReceiver struct {
	cfg             *Config
	analyzerManager *analyzerpackage.Manager
	zmqPullSocket   Socket
	zmqReqSocket    Socket
	shutdownWG      sync.WaitGroup
	shutdwonState   bool
	telemetry       *component.TelemetryTools
	stats           eventCounter
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

func NewUdsReceiver(config interface{}, telemetry *component.TelemetryTools, analyzerManager *analyzerpackage.Manager) receiver.Receiver {
	cfg, ok := config.(*Config)
	if !ok {
		telemetry.Logger.Sugar().Panicf("Cannot convert [%s] config", Uds)
	}
	udsReceiver := &UdsReceiver{
		cfg:             cfg,
		analyzerManager: analyzerManager,
		telemetry:       telemetry,
		stats:           &stats{},
	}
	newSelfMetrics(telemetry.MeterProvider, udsReceiver.stats)
	return udsReceiver
}

func (r *UdsReceiver) startZeroMqPull() error {
	cfg := r.cfg.ZEROMQPULL
	r.telemetry.Logger.Info("Starting ZeroMq Pull connect on endpoint", zap.String("endpoint", cfg.Endpoint))
	pullSocket := r.zmqPullSocket
	err := pullSocket.connect(cfg.Endpoint)
	if err != nil {
		r.telemetry.Logger.Panic("Connecting ZeroMq Pull failed on endpoint", zap.String("endpoint", cfg.Endpoint))
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
					r.telemetry.Logger.Error("Error unmarshalling event: %v", zap.Error(err))
					continue
				}
				err = r.SendToNextConsumer(events)
				if err != nil {
					r.telemetry.Logger.Error("Error sending event to next consumer: %v", zap.Error(err))
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
		r.telemetry.Logger.Error("Connecting ZeroMq Req failed on endpoint " + cfg.Endpoint)
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
		r.telemetry.Logger.Error("subscribe events marshal failed.")
	}
	_, err = reqSocket.SendMessage(p)
	if err != nil {
		r.telemetry.Logger.Error("subscribe request failed")
	}
	rep, err := reqSocket.RecvMessage(8)
	if ce := r.telemetry.Logger.Check(zapcore.DebugLevel, "Receiver Message"); ce != nil {
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
	r.telemetry.Logger.Info("startZeroMqReq")
	err = r.startZeroMqReq()
	if err != nil {
		return err
	}
	r.telemetry.Logger.Info("startZeroMqPull")
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
	for _, evt := range events.KindlingEventList {
		r.stats.add(evt.Name, 1)
		if ce := r.telemetry.Logger.Check(zapcore.DebugLevel, "Receive Event"); ce != nil {
			ce.Write(
				zap.String("event", evt.String()),
			)
		}
		analyzers := r.analyzerManager.GetConsumableAnalyzers(evt.Name)
		if analyzers == nil || len(analyzers) == 0 {
			r.telemetry.Logger.Info("analyzer not found for event ", zap.String("eventName", evt.Name))
			continue
		}
		for _, analyzer := range analyzers {
			err := analyzer.ConsumeEvent(evt)
			if err != nil {
				return err
			}
			if err != nil {
				r.telemetry.Logger.Warn("Error sending event to next consumer: ", zap.Error(err))
			}
		}
	}
	return nil
}
