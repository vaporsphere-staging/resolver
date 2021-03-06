// Copyright 2020 The Swarm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ensclient

import (
	"errors"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/wealdtech/go-ens/v3"

	"github.com/ethersphere/bee/pkg/swarm"
	"github.com/ethersphere/resolver/pkg/resolver"
)

// Address is the swarm bzz address.
type Address = resolver.Address

// Make sure Client implements the resolver.Client interface.
var _ resolver.Client = (*Client)(nil)

type dialFn func(string) (*ethclient.Client, error)
type resolveFn func(bind.ContractBackend, string) (string, error)

// Client is a name resolution client that can connect to ENS/RNS via an
// Ethereum or RSK node endpoint.
type Client struct {
	Endpoint  string
	ethCl     *ethclient.Client
	dialFn    dialFn
	resolveFn resolveFn
}

// Option is a function that applies an option to a Client.
type Option func(*Client)

func wrapDial(ep string) (*ethclient.Client, error) {

	// Open a connection to the ethereum node through the endpoint.
	cl, err := ethclient.Dial(ep)
	if err != nil {
		return nil, err
	}

	// Ensure the ENS resolver contract is deployed on the network we are now
	// connected to.
	if _, err := ens.PublicResolverAddress(cl); err != nil {
		return nil, err
	}

	return cl, nil
}

func wrapResolve(backend bind.ContractBackend, name string) (string, error) {

	// Connect to the ENS resolver for the provided name.
	ensR, err := ens.NewResolver(backend, name)
	if err != nil {
		return "", err
	}

	// Try and read out the content hash record.
	ch, err := ensR.Contenthash()
	if err != nil {
		return "", err
	}

	adr, err := ens.ContenthashToString(ch)
	if err != nil {
		return "", err
	}

	return adr, nil
}

// NewClient will return a new Client.
func NewClient(opts ...Option) *Client {
	c := &Client{
		dialFn:    wrapDial,
		resolveFn: wrapResolve,
	}

	// Apply all options to the Client.
	for _, o := range opts {
		o(c)
	}

	return c
}

// Connect implements the resolver.Client interface.
func (c *Client) Connect(ep string) error {
	if c.dialFn == nil {
		return errors.New("no dial function implementation")
	}

	ethCl, err := c.dialFn(ep)
	if err != nil {
		return err
	}

	c.Endpoint = ep
	c.ethCl = ethCl
	return nil
}

// Resolve implements the resolver.Client interface.
func (c *Client) Resolve(name string) (Address, error) {
	if c.resolveFn == nil {
		return swarm.ZeroAddress, errors.New("no resolve function implementation")
	}

	// Retrieve the content hash from ENS.
	hash, err := c.resolveFn(c.ethCl, name)
	if err != nil {
		return swarm.ZeroAddress, err
	}

	// Ensure that the content hash string is in a valid format, eg.
	// "/swarm/<address>".
	if !strings.HasPrefix(hash, "/swarm/") {
		return swarm.ZeroAddress, errors.New("ENS contenthash invalid")
	}

	// Trim the prefix and try to parse the result as a bzz address.
	return swarm.ParseHexAddress(strings.TrimPrefix(hash, "/swarm/"))
}

// Close closes the RPC connection with the client, terminating all unfinished
// requests.
func (c *Client) Close() {
	c.ethCl.Close()
}
