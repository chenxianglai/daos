//
// (C) Copyright 2019-2022 Intel Corporation.
//
// SPDX-License-Identifier: BSD-2-Clause-Patent
//

package common

import (
	"os"
	"strconv"
	"strings"
)

// ScrubEnvironment modifies the environment variables set for
// this process and any children which inherit its environment
// by unsetting any variables supplied in the blocklist.
func ScrubEnvironment(blocklist []string) {
	for _, key := range blocklist {
		os.Unsetenv(key)
	}
}

// ScrubEnvironmentExcept modifies the environment variables set for
// this process and any children which inherit its environment
// by unsetting any variables that are not supplied in the allowlist.
func ScrubEnvironmentExcept(allowlist []string) {
	lookup := make(map[string]struct{})
	for _, key := range allowlist {
		lookup[key] = struct{}{}
	}

	for _, keyVal := range os.Environ() {
		key := strings.SplitN(keyVal, "=", 2)[0]
		if _, inList := lookup[key]; !inList {
			os.Unsetenv(key)
		}
	}
}

// DisableProxyScrubEnv disabled removal of proxy variables from the process
// environment.
const DisableProxyScrubEnv = "DAOS_DISABLE_PROXY_SCRUB"

func proxyScrubIsDisabled() bool {
	val, set := os.LookupEnv(DisableProxyScrubEnv)
	if !set {
		return false
	}

	disabled, err := strconv.ParseBool(val)
	if err != nil {
		return false
	}
	return disabled
}

// ScrubProxyVariables removes proxy variables from the process environment.
func ScrubProxyVariables() {
	proxyVars := []string{
		"http_proxy", "https_proxy", "ftp_proxy", "socks_proxy", "no_proxy",
		"HTTP_PROXY", "HTTPS_PROXY", "FTP_PROXY", "SOCKS_PROXY", "NO_PROXY",
	}

	if !proxyScrubIsDisabled() {
		ScrubEnvironment(proxyVars)
	}
}

// BoolAsInt converts a bool to an int.
func BoolAsInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// MergeEnvVars merges and deduplicates two slices of environment
// variables. Conflicts are resolved by taking the value from the
// second list.
func MergeEnvVars(curVars []string, newVars []string) (merged []string) {
	mergeMap := make(map[string]string)
	for _, pair := range curVars {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) != 2 || kv[0] == "" || kv[1] == "" {
			continue
		}
		// strip duplicates in curVars; shouldn't be any
		// but this will ensure it.
		if _, found := mergeMap[kv[0]]; found {
			continue
		}
		mergeMap[kv[0]] = kv[1]
	}

	mergedKeys := make(map[string]struct{})
	for _, pair := range newVars {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) != 2 || kv[0] == "" || kv[1] == "" {
			continue
		}
		// strip duplicates in newVars
		if _, found := mergedKeys[kv[0]]; found {
			continue
		}
		mergedKeys[kv[0]] = struct{}{}
		mergeMap[kv[0]] = kv[1]
	}

	merged = make([]string, 0, len(mergeMap))
	for key, val := range mergeMap {
		merged = append(merged, strings.Join([]string{key, val}, "="))
	}

	return
}
