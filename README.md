## Message Broker 
Message broker built in go using gorilla mux and websockets. Followed Publish/Subscribe Architecture. Publisher can Post in Various Sections and Consumers can subscribe to various Sections and Pull the messages Published. Subscribers offset Management is there for every Section they Subscribed.

### Set-up 
- Clone this repository  
- Install the necessary packages using go.mod  
- Run the server  
```bash
make
```

### Usage
- Use curl or anyother client to publish data to a section. Auto Section creation is implemented.  
```bash
curl -X POST http:127.0.0.1:4000/pub/< your section > --data-binary < your data >
```

- Use Websocket Clients to Consume the data Published  
- Make a ws connection to **ws://127.0.0.1:3000/sub**  
- After successful upgrade to duplex, To subscribe to Sections, give a WSJson request  
```bash
{
    "Reason": "Subscribe",
    "Sections": ["foo","bar"]
}
```
- After successful subscription, Pull the messages by   
```bash
{
    "Reason": "Pull",
    "Sections": ["foo","bar"]
}
```
- You will recieve the published binary data in this form  
```bash
{
    "Status":  "Success",
	"Section": section,
	"Data":    data,
}
```
- Remember the data you recieved will be in binary   


### Future Scope 

- Porting over tcp or other protocols   
- Adding Authentication or Middleware for publish and subscribe  
- Improve overall performance  