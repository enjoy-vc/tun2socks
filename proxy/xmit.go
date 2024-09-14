package proxy

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"

	"github.com/enjoy-vc/router/sdk"
	"github.com/xjasonlyu/tun2socks/v2/log"
	M "github.com/xjasonlyu/tun2socks/v2/metadata"
	"github.com/xjasonlyu/tun2socks/v2/proxy/proto"
)

var _ Proxy = (*Xmit)(nil)

type Xmit struct {
	ctx   context.Context
	addr  string
	proxy *sdk.XmitSdk
}

func NewXmitProxy(u *url.URL) (*Xmit, error) {
	address, username := u.Host, u.User.Username()
	password, _ := u.User.Password()
	log.Infof("user(%v,%v) -> %v", username, password, address)
	var (
		ctx   = context.Background()
		opt   = sdk.XmitSdkOpt{}
		proxy = sdk.NewXmitSdk(ctx, opt)
	)
	if proxy == nil {
		return nil, fmt.Errorf("new xmit failed")
	}
	var (
		xmit = &Xmit{
			ctx:   context.Background(),
			proxy: proxy,
		}
	)
	if err := xmit.proxy.Init(); err != nil {
		return nil, fmt.Errorf("init xmit failed, %v", err.Error())
	}
	if err := xmit.proxy.Login(username); err != nil {
		return nil, fmt.Errorf("login failed, %v", err.Error())
	}
	return xmit, nil
}

func (b *Xmit) Addr() string {
	return b.addr
}

func (b *Xmit) Proto() proto.Proto {
	return proto.Xmit
}

func (b *Xmit) DialContext(ctx context.Context, m *M.Metadata) (net.Conn, error) {
	return b.proxy.DialTcp(ctx, m.Addr())
}

func (b *Xmit) DialUDP(*M.Metadata) (net.PacketConn, error) {
	return nil, errors.ErrUnsupported
}
