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
