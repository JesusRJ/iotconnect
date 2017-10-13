package main

import (
  "bufio"
  "fmt"
  "golang.org/x/net/ipv4"
  "log"
  "net"
  "os"
)

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
  scanner := bufio.NewScanner(os.Stdin)
  for scanner.Scan() {
    {
      text := scanner.Text()
      _, err := c.WriteToUDP([]byte(text), maddr)
      if err != nil {
        log.Fatal(err)
      }
    }
  }
}

func Listen() {
  c, err := net.ListenPacket("udp4", "0.0.0.0:8089") // mDNS over UDP
  if err != nil {
    log.Fatal(err)
  }
  defer c.Close()
  p := ipv4.NewPacketConn(c)

  ifaces, err := net.Interfaces()
  if err != nil {
    log.Fatal(err)
  }
  //TODO lets user choose iface or select one with internet access.
  //TODO OR! listen on all ports (need agreed range).

  //TODO add layer of abstraction for game mechanics input output
  for _, iface := range ifaces {
    fmt.Printf("%v\n", iface.Name)
    go listenOnSpecificInterface(p, &iface)
  }
  // for {
  // }
}

func listenOnSpecificInterface(p *ipv4.PacketConn, iface *net.Interface) {
  mDNSLinkLocal := net.UDPAddr{IP: net.IPv4(224, 0, 0, 251)}
  if err := p.JoinGroup(iface, &mDNSLinkLocal); err != nil {
    log.Fatal(err)
  }
  defer p.LeaveGroup(iface, &mDNSLinkLocal)
  if err := p.SetControlMessage(ipv4.FlagDst, true); err != nil {
    log.Fatal(err)
  }

  b := make([]byte, 1500)
  for {
    fmt.Printf("Waiting for data on interface: %v\n", iface.Name)
    n, cm, peer, err := p.ReadFrom(b)
    fmt.Printf("Received on interface %v: %v \n", iface.Name, string(b[:n]))
    if err != nil {
      log.Fatal(err)
    }
    if !cm.Dst.IsMulticast() || !cm.Dst.Equal(mDNSLinkLocal.IP) {
      continue
    }
    answers := []byte("FAKE-MDNS-ANSWERS") // fake mDNS answers, you need to implement this
    if _, err := p.WriteTo(answers, nil, peer); err != nil {
      log.Fatal(err)
    }
  }
}

func main() {
  // Listen()
  Announce()
}
