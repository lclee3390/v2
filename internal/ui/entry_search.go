// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"net/http"
	"strconv"
	"strings"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/storage"
	"miniflux.app/v2/internal/ui/view"
)

func (h *handler) showSearchEntryPage(w http.ResponseWriter, r *http.Request) {
	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		response.HTMLServerError(w, r, err)
		return
	}

	entryID := request.RouteInt64Param(r, "entryID")
	searchQuery := request.QueryStringParam(r, "q", "")
	unreadOnly := true
	if request.HasQueryParam(r, "unread") {
		unreadOnly = request.QueryBoolParam(r, "unread", false)
	}
	useLikeSearch := true
	if request.HasQueryParam(r, "like") {
		useLikeSearch = request.QueryBoolParam(r, "like", false)
	}
	titleOnly := true
	if request.HasQueryParam(r, "title_only") {
		titleOnly = request.QueryBoolParam(r, "title_only", false)
	}
	categoryParams := request.QueryStringParamList(r, "categories")
	var categoryIDs []int64
	for _, cat := range categoryParams {
		if id, err := strconv.ParseInt(strings.TrimSpace(cat), 10, 64); err == nil && id > 0 {
			categoryIDs = append(categoryIDs, id)
		}
	}
	builder := h.store.NewEntryQueryBuilder(user.ID)
	builder.WithSearchQuery(searchQuery, useLikeSearch, titleOnly)
	builder.WithCategoryIDs(categoryIDs)
	builder.WithEntryID(entryID)

	entry, err := builder.GetEntry()
	if err != nil {
		response.HTMLServerError(w, r, err)
		return
	}

	if entry == nil {
		response.HTMLNotFound(w, r)
		return
	}

	if entry.ShouldMarkAsReadOnView(user) {
		err = h.store.SetEntriesStatus(user.ID, []int64{entry.ID}, model.EntryStatusRead)
		if err != nil {
			response.HTMLServerError(w, r, err)
			return
		}

		entry.Status = model.EntryStatusRead
	}

	if user.AlwaysOpenExternalLinks {
		response.HTMLRedirect(w, r, entry.URL)
		return
	}

	entryPaginationBuilder := storage.NewEntryPaginationBuilder(h.store, user.ID, entry.ID, user.EntryOrder, user.EntryDirection)
	entryPaginationBuilder.WithSearchQuery(searchQuery, useLikeSearch, titleOnly)
	entryPaginationBuilder.WithCategoryIDs(categoryIDs)
	if unreadOnly {
		if entry.Status == model.EntryStatusRead {
			entryPaginationBuilder.WithStatusOrEntryID(model.EntryStatusUnread, entry.ID)
		} else {
			entryPaginationBuilder.WithStatus(model.EntryStatusUnread)
		}
	}

	prevEntry, nextEntry, err := entryPaginationBuilder.Entries()
	if err != nil {
		response.HTMLServerError(w, r, err)
		return
	}

	nextEntryRoute := ""
	if nextEntry != nil {
		nextEntryRoute = h.routePath("/search/entry/%d", nextEntry.ID)
	}

	prevEntryRoute := ""
	if prevEntry != nil {
		prevEntryRoute = h.routePath("/search/entry/%d", prevEntry.ID)
	}

	view := view.New(h.tpl, r)
	view.Set("searchQuery", searchQuery)
	view.Set("searchUnreadOnly", unreadOnly)
	view.Set("searchUseLike", useLikeSearch)
	view.Set("searchTitleOnly", titleOnly)
	view.Set("searchSelectedCategories", categoryIDs)
	view.Set("entry", entry)
	view.Set("prevEntry", prevEntry)
	view.Set("nextEntry", nextEntry)
	view.Set("nextEntryRoute", nextEntryRoute)
	view.Set("prevEntryRoute", prevEntryRoute)
	view.Set("menu", "search")
	view.Set("user", user)
	view.Set("countUnread", h.store.CountUnreadEntries(user.ID))
	view.Set("countErrorFeeds", h.store.CountUserFeedsWithErrors(user.ID))
	view.Set("hasSaveEntry", h.store.HasSaveEntry(user.ID))

	response.HTML(w, r, view.Render("entry"))
}
