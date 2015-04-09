package main

import (
	"encoding/json"
	"fmt"
	"gopkg.in/mgo.v2"
	"net"
	"os"
)

type Person struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

func main() {
	var (
		host   = "115.156.249.17"
		port   = "12345"
		remote = host + ":" + port
		data   = make([]byte, 1024)
	)
	fmt.Println("Initiating server...")

	lis, err := net.Listen("tcp", remote)
	defer lis.Close()

	if err != nil {
		fmt.Println("Error when listen: ", remote)
		os.Exit(-1)
	}

	for {
		var res string
		conn, err := lis.Accept()
		if err != nil {
			fmt.Println("Error accepting client: ", err.Error())
			os.Exit(0)
		}

		go func(con net.Conn) {
			fmt.Println("New connection: ", con.RemoteAddr())
			for {
				length, err := con.Read(data)
				if err != nil {
					fmt.Printf("Client %v quit.\n", con.RemoteAddr())
					con.Close()
					return
				}

				res = string(data[0:length])
				//打印接收到的字符串
				fmt.Printf("%s said: %s", con.RemoteAddr(), res)

				//解析json
				var people Person
				if err := json.Unmarshal([]byte(res), &people); err == nil {
					fmt.Println(people.Name)
				}

				//将此数据插入数据库
				session, err := mgo.Dial("127.0.0.1:27017")
				if err != nil {
					panic(err)
				}
				defer session.Close()

				session.SetMode(mgo.Monotonic, true)

				//拿到需要使用的集合
				collection := session.DB("test").C("person")

				//插入数据
				collection.Insert(&Person{people.Id, people.Name})

				con.Write([]byte(res))
			}
		}(conn)
	}
}
