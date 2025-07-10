package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/stripe/stripe-go/v82"
	billingportal "github.com/stripe/stripe-go/v82/billingportal/session"
	"github.com/stripe/stripe-go/v82/checkout/session"
	"github.com/stripe/stripe-go/v82/price"
	"github.com/stripe/stripe-go/v82/webhook"
)

func main() {
	// Load .env file if present
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or error loading .env file, relying on environment variables.")
	}

	// This is your test secret API key.
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	http.Handle("/", http.FileServer(http.Dir("public")))
	http.HandleFunc("/create-checkout-session", createCheckoutSession)
	http.HandleFunc("/list-prices", listPrices)

	http.HandleFunc("/checkout-sessions", listCheckoutSessions) // Fixed route (removed backticks)
	http.HandleFunc("/expire-checkout-session", expireCheckoutSessionHandler)
	http.HandleFunc("/get-open-session", getOpenSessionHandler)
	http.HandleFunc("/create-customer-portal-session", createCustomerPortalSession)
	http.HandleFunc("/webhook", handleWebhook)

	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":4242"
	}
	log.Printf("Listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func createCheckoutSession(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Accept product and customer data from frontend
	var req struct {
		PriceID    string `json:"priceId"`
		ProductID  string `json:"productId"`
		CustomerID string `json:"customerId"` // Optional, if you want to use an existing customer
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.PriceID == "" {
		http.Error(w, "Missing or invalid priceId", http.StatusBadRequest)
		return
	}

	// Fetch the price and expand the product
	pr, err := price.Get(
		req.PriceID,
		&stripe.PriceParams{
			Expand: []*string{stripe.String("product")},
		},
	)
	if err != nil {
		log.Printf("price.Get: %v", err)
		http.Error(w, "Invalid priceId", http.StatusBadRequest)
		return
	}

	if b, err := json.MarshalIndent(pr, "", "  "); err == nil {
		log.Printf("Product details: %s", b)
	} else {
		log.Printf("Failed to marshal product details: %v", err)
	}

	// // Create a Stripe customer first (optional, but recommended for tracking)
	// var customerID string
	// customerParams := &stripe.CustomerParams{}
	// cust, err := customer.New(customerParams)
	// if err != nil {
	// 	log.Printf("customer.New: %v", err)
	// 	http.Error(w, "Failed to create customer", http.StatusInternalServerError)
	// 	return
	// }
	// customerID = cust.ID

	customerID := "cus_SbwlgcHF9QwIIi"

	// Check for existing open session for this customer
	if req.CustomerID != "" {
		customerID = req.CustomerID
	}

	existingSession, err := getOpenCheckoutSession(customerID)
	if err != nil {
		log.Printf("Error checking for existing session: %v", err)
	} else if existingSession != nil {
		// Check if the existing session has the same price
		if existingSession.LineItems != nil && len(existingSession.LineItems.Data) > 0 {
			for _, item := range existingSession.LineItems.Data {
				if item.Price != nil && item.Price.ID == req.PriceID {
					// Found an open session for this priceId, reuse it
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(map[string]string{"url": existingSession.URL})
					return
				}
			}
		}
		// If existing session has different price, expire it first
		log.Printf("Expiring existing session %s to create new one", existingSession.ID)
		if expireErr := expireCheckoutSession(existingSession.ID); expireErr != nil {
			log.Printf("Failed to expire existing session: %v", expireErr)
		}
	}

	var mode string
	if pr.Recurring != nil {
		mode = string(stripe.CheckoutSessionModeSubscription)
	} else {
		mode = string(stripe.CheckoutSessionModePayment) // Default to payment if not recurring
	}

	successURL := os.Getenv("SUCCESS_URL")
	cancelURL := os.Getenv("CANCEL_URL")

	params := &stripe.CheckoutSessionParams{
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(req.PriceID),
				Quantity: stripe.Int64(1),
			},
		},
		ExpiresAt:  stripe.Int64(time.Now().Add(1 * time.Hour).Unix()), // Session expires in 1 hour
		Customer:   stripe.String(customerID),                          // Optional, can be empty if not using a customer
		Mode:       stripe.String(mode),
		SuccessURL: stripe.String(successURL),
		CancelURL:  stripe.String(cancelURL),
	}

	s, err := session.New(params)
	if err != nil {
		log.Printf("session.New: %v", err)
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	// Return the session URL as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"url": s.URL})
}

// listPrices handles GET /list-prices and returns a list of Stripe prices with expanded product data as JSON
func listPrices(w http.ResponseWriter, r *http.Request) {
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

	params := &stripe.PriceListParams{
		Expand: []*string{
			stripe.String("data.product"), // Expand product to get product details
		},
		Active: stripe.Bool(true), // Only get active prices
	}
	params.Limit = stripe.Int64(10)
	iter := price.List(params)

	var prices []interface{}
	for iter.Next() {
		p := iter.Price()
		// Marshal the full price object to map[string]interface{}
		var m map[string]interface{}

		if !p.Product.Active {
			continue // Skip inactive products
		}

		b, err := json.Marshal(p)
		if err != nil {
			continue
		}
		if err := json.Unmarshal(b, &m); err != nil {
			continue
		}
		prices = append(prices, m)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prices)
}

// listCheckoutSessions handles GET /checkout-sessions?customer={customer_id} and returns all checkout sessions for a customer
func listCheckoutSessions(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	customerID := r.URL.Query().Get("customer")
	if customerID == "" {
		http.Error(w, "Missing customer parameter", http.StatusBadRequest)
		return
	}

	params := &stripe.CheckoutSessionListParams{
		Customer: stripe.String(customerID),
		Status:   stripe.String("open"),
	}
	params.Limit = stripe.Int64(20)
	iter := session.List(params)

	var sessions []interface{}
	for iter.Next() {
		s := iter.CheckoutSession()
		var m map[string]interface{}
		b, err := json.Marshal(s)
		if err != nil {
			continue
		}
		if err := json.Unmarshal(b, &m); err != nil {
			continue
		}
		sessions = append(sessions, m)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sessions)
}

// getOpenSessionHandler handles GET /get-open-session?customer={customer_id} and returns the open checkout session for a customer
func getOpenSessionHandler(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	customerID := r.URL.Query().Get("customer")
	if customerID == "" {
		http.Error(w, "Missing customer parameter", http.StatusBadRequest)
		return
	}

	session, err := getOpenCheckoutSession(customerID)
	if err != nil {
		log.Printf("Error getting open session: %v", err)
		http.Error(w, "Failed to get open session", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if session == nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"session": nil})
	} else {
		var m map[string]interface{}
		b, err := json.Marshal(session)
		if err != nil {
			http.Error(w, "Failed to marshal session", http.StatusInternalServerError)
			return
		}
		if err := json.Unmarshal(b, &m); err != nil {
			http.Error(w, "Failed to unmarshal session", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"session": m})
	}
}

func getOpenCheckoutSession(customerID string) (*stripe.CheckoutSession, error) {
	params := &stripe.CheckoutSessionListParams{
		Customer: stripe.String(customerID),
		Status:   stripe.String("open"),
		ListParams: stripe.ListParams{
			Limit: stripe.Int64(1), // Limit to 1 session for efficiency
		},
	}

	iter := session.List(params)
	if iter.Next() {
		return iter.CheckoutSession(), nil
	}

	if err := iter.Err(); err != nil {
		return nil, err
	}

	return nil, nil // No open session found
}

func expireCheckoutSession(sessionID string) error {
	params := &stripe.CheckoutSessionExpireParams{}
	_, err := session.Expire(sessionID, params)
	if err != nil {
		log.Printf("session.Expire: %v", err)
	}

	return err
}

// expireCheckoutSessionHandler handles POST /expire-checkout-session and expires a checkout session
func expireCheckoutSessionHandler(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		SessionID string `json:"sessionId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.SessionID == "" {
		http.Error(w, "Missing or invalid sessionId", http.StatusBadRequest)
		return
	}

	err := expireCheckoutSession(req.SessionID)
	if err != nil {
		log.Printf("Failed to expire checkout session: %v", err)
		http.Error(w, "Failed to expire checkout session", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "expired"})
}

// handleWebhook handles POST /webhook and processes Stripe webhook events
func handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	const MaxBodyBytes = int64(65536)
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	// Get the webhook signing secret from environment
	endpointSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	if endpointSecret == "" {
		log.Printf("Warning: STRIPE_WEBHOOK_SECRET not set, skipping signature verification")
	}

	// Verify the webhook signature
	event := stripe.Event{}
	if endpointSecret != "" {
		signatureHeader := r.Header.Get("Stripe-Signature")
		event, err = webhook.ConstructEvent(payload, signatureHeader, endpointSecret)
		if err != nil {
			log.Printf("Webhook signature verification failed: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	} else {
		// If no secret is set, just parse the event (not recommended for production)
		if err := json.Unmarshal(payload, &event); err != nil {
			log.Printf("Error parsing webhook JSON: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	// Handle the event
	switch event.Type {
	case "checkout.session.completed":
		handleCheckoutSessionCompleted(event)
	case "checkout.session.expired":
		handleCheckoutSessionExpired(event)
	case "payment_intent.succeeded":
		handlePaymentIntentSucceeded(event)
	case "payment_intent.payment_failed":
		handlePaymentIntentFailed(event)
	case "invoice.payment_succeeded":
		handleInvoicePaymentSucceeded(event)
	case "invoice.payment_failed":
		handleInvoicePaymentFailed(event)
	default:
		log.Printf("Unhandled event type: %s", event.Type)
	}

	w.WriteHeader(http.StatusOK)
}

func handleCheckoutSessionCompleted(event stripe.Event) {
	var session stripe.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
		log.Printf("Error parsing checkout session completed event: %v", err)
		return
	}

	log.Printf("Checkout session completed: %s", session.ID)
	log.Printf("Customer: %s", session.Customer.ID)
	log.Printf("Payment Status: %s", session.PaymentStatus)
	log.Printf("Amount Total: %d %s", session.AmountTotal, session.Currency)

	// Here you can add your business logic for successful payments
	// For example:
	// - Update user's subscription status
	// - Send confirmation email
	// - Update database records
	// - Grant access to paid content

	if session.PaymentStatus == "paid" {
		log.Printf("Payment successful for session: %s", session.ID)
		// Add your success logic here
	}
}

func handleCheckoutSessionExpired(event stripe.Event) {
	var session stripe.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
		log.Printf("Error parsing checkout session expired event: %v", err)
		return
	}

	log.Printf("Checkout session expired: %s", session.ID)
	log.Printf("Customer: %s", session.Customer.ID)

	// Here you can add logic for expired sessions
	// For example:
	// - Clean up temporary data
	// - Send abandonment email
	// - Update analytics
}

func handlePaymentIntentSucceeded(event stripe.Event) {
	var paymentIntent stripe.PaymentIntent
	if err := json.Unmarshal(event.Data.Raw, &paymentIntent); err != nil {
		log.Printf("Error parsing payment intent succeeded event: %v", err)
		return
	}

	log.Printf("Payment intent succeeded: %s", paymentIntent.ID)
	log.Printf("Amount: %d %s", paymentIntent.Amount, paymentIntent.Currency)
	log.Printf("Customer: %s", paymentIntent.Customer.ID)

	// Here you can add your business logic for successful payments
	// This is for one-time payments
}

func handlePaymentIntentFailed(event stripe.Event) {
	var paymentIntent stripe.PaymentIntent
	if err := json.Unmarshal(event.Data.Raw, &paymentIntent); err != nil {
		log.Printf("Error parsing payment intent failed event: %v", err)
		return
	}

	log.Printf("Payment intent failed: %s", paymentIntent.ID)
	log.Printf("Amount: %d %s", paymentIntent.Amount, paymentIntent.Currency)
	log.Printf("Customer: %s", paymentIntent.Customer.ID)
	log.Printf("Last payment error: %v", paymentIntent.LastPaymentError)

	// Here you can add your business logic for failed payments
	// For example:
	// - Send failure notification
	// - Update user's payment status
	// - Retry payment logic
	// - Log for analytics
}

func handleInvoicePaymentSucceeded(event stripe.Event) {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		log.Printf("Error parsing invoice payment succeeded event: %v", err)
		return
	}

	log.Printf("Invoice payment succeeded: %s", invoice.ID)
	log.Printf("Customer: %s", invoice.Customer.ID)
	log.Printf("Amount: %d %s", invoice.AmountPaid, invoice.Currency)

	// Here you can add your business logic for successful subscription payments
	// This is for recurring payments/subscriptions
}

func handleInvoicePaymentFailed(event stripe.Event) {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		log.Printf("Error parsing invoice payment failed event: %v", err)
		return
	}

	log.Printf("Invoice payment failed: %s", invoice.ID)
	log.Printf("Customer: %s", invoice.Customer.ID)
	log.Printf("Amount: %d %s", invoice.AmountDue, invoice.Currency)

	// Here you can add your business logic for failed subscription payments
	// For example:
	// - Send payment failure notification
	// - Update subscription status
	// - Implement dunning management
}

// createCustomerPortalSession creates a Stripe Customer Portal session to allow customers
// to manage their subscriptions, payment methods, and billing information
func createCustomerPortalSession(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Parse request body to get customer ID
	var req struct {
		CustomerID string `json:"customerId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.CustomerID == "" {
		http.Error(w, "Missing customerId", http.StatusBadRequest)
		return
	}

	// Set return URL from environment variable or use default
	returnURL := os.Getenv("CUSTOMER_PORTAL_RETURN_URL")
	if returnURL == "" {
		returnURL = "http://localhost:3000/my-sessions" // Default return URL
	}

	// Create billing portal session
	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(req.CustomerID),
		ReturnURL: stripe.String(returnURL),
	}

	portalSession, err := billingportal.New(params)
	if err != nil {
		log.Printf("Error creating customer portal session: %v", err)
		http.Error(w, "Failed to create customer portal session", http.StatusInternalServerError)
		return
	}

	// Return the portal session URL as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"url": portalSession.URL})
}
