package main

func main() {
	cfg := &Config{
		ListenAddr: ":3000",
	}
	server := Newserver(cfg)
	server.Serve()

}
