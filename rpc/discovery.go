package rpc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
)

func (c *ProtocolRpcClient) GetChains(ctx context.Context) ([]Chain, error) {
	target := c.url
	if !strings.HasSuffix(strings.TrimSuffix(target.Path, "/"), "/v1") {
		return nil, errors.New("chain discovery requires service /v1 URL")
	}
	var response struct {
		APIVersion string  `json:"apiVersion"`
		Chains     []Chain `json:"chains"`
	}
	if e := c.request(ctx, target, "GET", nil, &response); e != nil {
		return nil, e
	}
	if response.APIVersion != "v1" || response.Chains == nil {
		return nil, errors.New("unsupported service response")
	}
	seen := map[string]bool{}
	for i := range response.Chains {
		ch := &response.Chains[i]
		n, e := strconv.ParseUint(ch.ChainID, 10, 64)
		if e != nil || n == 0 || strconv.FormatUint(n, 10) != ch.ChainID || seen[ch.ChainID] || ch.ProtocolVersion != "1" || ch.RPCPath != "/v1/chains/"+ch.ChainID {
			return nil, errors.New("invalid chain discovery entry")
		}
		seen[ch.ChainID] = true
		ch.RPCURL = target.String()
		u := target
		u.Path = strings.TrimSuffix(u.Path, "/") + "/chains/" + ch.ChainID
		ch.RPCURL = u.String()
	}
	return response.Chains, nil
}

func (c *ProtocolRpcClient) GetDeployment(ctx context.Context, chain uint64) (ProtocolDeployment, error) {
	var response struct {
		APIVersion      string             `json:"apiVersion"`
		ProtocolVersion string             `json:"protocolVersion"`
		Deployment      ProtocolDeployment `json:"deployment"`
	}
	if e := c.call(ctx, chain, "invisi_getDeployment", []any{}, &response); e != nil {
		return ProtocolDeployment{}, e
	}
	d := response.Deployment
	if e := d.Validate(); e != nil {
		return d, e
	}
	if response.APIVersion != "v1" || response.ProtocolVersion != "1" || d.ChainID != strconv.FormatUint(chain, 10) {
		return d, errors.New("deployment version or chain mismatch")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.deployment != nil {
		a, _ := json.Marshal(c.deployment)
		b, _ := json.Marshal(d)
		if !bytes.Equal(a, b) {
			return d, errors.New("deployment pins mismatch")
		}
	}
	copy := cloneDeployment(d)
	c.deployment = &copy
	return cloneDeployment(d), nil
}
