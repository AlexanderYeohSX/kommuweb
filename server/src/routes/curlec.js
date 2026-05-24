const express = require('express');
const { rmToSen } = require('../lib/amount');
const { getClient, verifyPaymentSignature, trimNotes } = require('../lib/razorpay');
const { loadStore, persistOrder, persistSubscription } = require('../lib/store');

const router = express.Router();
let store = loadStore();

function uniqueReceipt(prefix) {
  return `${prefix}_${Date.now()}_${Math.random().toString(36).slice(2, 9)}`;
}

function publicKey() {
  return process.env.RAZORPAY_KEY_ID;
}

function callbackUrl(path) {
  const base = (process.env.CALLBACK_BASE_URL || process.env.STOREFRONT_ORIGIN || 'https://kommu.ai').replace(/\/$/, '');
  return `${base}${path.startsWith('/') ? path : `/${path}`}`;
}

/** POST /curlec/orders — create Razorpay order for one-off checkout */
router.post('/orders', async (req, res) => {
  try {
    const body = req.body || {};
    const currency = (body.currency || 'MYR').toUpperCase();
    const amountMajor = parseFloat(body.amount);
    if (Number.isNaN(amountMajor) || amountMajor <= 0) {
      return res.status(400).json({ error: 'Invalid amount' });
    }
    const amountSen = rmToSen(amountMajor);
    const receipt = body.receipt || uniqueReceipt('kommu');

    const notes = trimNotes({
      source: body.source || '',
      name: body.name || '',
      email: body.email || '',
      harness: body.harness || '',
      installation: body.installation || '',
      promoCode: body.promoCode || '',
      deviceprice: body.deviceprice != null ? String(body.deviceprice) : '',
      shippingrate: body.shippingrate != null ? String(body.shippingrate) : '',
      ic: body.ic || ''
    });

    const order = await getClient().orders.create({
      amount: amountSen,
      currency,
      receipt,
      notes
    });

    persistOrder(store, receipt, {
      razorpay_order_id: order.id,
      amountSen,
      currency,
      payload: body
    });

    res.json({
      key_id: publicKey(),
      order_id: order.id,
      amount: order.amount,
      currency: order.currency,
      receipt
    });
  } catch (err) {
    console.error('POST /curlec/orders', err);
    res.status(500).json({ error: err.message || 'Order creation failed' });
  }
});

/** POST /curlec/subscriptions — rent-to-own with upfront addon (deposit) */
router.post('/subscriptions', async (req, res) => {
  try {
    const body = req.body || {};
    const plan_id = body.plan_id || body.subscription;
    const total_count = parseInt(body.total_count || body.duration, 10);
    if (!plan_id) return res.status(400).json({ error: 'plan_id is required' });
    if (!total_count || total_count < 1) {
      return res.status(400).json({ error: 'total_count must be a positive integer' });
    }

    const depositMajor = parseFloat(body.deposit || 0);
    const currency = (body.currency || 'MYR').toUpperCase();
    const addons = [];
    if (depositMajor > 0) {
      addons.push({
        item: {
          name: body.addon_name || 'Deposit',
          amount: rmToSen(depositMajor),
          currency
        }
      });
    }

    const notes = trimNotes({
      harness: body.harness || '',
      promoCode: body.promoCode || '',
      sub_device_price: body.sub_device_price != null ? String(body.sub_device_price) : '',
      sim: body.sim || '',
      tradeIn: body.tradeIn || '',
      installation: body.installation || '',
      name: body.name || '',
      email: body.email || ''
    });

    const subPayload = {
      plan_id,
      total_count,
      quantity: parseInt(body.quantity, 10) || 1,
      customer_notify: body.customer_notify !== false ? 1 : 0,
      notes
    };
    if (addons.length) subPayload.addons = addons;

    const subscription = await getClient().subscriptions.create(subPayload);

    persistSubscription(store, subscription.id, {
      plan_id,
      total_count,
      payload: body
    });

    res.json({
      key_id: publicKey(),
      subscription_id: subscription.id,
      short_url: subscription.short_url || null
    });
  } catch (err) {
    console.error('POST /curlec/subscriptions', err);
    res.status(500).json({ error: err.message || 'Subscription creation failed' });
  }
});

/** POST /curlec/verify — payment signature verification */
router.post('/verify', (req, res) => {
  const {
    razorpay_order_id,
    razorpay_payment_id,
    razorpay_signature
  } = req.body || {};

  if (!razorpay_order_id || !razorpay_payment_id || !razorpay_signature) {
    return res.status(400).json({ error: 'Missing payment verification fields' });
  }

  const valid = verifyPaymentSignature(
    razorpay_order_id,
    razorpay_payment_id,
    razorpay_signature
  );

  res.json({ valid });
});

/** POST /curlec/callback — FPX / redirect flow: verify and redirect to storefront */
router.post('/callback', express.urlencoded({ extended: true }), (req, res) => {
  const {
    razorpay_order_id,
    razorpay_payment_id,
    razorpay_signature,
    razorpay_subscription_id
  } = req.body || {};

  const successUrl = callbackUrl('/trx-success/');
  const failureUrl = callbackUrl('/trx-failure/');

  if (razorpay_subscription_id) {
    return res.redirect(302, `${successUrl}?subscription_id=${encodeURIComponent(razorpay_subscription_id)}`);
  }

  if (
    razorpay_order_id &&
    razorpay_payment_id &&
    razorpay_signature &&
    verifyPaymentSignature(razorpay_order_id, razorpay_payment_id, razorpay_signature)
  ) {
    const q = new URLSearchParams({
      razorpay_order_id,
      razorpay_payment_id,
      razorpay_signature
    });
    return res.redirect(302, `${successUrl}?${q.toString()}`);
  }

  res.redirect(302, failureUrl);
});

/** Legacy GET redirects — keep query-param flows working during migration */
router.get('/otp', (req, res) => {
  res.status(410).json({
    error: 'Deprecated',
    message: 'Use POST /curlec/orders and Razorpay Standard Checkout from kommuweb checkout.'
  });
});

router.get('/visa', (req, res) => {
  res.status(410).json({
    error: 'Deprecated',
    message: 'Use POST /curlec/subscriptions and Razorpay Standard Checkout from kommuweb checkout.'
  });
});

module.exports = router;
