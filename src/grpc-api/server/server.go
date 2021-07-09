package server

import (
	"errors"
	"fmt"
	"net"
	"os"

	gc "github.com/cloud-barista/cb-ladybug/src/grpc-api/common"
	"github.com/cloud-barista/cb-ladybug/src/grpc-api/config"
	"github.com/cloud-barista/cb-ladybug/src/grpc-api/logger"
	pb "github.com/cloud-barista/cb-ladybug/src/grpc-api/protobuf/cbladybug"
	grpc_mcar "github.com/cloud-barista/cb-ladybug/src/grpc-api/server/mcar"

	"google.golang.org/grpc/reflection"
)

// RunServer - LADYBUG GRPC 서버 실행
func RunServer() {
	logger := logger.NewLogger()

	configPath := os.Getenv("APP_ROOT") + "/conf/grpc_conf.yaml"
	gConf, err := configLoad(configPath)
	if err != nil {
		logger.Error("failed to load config : ", err)
		return
	}

	ladybugsrv := gConf.GSL.LadybugSrv

	conn, err := net.Listen("tcp", ladybugsrv.Addr)
	if err != nil {
		logger.Error("failed to listen: ", err)
		return
	}

	cbserver, closer, err := gc.NewCBServer(ladybugsrv)
	if err != nil {
		logger.Error("failed to create grpc server: ", err)
		return
	}

	if closer != nil {
		defer closer.Close()
	}

	gs := cbserver.Server
	pb.RegisterMCARServer(gs, &grpc_mcar.MCARService{})

	if ladybugsrv.Reflection == "enable" {
		if ladybugsrv.Interceptors != nil && ladybugsrv.Interceptors.AuthJWT != nil {
			fmt.Printf("\n\n*** you can run reflection when jwt auth interceptor is not used ***\n\n")
		} else {
			reflection.Register(gs)
		}
	}

	fmt.Printf("\n[CB-Ladybug: Multi-Cloud Application Management Framework]")
	fmt.Printf("\n   Initiating GRPC API Server....__^..^__....")
	fmt.Printf("\n\n => grpc server started on %s\n\n", ladybugsrv.Addr)

	if err := gs.Serve(conn); err != nil {
		logger.Error("failed to serve: ", err)
	}
}

func configLoad(cf string) (config.GrpcConfig, error) {
	logger := logger.NewLogger()

	// Viper 를 사용하는 설정 파서 생성
	parser := config.MakeParser()

	var (
		gConf config.GrpcConfig
		err   error
	)

	if cf == "" {
		logger.Error("Please, provide the path to your configuration file")
		return gConf, errors.New("configuration file are not specified")
	}

	logger.Debug("Parsing configuration file: ", cf)
	if gConf, err = parser.GrpcParse(cf); err != nil {
		logger.Error("ERROR - Parsing the configuration file.\n", err.Error())
		return gConf, err
	}

	// Command line 에 지정된 옵션을 설정에 적용 (우선권)

	// LADYBUG 필수 입력 항목 체크
	ladybugsrv := gConf.GSL.LadybugSrv

	if ladybugsrv == nil {
		return gConf, errors.New("ladybugsrv field are not specified")
	}

	if ladybugsrv.Addr == "" {
		return gConf, errors.New("ladybugsrv.addr field are not specified")
	}

	if ladybugsrv.TLS != nil {
		if ladybugsrv.TLS.TLSCert == "" {
			return gConf, errors.New("ladybugsrv.tls.tls_cert field are not specified")
		}
		if ladybugsrv.TLS.TLSKey == "" {
			return gConf, errors.New("ladybugsrv.tls.tls_key field are not specified")
		}
	}

	if ladybugsrv.Interceptors != nil {
		if ladybugsrv.Interceptors.AuthJWT != nil {
			if ladybugsrv.Interceptors.AuthJWT.JWTKey == "" {
				return gConf, errors.New("ladybugsrv.interceptors.auth_jwt.jwt_key field are not specified")
			}
		}
		if ladybugsrv.Interceptors.PrometheusMetrics != nil {
			if ladybugsrv.Interceptors.PrometheusMetrics.ListenPort == 0 {
				return gConf, errors.New("ladybugsrv.interceptors.prometheus_metrics.listen_port field are not specified")
			}
		}
		if ladybugsrv.Interceptors.Opentracing != nil {
			if ladybugsrv.Interceptors.Opentracing.Jaeger != nil {
				if ladybugsrv.Interceptors.Opentracing.Jaeger.Endpoint == "" {
					return gConf, errors.New("ladybugsrv.interceptors.opentracing.jaeger.endpoint field are not specified")
				}
			}
		}
	}

	return gConf, nil
}
