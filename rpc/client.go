package rpc

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type ClientOptions struct {
	HTTPClient *http.Client
	Headers    http.Header
	Deployment *ProtocolDeployment
}

type ProtocolRpcClient struct {
	url        url.URL
	http       *http.Client
	headers    http.Header
	sequence   atomic.Uint64
	mu         sync.Mutex
	deployment *ProtocolDeployment
	provider   ChainProvider
	pinned     bool
}

func NewProtocolRpcClient(endpoint string, options ...ClientOptions) (*ProtocolRpcClient, error) {
	u, e := url.Parse(endpoint)
	if e != nil || u == nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil || u.Fragment != "" {
		return nil, errors.New("invalid Protocol RPC URL")
	}
	if len(options) > 1 {
		return nil, errors.New("expected at most one client option")
	}
	var opt ClientOptions
	if len(options) == 1 {
		opt = options[0]
	}
	client := opt.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	clone := *client
	clone.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	c := &ProtocolRpcClient{url: *u, http: &clone, headers: opt.Headers.Clone()}
	if opt.Deployment != nil {
		if e := opt.Deployment.Validate(); e != nil {
			return nil, e
		}
		d := cloneDeployment(*opt.Deployment)
		if d.Registry != nil {
			r := *d.Registry
			d.Registry = &r
		}
		c.deployment = &d
		c.pinned = true
	}
	return c, nil
}

func (c *ProtocolRpcClient) WithProvider(provider ChainProvider) *ProtocolRpcClient {
	c.mu.Lock()
	defer c.mu.Unlock()
	var d *ProtocolDeployment
	if c.deployment != nil {
		copy := cloneDeployment(*c.deployment)
		d = &copy
	}
	return &ProtocolRpcClient{url: c.url, http: c.http, headers: c.headers.Clone(), deployment: d, provider: provider, pinned: c.pinned}
}

func (c *ProtocolRpcClient) context(context DeploymentContext) (ProtocolDeployment, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.deployment == nil || c.deployment.ChainID != strconv.FormatUint(context.ChainID, 10) || c.deployment.Protocol.Address != context.ProtocolAddress {
		return ProtocolDeployment{}, errors.New("deployment not loaded or context mismatch")
	}
	return cloneDeployment(*c.deployment), nil
}

func cloneDeployment(d ProtocolDeployment) ProtocolDeployment {
	if d.Registry != nil {
		registry := *d.Registry
		d.Registry = &registry
	}
	if d.BroadcasterRegistry != nil {
		r := *d.BroadcasterRegistry
		d.BroadcasterRegistry = &r
	}
	return d
}
