package server

import (
	"fmt"
	"github.com/gliderlabs/ssh"
	"net"
	"strconv"
	"sync"
)
import gossh "golang.org/x/crypto/ssh"

const (
	forwardedTCPChannelType = "forwarded-tcpip"
)

type remoteForwardRequest struct {
	BindAddr string
	BindPort uint32
}

type remoteForwardChannelData struct {
	DestAddr   string
	DestPort   uint32
	OriginAddr string
	OriginPort uint32
}

type remoteForwardSuccess struct {
	BindPort uint32
}

type remoteForwardCancelRequest struct {
	BindAddr string
	BindPort uint32
}

type Balancer struct {
	forwards map[string]net.Listener

	sync.Mutex
}

func (h *Balancer) HandleSSHRequest(ctx ssh.Context, srv *ssh.Server, req *gossh.Request) (bool, []byte) {
	h.Lock()
	if h.forwards == nil {
		h.forwards = make(map[string]net.Listener)
	}
	h.Unlock()
	conn := ctx.Value(ssh.ContextKeyConn).(*gossh.ServerConn)
	switch req.Type {
	case "tcpip-forward":
		var reqPayload remoteForwardRequest
		if err := gossh.Unmarshal(req.Payload, &reqPayload); err != nil {
			// TODO: log parse failure
			return false, []byte{}
		}
		if srv.ReversePortForwardingCallback == nil || !srv.ReversePortForwardingCallback(ctx, reqPayload.BindAddr, reqPayload.BindPort) {
			return false, []byte("port forwarding is disabled")
		}
		addr := net.JoinHostPort(reqPayload.BindAddr, strconv.Itoa(int(reqPayload.BindPort)))
		fmt.Println(addr)

		h.AdjustNewForward(ctx, addr, conn, reqPayload)

		// ln, err := net.Listen("tcp", ":"+strconv.Itoa(int(reqPayload.BindPort)))
		//
		// if err != nil {
		// 	// TODO: log listen failure
		// 	return false, []byte{}
		// }
		// _, destPortStr, _ := net.SplitHostPort(ln.Addr().String())
		// destPort, _ := strconv.Atoi(destPortStr)
		// h.Lock()
		// h.forwards[addr] = ln
		// h.Unlock()
		// go func() {
		// 	<-ctx.Done()
		// 	h.Lock()
		// 	ln, ok := h.forwards[addr]
		// 	h.Unlock()
		// 	if ok {
		// 		ln.Close()
		// 	}
		// }()
		//
		// go func() {
		// 	for {
		// 		c, err := ln.Accept()
		// 		if err != nil {
		// 			// TODO: log accept failure
		// 			break
		// 		}
		//
		// 		originAddr, orignPortStr, _ := net.SplitHostPort(c.RemoteAddr().String())
		// 		originPort, _ := strconv.Atoi(orignPortStr)
		// 		payload := gossh.Marshal(&remoteForwardChannelData{
		// 			DestAddr:   reqPayload.BindAddr,
		// 			DestPort:   uint32(destPort),
		// 			OriginAddr: originAddr,
		// 			OriginPort: uint32(originPort),
		// 		})
		// 		go func() {
		// 			ch, reqs, err := conn.OpenChannel(forwardedTCPChannelType, payload)
		// 			if err != nil {
		// 				// TODO: log failure to open channel
		// 				log.Println(err)
		// 				c.Close()
		// 				return
		// 			}
		// 			go gossh.DiscardRequests(reqs)
		// 			go func() {
		// 				defer ch.Close()
		// 				defer c.Close()
		// 				writt, _ := io.Copy(ch, c)
		// 				log.Printf("%d bytes were written to channel", writt)
		// 			}()
		// 			go func() {
		// 				defer ch.Close()
		// 				defer c.Close()
		// 				writt, _ := io.Copy(c, ch)
		// 				log.Printf("%d bytes were written from channel", writt)
		// 			}()
		// 		}()
		// 	}
		// 	h.Lock()
		// 	delete(h.forwards, addr)
		// 	h.Unlock()
		// }()
		return true, gossh.Marshal(&remoteForwardSuccess{reqPayload.BindPort})

	case "cancel-tcpip-forward":
		var reqPayload remoteForwardCancelRequest
		if err := gossh.Unmarshal(req.Payload, &reqPayload); err != nil {
			// TODO: log parse failure
			return false, []byte{}
		}
		addr := net.JoinHostPort(reqPayload.BindAddr, strconv.Itoa(int(reqPayload.BindPort)))
		h.Lock()
		ln, ok := h.forwards[addr]
		h.Unlock()
		if ok {
			ln.Close()
		}
		return true, nil
	default:
		return false, nil
	}
}
