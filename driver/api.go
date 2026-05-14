package driver

import (
	"encoding/json"
	"fmt"
	"net/http"

	spotv1 "github.com/rackspace-spot/spot-go-sdk/api/v1"
)

// ServeAPI starts the HTTP lookup API server on addr (e.g. ":8080").
// The server proxies region and server-class queries to the Rackspace Spot API
// using credentials supplied per-request, so the browser never needs to reach
// the Rackspace endpoints directly.
func ServeAPI(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/lookup", handleLookup)
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	return http.ListenAndServe(addr, corsMiddleware(mux))
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type lookupRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type lookupResponse struct {
	Regions       []regionResult      `json:"regions"`
	ServerClasses []serverClassResult `json:"serverClasses"`
}

type regionResult struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type serverClassResult struct {
	Name          string `json:"name"`
	Region        string `json:"region"`
	CPU           string `json:"cpu"`
	Memory        string `json:"memory"`
	GPU           string `json:"gpu,omitempty"`
	MarketPrice   string `json:"marketPrice"`
	MinBidPrice   string `json:"minBidPrice"`
	OnDemandPrice string `json:"onDemandPrice"`
}

func handleLookup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "POST required"})
		return
	}

	var req lookupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RefreshToken == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "refreshToken is required"})
		return
	}

	ctx := r.Context()

	c, err := spotv1.NewSpotClient(&spotv1.Config{RefreshToken: req.RefreshToken})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("client init: %v", err)})
		return
	}
	if _, err := c.Authenticate(ctx); err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": fmt.Sprintf("authentication failed: %v", err)})
		return
	}

	regions, err := c.ListRegions(ctx)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": fmt.Sprintf("regions: %v", err)})
		return
	}

	scList, err := c.ListServerClasses(ctx, "")
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": fmt.Sprintf("server classes: %v", err)})
		return
	}

	resp := lookupResponse{
		Regions:       make([]regionResult, 0, len(regions)),
		ServerClasses: make([]serverClassResult, 0, len(scList.Items)),
	}
	for _, r := range regions {
		resp.Regions = append(resp.Regions, regionResult{Name: r.Name, Description: r.Description})
	}
	for _, sc := range scList.Items {
		resp.ServerClasses = append(resp.ServerClasses, serverClassResult{
			Name:          sc.Name,
			Region:        sc.Region,
			CPU:           sc.Resources.CPU,
			Memory:        sc.Resources.Memory,
			GPU:           sc.Resources.GPU,
			MarketPrice:   sc.CurrentMarketPricePerHour,
			MinBidPrice:   sc.MinBidPricePerHour,
			OnDemandPrice: sc.OnDemandPricePerHour,
		})
	}

	writeJSON(w, http.StatusOK, resp)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}
