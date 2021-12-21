# go-gps-receiver
GPS receiver from serial port

* PUB
  * Build then execute or Run \
`go build .` \
`go run main.go`
  * Choose port name
  * Send the message (zmq)

* SUB
  * Receive the message (zmq)
  * Unpack to object (msgpack)
  * Read the 'Content' property
