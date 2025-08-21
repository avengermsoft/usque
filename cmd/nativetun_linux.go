//go:build linux

package cmd

import (
	"fmt"
	"log"
	"net"

	"github.com/Diniboy1123/usque/api"
	"github.com/Diniboy1123/usque/config"
	"github.com/songgao/water"
	"github.com/vishvananda/netlink"
)

var longDescription = "Expose Warp as a native TUN device that accepts any IP traffic." +
	" Requires root, tun.ko, and iproute2."

func (t *tunDevice) create() (api.TunnelDevice, error) {
	platformSpecificParams := water.PlatformSpecificParams{
		Name: t.name,
		Persist: t.name != "",
	}

	dev, err := water.New(water.Config{DeviceType: water.TUN, PlatformSpecificParams: platformSpecificParams})
	if err != nil {
		return nil, err
	}

	t.name = dev.Name()

	if t.iproute2 {
		link, err := netlink.LinkByName(dev.Name())
		if err != nil {
			return nil, fmt.Errorf("failed to get link: %v", err)
		}

		if err := netlink.LinkSetMTU(link, t.mtu); err != nil {
			return nil, fmt.Errorf("failed to set MTU: %v", err)
		}
		if t.ipv4 {
			if err := netlink.AddrAdd(link, &netlink.Addr{
				IPNet: &net.IPNet{
					IP:   net.ParseIP(config.AppConfig.IPv4),
					Mask: net.CIDRMask(32, 32),
				}}); err != nil {
				return nil, fmt.Errorf("failed to add IPv4 address: %v", err)
			}
		}
		if t.ipv6 {
			if err := netlink.AddrAdd(link, &netlink.Addr{
				IPNet: &net.IPNet{
					IP:   net.ParseIP(config.AppConfig.IPv6),
					Mask: net.CIDRMask(128, 128),
				}}); err != nil {
				return nil, fmt.Errorf("failed to add IPv6 address: %v", err)
			}
		}
		if err := netlink.LinkSetUp(link); err != nil {
			return nil, fmt.Errorf("failed to set link up: %v", err)
		}
	} else {
		log.Println("Skipping IP address and link setup. You should set the link up manually.")
		log.Println("Config has the following IP addresses:")
		log.Printf("IPv4: %s", config.AppConfig.IPv4)
		log.Printf("IPv6: %s", config.AppConfig.IPv6)
	}

	return api.NewWaterAdapter(dev), nil
}
