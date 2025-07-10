# Stripe Webhook Setup

This document explains how to set up Stripe webhooks for your application.

## Setting up Webhooks

### 1. Create a Webhook Endpoint in Stripe Dashboard

1. Log in to your [Stripe Dashboard](https://dashboard.stripe.com/)
2. Go to **Developers** → **Webhooks**
3. Click **Add endpoint**
4. Enter your endpoint URL: `https://your-domain.com/webhook`
   - For local testing: `http://localhost:4242/webhook`
   - For production: `https://your-production-domain.com/webhook`

### 2. Select Events to Listen For

Select the following events that your application handles:

**For Checkout Sessions:**
- `checkout.session.completed` - When a checkout session is successfully completed
- `checkout.session.expired` - When a checkout session expires

**For One-time Payments:**
- `payment_intent.succeeded` - When a payment is successful
- `payment_intent.payment_failed` - When a payment fails

**For Subscriptions:**
- `invoice.payment_succeeded` - When a subscription payment is successful
- `invoice.payment_failed` - When a subscription payment fails

### 3. Get Your Webhook Secret

1. After creating the webhook, click on it in the Stripe Dashboard
2. Copy the **Signing secret** (starts with `whsec_`)
3. Add it to your `.env` file as `STRIPE_WEBHOOK_SECRET`

## Environment Variables

Make sure to set these environment variables:

```bash
# Required
STRIPE_SECRET_KEY=sk_test_your_secret_key_here
STRIPE_WEBHOOK_SECRET=whsec_your_webhook_secret_here

# Optional (defaults provided)
ADDR=:4242
SUCCESS_URL=http://localhost:3000/success
CANCEL_URL=http://localhost:3000/cancel
```

## Testing Webhooks Locally

### Using Stripe CLI

1. Install the [Stripe CLI](https://stripe.com/docs/stripe-cli)
2. Log in to your Stripe account:
   ```bash
   stripe login
   ```
3. Forward webhooks to your local server:
   ```bash
   stripe listen --forward-to localhost:4242/webhook
   ```
4. The CLI will display a webhook signing secret - use this as your `STRIPE_WEBHOOK_SECRET`

### Using ngrok (Alternative)

1. Install [ngrok](https://ngrok.com/)
2. Expose your local server:
   ```bash
   ngrok http 4242
   ```
3. Use the ngrok URL in your Stripe webhook configuration

## Webhook Events Handled

The application handles the following webhook events:

### `checkout.session.completed`
- Triggered when a customer completes a checkout session
- Use this to:
  - Update user's subscription status
  - Send confirmation emails
  - Grant access to paid content
  - Update database records

### `checkout.session.expired`
- Triggered when a checkout session expires
- Use this to:
  - Clean up temporary data
  - Send abandonment emails
  - Update analytics

### `payment_intent.succeeded`
- Triggered when a one-time payment is successful
- Use this for processing successful one-time payments

### `payment_intent.payment_failed`
- Triggered when a payment fails
- Use this to:
  - Send failure notifications
  - Update payment status
  - Implement retry logic

### `invoice.payment_succeeded`
- Triggered when a subscription payment is successful
- Use this for processing recurring payments

### `invoice.payment_failed`
- Triggered when a subscription payment fails
- Use this to:
  - Send payment failure notifications
  - Update subscription status
  - Implement dunning management

## Security Notes

1. **Always verify webhook signatures** - The application checks the `Stripe-Signature` header
2. **Use HTTPS in production** - Stripe requires HTTPS for webhook endpoints
3. **Keep your webhook secret secure** - Never expose it in client-side code
4. **Handle idempotency** - Stripe may send the same webhook multiple times

## Troubleshooting

### Common Issues

1. **Webhook signature verification failed**
   - Check that `STRIPE_WEBHOOK_SECRET` is correctly set
   - Ensure you're using the correct signing secret from Stripe Dashboard

2. **Webhook not receiving events**
   - Verify the endpoint URL is correct
   - Check that your server is accessible from the internet
   - Ensure the webhook is enabled in Stripe Dashboard

3. **Local testing issues**
   - Use Stripe CLI or ngrok for local testing
   - Check firewall settings
   - Verify port forwarding is working

### Debugging

Enable detailed logging by checking the server logs. The application logs:
- Webhook events received
- Processing status
- Any errors encountered

## Example Response

Your webhook endpoint should return a `200 OK` status code to acknowledge successful receipt of the webhook.

The application automatically handles this for all supported events.
