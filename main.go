package main

var host = "localhost"
var port = "9000"

func main() {
	dbErr := db.InitializeDB()
	if dbErr != nil {
		panic(dbErr)
	}
	jwtErr := myJwt.InitializeJWT()
	if jwtErr != nil {
		panic(jwtErr)
	}
	serverErr := server.StartServer(host, port)
	if serverErr != nil {
		panic(serverErr)
	}

}
