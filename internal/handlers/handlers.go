package handlers

import (
	"fmt"
	"net/http"
)

func Protected(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "request allowed")
}
