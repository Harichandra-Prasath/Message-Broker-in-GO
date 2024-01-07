package main

func main() {
	cfg := &Config{
		ListenAddr: ":3000",
		StoreFunc:  produce,
	}
	server := Newserver(cfg)
	server.Serve()

}
