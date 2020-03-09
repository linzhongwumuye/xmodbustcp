package main

import (
	"fmt"
	"math/rand"
	"net"
)

func main() {
	con, err := net.Dial("tcp4", "127.0.0.1:12789")
	if err != nil {
		fmt.Println(err)
	}
	data1 := []byte{0x19, 0x18, 0x00, 0x00, 0x00, 0x06, 0x00, 0x03, 0x00, 0x00, 0x00, 0x16}

	for {
		data2 := make([]byte, 512)
		data1[9] = byte(rand.Intn(10) + 1)
		data1[11] = byte(rand.Intn(80) + 1)
		_, err := con.Write(data1)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Printf("\ndata1\t")
		for _, da := range data1 {
			fmt.Printf("%02x\t", da)
		}
		fmt.Printf("\n")

		n, err := con.Read(data2)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("n : ", n)
		data2 = data2[:n]

		fmt.Printf("data2\t")
		for _, da := range data2 {
			fmt.Printf("%02x\t", da)
		}
		fmt.Printf("\n")
	}
}
