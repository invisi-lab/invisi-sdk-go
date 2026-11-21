package broadcaster

import (
	_ "embed"
	"encoding/json"
)

//go:embed testdata/router-gas-profile.json
var routerGasProfileJSON []byte
var routerGasProfile = func() struct {
	ModeOverhead                                                                               [4]uint64
	V3Base, V3PerHop, Permit, Authorization, RouteByte, MaxHops, MaxRouteBytes, MaxSafeInteger uint64
} {
	var p struct {
		ModeOverhead                                                                               [4]uint64
		V3Base, V3PerHop, Permit, Authorization, RouteByte, MaxHops, MaxRouteBytes, MaxSafeInteger uint64
	}
	if e := json.Unmarshal(routerGasProfileJSON, &p); e != nil {
		panic(e)
	}
	return p
}()

// RouterGasInput contains model quantities, never RPC receipt measurements.
// ZeroExGasUnits must come from trusted independent route policy.
type RouterGasInput struct {
	CoreGasUnits       uint64 `json:"coreGasUnits"`
	Mode               uint8  `json:"mode"`
	Channel            uint8  `json:"channel"`
	HopCount           uint16 `json:"hopCount"`
	RouteBytes         uint64 `json:"routeBytes"`
	Permit             bool   `json:"permit"`
	AuthorizationCount uint8  `json:"authorizationCount"`
	ZeroExGasUnits     uint64 `json:"zeroExGasUnits,omitempty"`
}
type RouterGasBreakdown struct {
	Core          uint64 `json:"core"`
	Mode          uint64 `json:"router"`
	Route         uint64 `json:"swap"`
	Permit        uint64 `json:"permit"`
	Authorization uint64 `json:"authorization"`
	Calldata      uint64 `json:"routeCalldata"`
}
type RouterGasResult struct {
	Available bool                `json:"available"`
	Source    string              `json:"source"`
	ModelID   string              `json:"modelId"`
	GasUnits  uint64              `json:"gasUnits,omitempty"`
	Breakdown *RouterGasBreakdown `json:"breakdown,omitempty"`
	Reason    string              `json:"reason,omitempty"`
}

const RouterGasModelID = "payment-router-gas-v1"
const routerMaxSafe = uint64(9007199254740991)

// EstimateRouterGas adds conservative engineering allowances to the Core preview,
// counting the common transaction base once. Final outer simulation is required.
func EstimateRouterGas(i RouterGasInput) RouterGasResult {
	r := RouterGasResult{Source: "model", ModelID: RouterGasModelID}
	if i.CoreGasUnits == 0 || i.CoreGasUnits > routerMaxSafe || i.Mode > 3 || i.Channel > 2 || i.RouteBytes > routerGasProfile.MaxRouteBytes || i.AuthorizationCount > 1 || (i.Mode != 3 && i.AuthorizationCount != 0) || (i.Channel != 1 && i.HopCount != 0) || (i.Channel == 1 && (i.HopCount < 1 || uint64(i.HopCount) > routerGasProfile.MaxHops)) || (i.Channel == 0 && i.RouteBytes != 0) || i.ZeroExGasUnits > routerMaxSafe {
		r.Reason = "invalid-router-gas-input"
		return r
	}
	b := RouterGasBreakdown{Core: i.CoreGasUnits, Mode: routerGasProfile.ModeOverhead[i.Mode], Calldata: i.RouteBytes * routerGasProfile.RouteByte, Authorization: uint64(i.AuthorizationCount) * routerGasProfile.Authorization}
	if i.Permit {
		b.Permit = routerGasProfile.Permit
	}
	if i.Channel == 1 {
		b.Route = routerGasProfile.V3Base + routerGasProfile.V3PerHop*uint64(i.HopCount)
	}
	if i.Channel == 2 {
		if i.ZeroExGasUnits == 0 {
			r.Reason = "zeroex-route-gas-unavailable"
			return r
		}
		b.Route = i.ZeroExGasUnits
	}
	total := b.Core + b.Mode + b.Route + b.Permit + b.Authorization + b.Calldata
	if total > routerMaxSafe {
		r.Reason = "router-gas-overflow"
		return r
	}
	r.Available = true
	r.GasUnits = total
	r.Breakdown = &b
	return r
}
