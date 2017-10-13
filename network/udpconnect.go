package main

import (
  "fmt"
  "golang.org/x/net/ipv4"
  "net"
)

https://www.socketloop.com/tutorials/golang-udp-client-server-read-write-example

func main() {
  ipv4Addr := &net.UDPAddr{IP: net.IPv4(255, 255, 255, 255), Port: 8089}
  conn, err := net.ListenUDP("udp4", ipv4Addr)
  if err != nil {
    fmt.Printf("ListenUDP error %v\n", err)
    return
  }

  pc := ipv4.NewPacketConn(conn)

  // assume your have a interface named enp2s0
  iface, err := net.InterfaceByName("enp2s0")
  if err != nil {
    fmt.Printf("can't find specified interface %v\n", err)
    return
  }

  if err := pc.JoinGroup(iface, &net.UDPAddr{IP: net.IPv4(255, 255, 255, 255)}); err != nil {
    fmt.Printf("%v\n", err)
    return
  }

  // test
  if loop, err := pc.MulticastLoopback(); err == nil {
    fmt.Printf("MulticastLoopback status:%v\n", loop)
    if !loop {
      if err := pc.SetMulticastLoopback(true); err != nil {
        fmt.Printf("SetMulticastLoopback error:%v\n", err)
      }
    }
  }

  if _, err := conn.WriteTo([]byte("hello"), ipv4Addr); err != nil {
    fmt.Printf("Write failed, %v\n", err)
  }

  buf := make([]byte, 1024)
  for {
    if n, addr, err := conn.ReadFrom(buf); err != nil {
      fmt.Printf("error %v", err)
    } else {
      fmt.Printf("recv %s from %v\n", string(buf[:n]), addr)
    }
  }

  return
}
