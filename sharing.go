package main

import (
	"encoding/hex"
	"strings"
)

// The route token carries the organization without placing its identifier in the URL.
const shareRouteScope = "!share-route-v1"

func (s *Store) shareToken(org, id string) (string, error) {
	if org == pilotOrg {
		return id, nil
	}
	sealed, err := encrypt(s.key, shareRouteScope, []byte(org+"\n"+id))
	if err != nil {
		return "", err
	}
	return "v1-" + hex.EncodeToString(sealed), nil
}

func (s *Store) shareTarget(route string) (org, id string, ok bool) {
	if len(route) == 48 {
		_, err := hex.DecodeString(route)
		return pilotOrg, route, err == nil
	}
	encoded, found := strings.CutPrefix(route, "v1-")
	if !found || len(encoded) > 240 {
		return "", "", false
	}
	sealed, err := hex.DecodeString(encoded)
	if err != nil {
		return "", "", false
	}
	raw, err := decrypt(s.key, shareRouteScope, sealed)
	if err != nil {
		return "", "", false
	}
	org, id, ok = strings.Cut(string(raw), "\n")
	if !ok || !accountIdentifier.MatchString(org) || len(id) != 48 {
		return "", "", false
	}
	_, err = hex.DecodeString(id)
	return org, id, err == nil
}
