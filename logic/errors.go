package main

import (
	"errors"
)

var (
	ErrRouter         = errors.New("router rpc is not available")
	ErrDecodeKey      = errors.New("decode key error")
	ErrNetworkAddr    = errors.New("network addrs error, must network@address")
	ErrConnectArgs    = errors.New("connect rpc args error")
	ErrDisconnectArgs = errors.New("disconnect rpc args error")

	ErrGetGuid       = errors.New("not found the guid")
	ErrGetChannelId  = errors.New("not found the channel id")
	ErrUidCannotZero = errors.New("The uid cannot be 0")
)
