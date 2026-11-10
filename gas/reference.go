// Package gas exposes measured benchmark references independently of transaction estimation.
package gas

import shared "github.com/invisi-lab/invisi-sdk-go/protocol/gas"

type Shape = shared.Shape
type Source = shared.Source
type ReleaseIdentity = shared.ReleaseIdentity
type Measurement = shared.Measurement
type Catalog = shared.Catalog
type Result = shared.Result

func DocumentedCatalog() Catalog           { return shared.DocumentedCatalog() }
func Reference(shape Shape) (Result, bool) { return shared.Reference(shape) }
