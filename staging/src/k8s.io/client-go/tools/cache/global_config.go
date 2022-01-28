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

// GlobalConfig contains global configuration for caches, indexers, keys, etc.
type GlobalConfig struct {
	// ObjectKeyFunc is the global KeyFunc.
	ObjectKeyFunc KeyFunc
	// DecodeKeyFunc is the global DecodeKeyFunc.
	DecodeKeyFunc DecodeKeyFunc

	// NamespaceIndexFunc is the global IndexFunc for NamespaceIndex.
	NamespaceIndexFunc IndexFunc
	// NamespaceNameKeyFunc is the global function that encodes a namespace and a name as a single key as a string.
	NamespaceNameKeyFunc func(namespace, name string) string

	// ScopeFromKeyFunc is the global ScopeFromKeyFunc.
	ScopeFromKeyFunc ScopeFromKeyFunc
}

// completedGlobalConfig is private to protect against direct access/modifications by consumers outside this package.
type completedGlobalConfig struct {
	GlobalConfig
}

var (
	// globalConfigSet is a channel that guards against multiple attempts to set the global configuration.
	globalConfigSet = make(chan struct{})

	// globalConfig contains the global configuration. Initial default values suitable for most use cases
	// are configured here.
	globalConfig = &completedGlobalConfig{
		GlobalConfig: GlobalConfig{
			ObjectKeyFunc:        DeletionHandlingMetaNamespaceKeyFunc,
			DecodeKeyFunc:        DecodeMetaNamespaceKey,
			NamespaceIndexFunc:   MetaNamespaceIndexFunc,
			NamespaceNameKeyFunc: namespaceNameToKey,
			ScopeFromKeyFunc:     unnamedScopeFromKey,
		},
	}
)

// SetGlobalConfig sets the global cache configuration. It may only be called once. Subsequent invocations
// will panic.
func SetGlobalConfig(c GlobalConfig) {
	// This will panic if called twice
	close(globalConfigSet)

	globalConfig = &completedGlobalConfig{
		GlobalConfig: c,
	}
}

// ObjectKey returns the key for obj using the global ObjectKeyFunc.
func ObjectKey(obj interface{}) (string, error) {
	return globalConfig.ObjectKeyFunc(obj)
}

// DecodeKey decodes key into a QueueKey using the global DecodeKeyFunc.
func DecodeKey(key string) (QueueKey, error) {
	return globalConfig.DecodeKeyFunc(key)
}

// NamespaceIndexFunc returns the IndexFunc to be used for NamespaceIndex using the global NamespaceIndexFunc.
func NamespaceIndexFunc() IndexFunc {
	return globalConfig.NamespaceIndexFunc
}

// NamespaceNameKey encodes namespace and name into a key as a string using the global NamespaceNameKeyFunc.
func NamespaceNameKey(namespace, name string) string {
	return globalConfig.NamespaceNameKeyFunc(namespace, name)
}

// ScopeFromKey returns the Scope for key using the global ScopeFromKeyFunc.
func ScopeFromKey(key string) (Scope, error) {
	return globalConfig.ScopeFromKeyFunc(key)
}
