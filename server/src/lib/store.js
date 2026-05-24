const fs = require('fs');
const path = require('path');

const DEFAULT_PATH = path.join(__dirname, '../../data/orders.json');

function loadStore(filePath) {
  const p = filePath || process.env.ORDERS_STORE_PATH || DEFAULT_PATH;
  try {
    if (fs.existsSync(p)) {
      return { path: p, data: JSON.parse(fs.readFileSync(p, 'utf8')) };
    }
  } catch (_) {}
  return { path: p, data: { orders: {}, subscriptions: {} } };
}

function saveStore(store) {
  const dir = path.dirname(store.path);
  if (!fs.existsSync(dir)) fs.mkdirSync(dir, { recursive: true });
  fs.writeFileSync(store.path, JSON.stringify(store.data, null, 2));
}

function persistOrder(store, receipt, payload) {
  store.data.orders[receipt] = {
    ...payload,
    createdAt: new Date().toISOString()
  };
  saveStore(store);
}

function persistSubscription(store, subscriptionId, payload) {
  store.data.subscriptions[subscriptionId] = {
    ...payload,
    createdAt: new Date().toISOString()
  };
  saveStore(store);
}

module.exports = {
  loadStore,
  persistOrder,
  persistSubscription
};
