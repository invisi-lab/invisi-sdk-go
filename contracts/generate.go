// Package contracts provides type-safe Go bindings for Invisible Protocol smart contracts,
// generated from ABI JSON files using abigen.
//
// To regenerate bindings after updating ABI files, run:
//
//	go generate ./contracts/...
package contracts

//go:generate /Users/wx/go/bin/abigen --abi abi/InvisiProtocol.abi.json --pkg contracts --type InvisiProtocol --out invisi_protocol.go
//go:generate /Users/wx/go/bin/abigen --abi abi/InvisiFactory.abi.json --pkg contracts --type InvisiFactory --out invisi_factory.go
//go:generate /Users/wx/go/bin/abigen --abi abi/VerifierRouter.abi.json --pkg contracts --type VerifierRouter --out verifier_router.go
//go:generate /Users/wx/go/bin/abigen --abi abi/IERC20.abi.json --pkg contracts --type IERC20 --out ierc20.go
