package handlers

// Export internal fields for testing only.
// This file is only compiled during tests (the _test.go convention applies
// to files in the same package, not the external _test package).

import "letsencrypt-manager/models"

// Expose store field so handler tests can inject state directly.
func (h *Handler) TestStore() *models.Store {
	return h.store
}
