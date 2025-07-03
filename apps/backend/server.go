package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/stripe/stripe-go/v82/product"

	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/checkout/session"
)

func main() {
	// This is your test secret API key.
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	http.Handle("/", http.FileServer(http.Dir("public")))
	http.HandleFunc("/create-checkout-session", createCheckoutSession)
	http.HandleFunc("/list-products", listProducts)

	addr := os.Getenv("DOMAIN")
	log.Printf("Listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func createCheckoutSession(w http.ResponseWriter, r *http.Request) {
	domain := os.Getenv("DOMAIN")
	params := &stripe.CheckoutSessionParams{
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			&stripe.CheckoutSessionLineItemParams{
				// Provide the exact Price ID (for example, price_1234) of the product you want to sell
				Price:    stripe.String("price_1Rgea6PeS473W0olEhHSN41q"),
				Quantity: stripe.Int64(1),
			},
		},
		Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL: stripe.String(domain + "?success=true"),
		CancelURL:  stripe.String(domain + "?canceled=true"),
	}

	s, err := session.New(params)

	if err != nil {
		log.Printf("session.New: %v", err)
	}

	http.Redirect(w, r, s.URL, http.StatusSeeOther)
}

// listProducts handles GET /list-products and returns a list of Stripe products as JSON
func listProducts(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle preflight request
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	params := &stripe.ProductListParams{
		Active: stripe.Bool(true), // Optional: filter for active products
	}
	params.Limit = stripe.Int64(10)
	iter := product.List(params)

	var products []interface{}
	for iter.Next() {
		p := iter.Product()
		// Marshal the full product object to map[string]interface{}
		var m map[string]interface{}
		b, err := json.Marshal(p)
		if err != nil {
			continue
		}
		if err := json.Unmarshal(b, &m); err != nil {
			continue
		}
		products = append(products, m)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}
