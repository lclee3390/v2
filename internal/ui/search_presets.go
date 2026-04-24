// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"errors"
	"net/http"
	"strings"

	"miniflux.app/v2/internal/http/response"
)

func (h *handler) addSearchPreset(w http.ResponseWriter, r *http.Request) {
	keyword := strings.TrimSpace(r.FormValue("keyword"))
	if keyword == "" {
		response.HTMLRedirect(w, r, h.routePath("/search"))
		return
	}
	if globalPresetStore == nil {
		response.HTMLServerError(w, r, errors.New("search preset store not initialised"))
		return
	}

	if err := globalPresetStore.add(keyword); err != nil {
		response.HTMLServerError(w, r, err)
		return
	}

	response.HTMLRedirect(w, r, h.routePath("/search"))
}

func (h *handler) removeSearchPreset(w http.ResponseWriter, r *http.Request) {
	keyword := strings.TrimSpace(r.FormValue("keyword"))
	if keyword == "" {
		response.HTMLRedirect(w, r, h.routePath("/search"))
		return
	}
	if globalPresetStore == nil {
		response.HTMLServerError(w, r, errors.New("search preset store not initialised"))
		return
	}

	if err := globalPresetStore.remove(keyword); err != nil {
		response.HTMLServerError(w, r, err)
		return
	}

	response.HTMLRedirect(w, r, h.routePath("/search"))
}
