// Copyright 2020 The Swarm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ensclient

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethersphere/resolver/pkg/resolver"
)

func GetDialFunc(cl *Client) func(ep string) (*ethclient.Client, error) {
	return cl.dialFn
}

// WithDialFunc will set the Dial function implementaton.
func WithDialFunc(fn func(ep string) (*ethclient.Client, error)) Option {
	return func(c *Client) {
		c.dialFn = fn
	}
}

func WithErrorDialFunc(err error) Option {
	return WithDialFunc(func(ep string) (*ethclient.Client, error) {
		return nil, err
	})
}

func WithNoopDialFunc() Option {
	return WithDialFunc(func(ep string) (*ethclient.Client, error) {
		return &ethclient.Client{}, nil
	})
}

// WithResolveFunc will set the Resolve function implementation.
func WithResolveFunc(fn func(backend bind.ContractBackend, input string) (common.Address, error)) Option {
	return func(c *Client) {
		c.resolveFn = fn
	}
}

func WithErrorResolveFunc(err error) Option {
	return WithResolveFunc(func(backend bind.ContractBackend, input string) (common.Address, error) {
		return resolver.Address{}, err
	})
}

func WithAdrResolveFunc(adr resolver.Address) Option {
	return WithResolveFunc(func(backend bind.ContractBackend, input string) (common.Address, error) {
		return adr, nil
	})
}
