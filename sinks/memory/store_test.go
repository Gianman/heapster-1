// Copyright 2015 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package memory

import (
	"testing"
	"time"

	"github.com/GoogleCloudPlatform/heapster/Godeps/_workspace/src/github.com/stretchr/testify/require"
	"github.com/stretchr/testify/assert"
)

func TestInitialization(t *testing.T) {
	store := NewStore(time.Second, time.Microsecond)
	assert.Empty(t, store.Get(time.Now().Add(-time.Minute), time.Now()))
}

func TestNilInsert(t *testing.T) {
	store := NewStore(time.Minute, time.Hour)
	assert.Error(t, store.Put(time.Now(), nil))
}

func TestOlderInsert(t *testing.T) {
	store := NewStore(time.Minute, time.Hour)
	now := time.Now()
	assert.NoError(t, store.Put(now, 1))
	assert.NoError(t, store.Put(now.Add(-time.Second), 2))
	actual := store.Get(now.Add(-time.Millisecond), now.Add(time.Millisecond))
	require.Len(t, actual, 1)
	assert.Equal(t, 1, actual[0].(int))
}

func TestGC(t *testing.T) {
	store := NewStore(time.Microsecond, time.Second)
	now := time.Now()
	for i := 0; i < 100; i++ {
		assert.NoError(t, store.Put(time.Now(), struct{}{}))
	}
	time.Sleep(time.Millisecond)
	assert.Len(t, store.Get(now, time.Now()), 0)
}

func TestGCDetail(t *testing.T) {
	store := NewStore(time.Second, time.Second)
	now := time.Now()
	for i := 0; i < 20; i++ {
		assert.NoError(t, store.Put(time.Now(), struct{}{}))
		time.Sleep(100 * time.Millisecond)
	}
	assert.NotEqual(t, len(store.Get(now, time.Now())), 0)
}

func TestLongGC(t *testing.T) {
	store := NewStore(time.Hour, time.Second)
	now := time.Now()
	for i := 0; i < 200; i++ {
		assert.NoError(t, store.Put(time.Now(), i))
	}
	assert.Equal(t, 200, len(store.Get(now, time.Now())))
}

func TestPut(t *testing.T) {
	store := NewStore(time.Hour, time.Second)
	expected := []int{}
	for i := 0; i < 100; i++ {
		expected = append(expected, i)
		store.Put(time.Now(), i)
	}
	actual := store.Get(time.Now().Add(-time.Minute), time.Now())
	for idx, val := range actual {
		assert.Equal(t, expected[idx], val.(int))
	}
}
