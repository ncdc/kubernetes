/*
Copyright 2022 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cache

// Scope represents a subdivision of a cache's key space.
type Scope interface {
	// Name returns the name of the scope.
	Name() string
	// ListAllIndexValue returns the value used with ListAllIndex to retrieve all items in this scope.
	ListAllIndexValue() string
	// CacheKey transforms original to a format suitable for use with Indexer#GetByKey for this scope.
	CacheKey(original string) string
}

// ScopeFromKeyFunc is a function that returns a Scope based on key.
type ScopeFromKeyFunc func(key string) (Scope, error)

// unnamedScope implements Scope and has no name.
type unnamedScope struct{}

// Name returns the name of the scope.
func (u *unnamedScope) Name() string {
	return ""
}

// ListAllIndexValue returns the value used with ListAllIndex to retrieve all items in this scope.
func (u *unnamedScope) ListAllIndexValue() string {
	return ""
}

// CacheKey transforms original to a format suitable for use with Indexer#GetByKey for this scope.
func (u *unnamedScope) CacheKey(original string) string {
	return original
}

// unnamedScopeFromKey always returns an unnamedScope regardless of key.
func unnamedScopeFromKey(key string) (Scope, error) {
	return &unnamedScope{}, nil
}
