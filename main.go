package main

func main() {
	cfg := &Config{
		ProduceListenAddr:     ":3000",
		ConsumerListenAddr:    ":4000",
		ConsumerListenAddrTCP: ":5000",
		StoreFunc:             produce,
	}
	server := Newserver(cfg)
	server.Serve()

}
