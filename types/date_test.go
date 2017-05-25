/*
 * Copyright (C) 2017 Dgraph Labs, Inc. and Contributors
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package types

import (
	"encoding/binary"
	"math"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/dgraph-io/dgraph/x"
)

func TestConvertDateToBool(t *testing.T) {
	dt := ValueForType(DateID)
	dt.Value = []byte{}
	_, err := Convert(dt, BoolID)
	if err == nil {
		t.Errorf("Expected error converting date to bool")
	}
}

func TestConvertDateToInt(t *testing.T) {
	data := []struct {
		in  time.Time
		out int64
	}{
		{createDate(2009, time.November, 10), 1257811200},
		{createDate(1969, time.November, 10), -4492800},
	}
	var dst Val
	var err error
	for _, tc := range data {
		bs := make([]byte, 8)
		binary.LittleEndian.PutUint64(bs, uint64(tc.in.Unix()))
		src := Val{DateID, bs}
		if dst, err = Convert(src, IntID); err != nil {
			t.Errorf("Unexpected error converting date to int: %v", err)
		} else if dst.Value.(int64) != tc.out {
			t.Errorf("Converting time to int: Expected %v, got %v", tc.out, dst.Value)
		}
	}
}

func TestConvertDateToFloat(t *testing.T) {
	data := []struct {
		in  time.Time
		out float64
	}{
		{createDate(2009, time.November, 10), 1257811200},
		{createDate(1969, time.November, 10), -4492800},
		{createDate(2039, time.November, 10), 2204496000},
		{createDate(1901, time.November, 10), -2150409600},
	}
	for _, tc := range data {
		bs := make([]byte, 8)
		binary.LittleEndian.PutUint64(bs, uint64(tc.in.Unix()))
		src := Val{DateID, bs}
		if dst, err := Convert(src, FloatID); err != nil {
			t.Errorf("Unexpected error converting date to float: %v", err)
		} else if dst.Value.(float64) != tc.out {
			t.Errorf("Converting date to float: Expected %v, got %v", tc.out, dst.Value)
		}
	}
}

func TestConvertDateToTime(t *testing.T) {
	data := []struct {
		in  time.Time
		out time.Time
	}{
		{createDate(2009, time.November, 10),
			time.Date(2009, time.November, 10, 0, 0, 0, 0, time.UTC)},
		{createDate(1969, time.November, 10),
			time.Date(1969, time.November, 10, 0, 0, 0, 0, time.UTC)},
		{createDate(2039, time.November, 10),
			time.Date(2039, time.November, 10, 0, 0, 0, 0, time.UTC)},
		{createDate(1901, time.November, 10),
			time.Date(1901, time.November, 10, 0, 0, 0, 0, time.UTC)},
	}
	for _, tc := range data {
		bs := make([]byte, 8)
		binary.LittleEndian.PutUint64(bs, uint64(tc.in.Unix()))
		src := Val{DateID, bs}
		if tout, err := Convert(src, DateTimeID); err != nil {
			t.Errorf("Unexpected error converting date to time: %v", err)
		} else if tout.Value.(time.Time) != tc.out {
			t.Errorf("Converting date to time: Expected %v, got %v", tc.out, tout.Value)
		}
	}
}

func TestConvertint64ToDate(t *testing.T) {
	data := []struct {
		in  int64
		out time.Time
	}{
		{1257811200, createDate(2009, time.November, 10)},
		{1257894000, createDate(2009, time.November, 10)}, //truncation
		{-4492800, createDate(1969, time.November, 10)},
		{0, createDate(1970, time.January, 1)},
	}
	for _, tc := range data {
		bs := make([]byte, 8)
		binary.LittleEndian.PutUint64(bs[:], uint64(tc.in))
		src := Val{IntID, bs[:]}
		if dst, err := Convert(src, DateID); err != nil {
			t.Errorf("Unexpected error converting int to date: %v", err)
		} else if dst.Value.(time.Time) != tc.out {
			t.Errorf("Converting int to date: Expected %v, got %v", tc.out, dst.Value)
		}
	}
}

func TestConvertFloatToDate(t *testing.T) {
	data := []struct {
		in  float64
		out time.Time
	}{
		{1257811200, createDate(2009, time.November, 10)},
		{1257894000.001, createDate(2009, time.November, 10)}, //truncation
		{-4492800, createDate(1969, time.November, 10)},
		{0, createDate(1970, time.January, 1)},
		{2204578800.5, createDate(2039, time.November, 10)},
		{-2150326800.12, createDate(1901, time.November, 10)},
	}
	for _, tc := range data {
		bs := make([]byte, 8)
		u := math.Float64bits(tc.in)
		binary.LittleEndian.PutUint64(bs[:], u)
		src := Val{FloatID, bs[:]}
		if dst, err := Convert(src, DateID); err != nil {
			t.Errorf("Unexpected error converting float to date: %v", err)
		} else if dst.Value.(time.Time) != tc.out {
			t.Errorf("Converting float to date: Expected %v, got %v", tc.out, dst.Value)
		}
	}
}

func TestConvertDateTimeToDate(t *testing.T) {
	data := []struct {
		in  time.Time
		out time.Time
	}{
		{time.Date(2009, time.November, 10, 0, 0, 0, 0, time.UTC),
			createDate(2009, time.November, 10)},
		{time.Date(2009, time.November, 10, 21, 2, 50, 1000000, time.UTC),
			createDate(2009, time.November, 10)}, // truncation
		{time.Date(1969, time.November, 10, 23, 0, 0, 0, time.UTC),
			createDate(1969, time.November, 10)},
	}
	for _, tc := range data {
		bs, err := tc.in.MarshalBinary()
		require.NoError(t, err)
		src := Val{DateTimeID, bs}
		if dst, err := Convert(src, DateID); err != nil {
			t.Errorf("Unexpected error converting time to date: %v", err)
		} else if dst.Value.(time.Time) != tc.out {
			t.Errorf("Converting time to date: Expected %v, got %v", tc.out, dst.Value)
		}
	}
}

func TestMain(m *testing.M) {
	x.Init()
	os.Exit(m.Run())
}
