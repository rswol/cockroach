// Copyright 2022 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package tree_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/cockroachdb/cockroach/pkg/sql/randgen"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondatapb"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/randutil"
)

func BenchmarkAsJSON(b *testing.B) {
	// Use fixed seed so that each invocation of this benchmark
	// produces exactly the same types, and datums streams.
	// This number can be changed to an arbitrary value; doing so
	// would result in new types/datums being produced.
	rng := randutil.NewTestRandWithSeed(-4365865412074131521)

	const numDatums = 1024
	makeDatums := func(typ *types.T) tree.Datums {
		const allowNulls = true
		res := make(tree.Datums, numDatums)
		for i := 0; i < numDatums; i++ {
			res[i] = randgen.RandDatum(rng, typ, allowNulls)
		}
		return res
	}

	bench := func(b *testing.B, typ *types.T) {
		b.ReportAllocs()
		b.StopTimer()
		datums := makeDatums(typ)
		b.StartTimer()

		for i := 0; i < b.N; i++ {
			_, err := tree.AsJSON(datums[i%numDatums], sessiondatapb.DataConversionConfig{}, time.UTC)
			if err != nil {
				b.Fatal(err)
			}
		}
	}

	for _, typ := range testTypes(rng) {
		b.Run(typ.String(), func(b *testing.B) {
			bench(b, typ)
		})

		if randgen.IsAllowedForArray(typ) {
			typ = types.MakeArray(typ)
			b.Run(typ.String(), func(b *testing.B) {
				bench(b, typ)
			})
		}
	}
}

// testTypes returns list of types to test against.
func testTypes(rng *rand.Rand) (typs []*types.T) {
	for _, typ := range randgen.SeedTypes {
		switch typ {
		case types.AnyTuple:
		// Ignore AnyTuple -- it's not very interesting; we'll generate test tuples below.
		case types.RegClass, types.RegNamespace, types.RegProc, types.RegProcedure, types.RegRole, types.RegType:
		// Ignore a bunch of pseudo-OID types (just want regular OID).
		case types.Geometry, types.Geography:
		// Ignore geometry/geography: these types are insanely inefficient;
		// AsJson(Geo) -> MarshalGeo -> go JSON bytes ->  ParseJSON -> Go native -> json.JSON
		// Benchmarking this generates too much noise.
		// TODO(yevgeniy): fix this.
		default:
			typs = append(typs, typ)
		}
	}

	// Add tuple types.
	var tupleTypes []*types.T
	makeTupleType := func() *types.T {
		contents := make([]*types.T, rng.Intn(6)) // Up to 6 fields
		for i := range contents {
			contents[i] = randgen.RandTypeFromSlice(rng, typs)
		}
		candidateTuple := types.MakeTuple(contents)
		// Ensure tuple type is unique.
		for _, t := range tupleTypes {
			if t.Equal(candidateTuple) {
				return nil
			}
		}
		tupleTypes = append(tupleTypes, candidateTuple)
		return candidateTuple
	}

	const numTupleTypes = 5
	for i := 0; i < numTupleTypes; i++ {
		var typ *types.T
		for typ == nil {
			typ = makeTupleType()
		}
		typs = append(typs, typ)
	}

	return typs
}
