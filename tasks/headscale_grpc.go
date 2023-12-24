/*
	The code first initializes the gRPC client using the InitGRPC() function.
	The  ReConnect() method is responsible for reconnecting the gRPC client if needed,
and the connect() method establishes a connection to the gRPC server.
	The  GetRequestMetadata()  and  RequireTransportSecurity() methods are used
to set the request metadata and determine whether transport security is required, respectively.
	The  newApiKey()  method creates a new API key and updates the configuration,
while the  unaryClientInterceptor()  method acts as an interceptor for unary gRPC calls,
logging information about the call and handling errors.
*/

package task

import (
	"context"
	"crypto/tls"
	"fmt"
	pb "github.com/juanfont/headscale/gen/go/headscale/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"headscale-panel/common"
	"headscale-panel/config"
	"headscale-panel/log"
	"headscale-panel/model"
	"headscale-panel/util"
	"time"
)

// Times is a constant that represents the maximum number of times the gRPC client will try to reconnect.
const Times = 3

var retryTimes = 0
var connecting = false

// headscaleRPC is a struct that represents the gRPC client for the Headscale service.
type headscaleRPC struct {
	opts []grpc.DialOption
	conn *grpc.ClientConn
	pb.HeadscaleServiceClient
	scale  HeadscaleService
	status int
}

// RPCHeader and RPCSetting are structs that define the gRPC request header and settings, respectively.
type RPCHeader struct {
	Authorization string `json:"authorization"`
}

type RPCSetting struct {
	ServerAddr string `json:"server_addr"`
}

// HeadscaleControl is a global variable that holds the gRPC client instance.
var HeadscaleControl *headscaleRPC

// InitGRPC is a function that initializes the gRPC client.
func InitGRPC() {
	h := &headscaleRPC{}
	defer func() {
		HeadscaleControl = h
	}()

	if err := h.Connect(); err != nil {
		log.Log.Error(err)
		return
	}
	log.Log.Info("init gRPC finished")
}

// ReConnect is a method of the headscaleRPC struct that tries to reconnect the gRPC client.
func (h *headscaleRPC) Connect() error {
	if connecting {
		return fmt.Errorf("gRPC is trying to connect, wait for again")
	}
	connecting = true
	defer func() {
		connecting = false
	}()
	// wait headscale grpc server ready
	time.Sleep(4 * time.Second)
	conf := common.GetHeadscaleConfig()

	h.opts = []grpc.DialOption{grpc.WithPerRPCCredentials(h), grpc.WithUnaryInterceptor(h.unaryClientInterceptor)}
	if len(conf.Cert) > 0 && len(conf.Key) > 0 {
		var cert tls.Certificate
		if len(conf.CA) > 0 {
			util.LoadCA(conf.CA)
		}

		// Load server cert and key
		cert, err := tls.X509KeyPair(conf.Cert, conf.Key)
		if err != nil {
			log.Log.Errorf("cannot load server cert and key：%v", err)
			return fmt.Errorf("cannot load server cert and key, %v", err)
		}

		// Creating TLS Credentials
		creds := credentials.NewTLS(&tls.Config{
			ServerName:   conf.ServerName,
			Certificates: []tls.Certificate{cert},
			RootCAs:      util.GetCAPool(),
		})
		h.opts = append(h.opts, grpc.WithTransportCredentials(creds))
	} else if conf.Insecure {
		// add the insecure option
		h.opts = append(h.opts, grpc.WithInsecure())
	}
	err := h.connect(conf)
	if err == nil {
		h.status = 0
		log.Log.Info("gRPC connected")
	}
	return err
}

func (h *headscaleRPC) ReConnect() error {
	if connecting {
		return fmt.Errorf("gRPC is trying to connect, wait for again")
	}
	connecting = true
	defer func() {
		connecting = false
	}()
	err := h.connect(common.GetHeadscaleConfig())
	if err == nil {
		h.status = 0
		log.Log.Info("gRPC connected")
	}
	return err
}

// connect is a method of the headscaleRPC struct that establishes a connection to the gRPC server.
func (h *headscaleRPC) connect(conf *model.HeadscaleConfig) error {
	retryTimes++
	if h.conn != nil {
		log.Log.Info("close gRPC connection")
		if err := h.conn.Close(); err != nil {
			return fmt.Errorf("close gRPC connection error: %v", err)
		}
		h.conn = nil
	}
	// connect
	conn, err := grpc.Dial(conf.GRPCListenAddr, h.opts...)
	if err != nil {
		return fmt.Errorf("failed to dial: %v", err)
	}
	h.conn = conn
	// init HeadscaleServiceClient
	h.HeadscaleServiceClient = pb.NewHeadscaleServiceClient(h.conn)

	_, err = h.HeadscaleServiceClient.ListApiKeys(context.Background(), &pb.ListApiKeysRequest{})
	if err != nil {
		log.Log.Errorf("Failed to ListApiKeys: %v\n", err)
	}
	if retryTimes < Times {
		time.Sleep(2 * time.Second)
		if err != nil {
			//errors.New("rpc error: code = Unavailable desc = context deadline exceeded")
			if retryTimes <= 1 && config.GetMode() < config.MULTI {
				if err.Error() == "rpc error: code = Internal desc = failed to validate token" || err.Error() == "rpc error: code = Unauthenticated desc = invalid token" {
					// if not have apikey or apikey is invalid, create an other one
					// then write it to database
					if err := h.newApiKey(conf); err != nil {
						return fmt.Errorf("init gRPC client failed to refresh api key: %v", err)
					}
				}
			}
			log.Log.Infof("retry Connection gRPC the %v times", retryTimes)
			err = h.connect(conf)
		}
	}
	retryTimes = 0
	return err
}

// GetRequestMetadata is a method of the headscaleRPC struct that returns the request metadata, including the authorization header.
func (h *headscaleRPC) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": "Bearer " + common.GetHeadscaleConfig().ApiKey,
	}, nil
}

// RequireTransportSecurity is a method of the headscaleRPC struct that returns whether the connection requires transport security (TLS).
func (h *headscaleRPC) RequireTransportSecurity() bool {
	return !common.GetHeadscaleConfig().Insecure
}

// newApiKey is a method of the headscaleRPC struct that creates a new API key and updates the configuration.
func (h *headscaleRPC) newApiKey(conf *model.HeadscaleConfig) error {
	log.Log.Info("create API KEY")
	apikey, err := h.scale.RefreshApiKey()
	if err != nil {
		return fmt.Errorf("gRPC Failed to refresh api key: %v", err)
	}
	conf.ApiKey = apikey
	config.SetApiKey(apikey)
	if err := common.DB.Model(&model.Headscale{}).Where("insecure in (true, false)").Update("api_key", apikey).Error; err != nil {
		return fmt.Errorf("gRPC Failed to update apikey: %v", err)
	}
	return nil
}

// unaryClientInterceptor is a method of the headscaleRPC struct that acts as an interceptor for unary gRPC calls, logging information about the call and handling errors.
func (h *headscaleRPC) unaryClientInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
	p := peer.Peer{}
	if opts == nil {
		opts = []grpc.CallOption{grpc.Peer(&p)}
	} else {
		opts = append(opts, grpc.Peer(&p))
	}

	start := time.Now()
	defer func() {
		//in, _ := json.Marshal(req)
		//out, _ := json.Marshal(reply)
		//inStr, outStr := string(in), string(out)
		duration := int64(time.Since(start) / time.Millisecond)

		var remoteServer string
		if p.Addr != nil {
			remoteServer = p.Addr.String()
		}
		if err != nil {
			log.Log.Errorf("grpc: %s, duration: %dms, remote_server: %s, err: %v", method, duration, remoteServer, err)
			h.status = -1
			if err.Error() == "rpc error: code = Internal desc = failed to validate token" || err.Error() == "rpc error: code = Unauthenticated desc = invalid token" {
				if err := h.newApiKey(common.GetHeadscaleConfig()); err != nil {
					log.Log.Error("gRPC connection Failed to refresh api key: ", err)
				}
			}
			return
		}
		log.Log.Infof("grpc: %s, duration: %dms, remote_server: %s", method, duration, remoteServer)
	}()

	return invoker(ctx, method, req, reply, cc, opts...)
}

//func UnaryErrorInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
//	resp, err := handler(ctx, req)
//	if err != nil {
//		log.Log.Errorf("Error: %v", err)
//		return nil, status.Errorf(codes.Internal, "Internal server error")
//	}
//	return resp, nil
//}
