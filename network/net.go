package main

import (
  "fmt"
  "log"
  "net"
)

func main() {
  interfaces, err := net.Interfaces()

  if err != nil {
    log.Fatal(err)
  }

  for _, iface := range interfaces {
    fmt.Println(iface.Name)

    addrs, err := iface.MulticastAddrs()
    if err != nil {
      log.Fatal(err)
    }

    for i, addr := range addrs {
      fmt.Printf("\t%d: %s\n", i, addr)
    }
  }

}
