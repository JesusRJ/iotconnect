/* UDPConnect
 * Identifica o IP usando broadcast na rede
 * ttps://www.socketloop.com/tutorials/golang-udp-client-server-read-write-example
 */
package main

import (
  "fmt"
  "log"
  "net"
  "strings"
)

// Default ip address
var ipservice, portNum string = "", "8089"

func Announce() {
  laddr, err := net.ResolveUDPAddr("udp", ":0")
  if err != nil {
    log.Fatal(err)
  }

  maddr, err := net.ResolveUDPAddr("udp4", "255.255.255.255:8089")
  if err != nil {
    log.Fatal(err)
  }

  c, err := net.ListenUDP("udp4", laddr)
  if err != nil {
    log.Fatal(err)
  }

  defer c.Close()

  for ipservice == "" {
    _, err := c.WriteToUDP([]byte("cmd=ping"), maddr)
    if err != nil {
      log.Fatal(err)
    }

    buf := make([]byte, 1024)

    if n, addr, err := c.ReadFromUDP(buf); err != nil {
      log.Fatal(err)
    } else {
      ipservice = getip(string(buf[:n]))
      fmt.Printf("Service Addres %s from %v\n", ipservice, addr)
    }

  }
}

// cmd=pong&sta_ip=&host_ip=192.168.4.1
// cmd=pong&sta_ip=192.168.0.6&host_ip=192.168.4.1
func getip(text string) string {
  sub := strings.Split(text, "&")[1:]
  ip := strings.Split(sub[0], "=")[1]
  if ip == "" {
    ip = strings.Split(sub[1], "=")[1]
  }
  return ip
}

func main() {
  Announce()

  service := ipservice + ":" + portNum

  RemoteAddr, err := net.ResolveUDPAddr("udp", service)

  conn, err := net.DialUDP("udp", nil, RemoteAddr)
  if err != nil {
    log.Fatal(err)
  }

  log.Printf("Established connection to %s \n", service)
  log.Printf("Remote UDP address : %s \n", conn.RemoteAddr().String())
  log.Printf("Local UDP client address : %s \n", conn.LocalAddr().String())

  defer conn.Close()

  // write a message to server
  // message := []byte("cmd=ping")
  message := []byte("cmd=control&d=1")

  _, err = conn.Write(message)

  if err != nil {
    log.Println(err)
  }

  // receive message from server
  buffer := make([]byte, 1024)
  n, addr, err := conn.ReadFromUDP(buffer)

  fmt.Println("UDP Server : ", addr)
  fmt.Println("Received from UDP server : ", string(buffer[:n]))

}
