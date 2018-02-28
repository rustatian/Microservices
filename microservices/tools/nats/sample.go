package nats

//handler := nats.NewServer(
//endpoint.HashNatsEnpoint,
//decodeUppercaseRequest,
//encodeResponse,
//5,
//10,
//5,
//time.Millisecond*10,
//nil,
//)
//
//nc, _ := gonats.Connect(gonats.DefaultURL)
//nc.QueueSubscribe("111", "111", handler.MsgHandler)

//func encodeResponse(_ context.Context, response interface{}) (r []byte, err error) {
//	resp := response.(hashResponse)
//	data, err := json.Marshal(resp)
//	return data, err
//}

//NATS Encode/decode
//func encodeHashResponse(ctx context.Context, response interface{}) (r []byte, err error) {
//	resp := response.(hashResponse)
//	data, err := json.Marshal(resp)
//	return data, err
//}
//
//func decodeHashRequest(ctx context.Context, msg *gonats.Msg) (interface{}, error) {
//	var request hashRequest
//	if err := json.Unmarshal(msg.Data, &request); err != nil {
//		return nil, err
//	}
//	return request, nil
//}
//
//func encodeValidateResponse(ctx context.Context, response interface{}) (r []byte, err error) {
//	resp := response.(validateResponse)
//	data, err := json.Marshal(resp)
//	return data, err
//}
//
//func decodeValidateRequest(ctx context.Context, msg *gonats.Msg) (interface{}, error) {
//	var request validateRequest
//	if err := json.Unmarshal(msg.Data, &request); err != nil {
//		return nil, err
//	}
//	return request, nil
//}

//func MakeVaultNatsHandler(endpoint Endpoints, logger log.Logger, natsUrl string, err chan error) http.Handler {
//	//Used for health checks and metrics
//	r := mux.NewRouter()
//	options := []httptransport.ServerOption{
//		httptransport.ServerErrorLogger(logger),
//		httptransport.ServerErrorEncoder(encodeError),
//	}
//
//	nc, er := gonats.Connect(natsUrl)
//	if er != nil {
//		err <- er
//	}
//	nc, er = gonats.Connect(natsUrl,
//		gonats.DisconnectHandler(func(nc *gonats.Conn) {
//			fmt.Printf("Got disconnected!\n")
//		}),
//		gonats.ReconnectHandler(func(_ *gonats.Conn) {
//			fmt.Printf("Got reconnected to %v!\n", nc.ConnectedUrl())
//		}),
//		gonats.ClosedHandler(func(nc *gonats.Conn) {
//			fmt.Printf("Connection closed. Reason: %q\n", nc.LastError())
//		}),
//	)
//
//	if er != nil {
//		err <- er
//	}
//
//	hashHandler := nats.NewServer(
//		endpoint.HashEndpoint,
//		decodeHashRequest,
//		encodeHashResponse,
//		10,
//		runtime.GOMAXPROCS(12),
//		10,
//		time.Microsecond*1,
//		err,
//	)
//	nc.QueueSubscribe("hash", "hash", hashHandler.MsgHandler)
//
//	validateHandler := nats.NewServer(
//		endpoint.ValidateEndpoint,
//		decodeValidateRequest,
//		encodeValidateResponse,
//		10,
//		runtime.GOMAXPROCS(12),
//		10,
//		time.Microsecond*1,
//		err,
//	)
//	nc.QueueSubscribe("validate", "validate", validateHandler.MsgHandler)
//
//	//GET /health
//	r.Methods("GET").Path("/health").Handler(httptransport.NewServer(
//		endpoint.VaultHealthEndpoint,
//		decodeHTTPHealthRequest,
//		encodeHTTPHealthResponse,
//		options...,
//	))
//	r.Path("/metrics").Handler(stdprometheus.Handler())
//
//	return r
//}

//r := vault.MakeVaultNatsHandler(endpoints, logger, "nats://172.24.231.70:4222", errCh)
