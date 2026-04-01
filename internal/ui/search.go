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
	"miniflux.app/v2/internal/ui/view"
)

func (h *handler) showSearchPage(w http.ResponseWriter, r *http.Request) {
	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		response.HTMLServerError(w, r, err)
		return
	}

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
	offset := request.QueryIntParam(r, "offset", 0)

	var entries model.Entries
	var entriesCount int

	if searchQuery != "" {
		builder := h.store.NewEntryQueryBuilder(user.ID)
		builder.WithSearchQuery(searchQuery, useLikeSearch, titleOnly)
		builder.WithCategoryIDs(categoryIDs)
		if unreadOnly {
			builder.WithStatus(model.EntryStatusUnread)
		}
		builder.WithoutContent()

		entriesCount, err = builder.CountEntries()
		if err != nil {
			response.HTMLServerError(w, r, err)
			return
		}

		if offset >= entriesCount {
			offset = 0
		}

		builder.WithOffset(offset)
		builder.WithLimit(user.EntriesPerPage)

		entries, err = builder.GetEntries()
		if err != nil {
			response.HTMLServerError(w, r, err)
			return
		}
	}

	view := view.New(h.tpl, r)
	pagination := getPagination(h.routePath("/search"), entriesCount, offset, user.EntriesPerPage)
	pagination.SearchQuery = searchQuery
	pagination.UnreadOnly = unreadOnly
	pagination.LikeSearch = useLikeSearch
	pagination.TitleOnly = titleOnly
	pagination.Categories = categoryIDs

	view.Set("searchQuery", searchQuery)
	view.Set("searchUnreadOnly", unreadOnly)
	view.Set("searchUseLike", useLikeSearch)
	view.Set("searchTitleOnly", titleOnly)
	view.Set("searchSelectedCategories", categoryIDs)
	selectedCategorySet := make(map[int64]bool, len(categoryIDs))
	for _, id := range categoryIDs {
		selectedCategorySet[id] = true
	}
	view.Set("searchSelectedCategorySet", selectedCategorySet)
	view.Set("entries", entries)
	view.Set("total", entriesCount)
	view.Set("pagination", pagination)
	view.Set("menu", "search")
	view.Set("user", user)
	view.Set("countUnread", h.store.CountUnreadEntries(user.ID))
	view.Set("countErrorFeeds", h.store.CountUserFeedsWithErrors(user.ID))
	view.Set("hasSaveEntry", h.store.HasSaveEntry(user.ID))
	categories, err := h.store.Categories(user.ID)
	if err != nil {
		response.HTMLServerError(w, r, err)
		return
	}
	view.Set("categories", categories)

	response.HTML(w, r, view.Render("search"))
}
