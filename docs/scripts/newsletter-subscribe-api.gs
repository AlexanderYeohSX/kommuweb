/**
 * Newsletter subscribe + unsubscribe API — paste into Apps Script on KA Inventory spreadsheet.
 * Deploy: Deploy → New deployment → Web app → Execute as Me → Who has access: Anyone
 *
 * POST (JSON body) → subscribe
 * GET ?action=unsubscribe&email=...&token=... → set status inactive
 *
 * Script property (Project settings → Script properties):
 *   UNSUBSCRIBE_SECRET = same value as Athena .env UNSUBSCRIBE_SECRET
 *
 * Use the /exec URL in kommuweb _config.yml → newsletter_api_url
 * Use the same /exec URL in Athena .env → NEWSLETTER_APPS_SCRIPT_URL (for drip unsubscribe links)
 */
var SPREADSHEET_ID = '11eE_xlzMILBkW9W1te96Q3x3TLa0Wihpk1Ypy0gU1jk';
var NEWSLETTER_TAB = 'Newsletter';
var HEADERS = [
  'email',
  'name',
  'source',
  'subscribed_at',
  'sequence_step',
  'last_sent_at',
  'status',
  'next_send_at'
];
var ALLOWED_SOURCES = { homepage: true, checkout: true, import: true };
var EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

function doPost(e) {
  try {
    var body = {};
    if (e && e.postData && e.postData.contents) {
      body = JSON.parse(e.postData.contents);
    }
    var result = upsertSubscriber(body);
    return jsonResponse(result);
  } catch (err) {
    return jsonResponse({ ok: false, error: String(err.message || err) }, 400);
  }
}

function doGet(e) {
  var params = (e && e.parameter) || {};
  if (params.action === 'unsubscribe') {
    return handleUnsubscribe(params);
  }
  return jsonResponse({ ok: true, service: 'kommu-newsletter' });
}

function handleUnsubscribe(params) {
  var email = String(params.email || '').trim().toLowerCase();
  var token = String(params.token || '').trim();

  if (!EMAIL_RE.test(email)) {
    return htmlResponse(unsubscribePage('Missing or invalid email address.'));
  }

  if (!verifyUnsubscribeToken(email, token)) {
    return htmlResponse(
      unsubscribePage('This unsubscribe link is invalid. Email support@kommu.ai for help.')
    );
  }

  var result = unsubscribeSubscriber(email);
  if (!result.ok) {
    return htmlResponse(
      unsubscribePage('We could not find that subscription. You may already be unsubscribed.')
    );
  }

  return htmlResponse(
    unsubscribePage(email + ' will no longer receive Kommu newsletter emails.', true)
  );
}

function unsubscribeSubscriber(email) {
  var sheet = getNewsletterSheet();
  var values = sheet.getDataRange().getValues();

  for (var r = 1; r < values.length; r++) {
    var rowEmail = String(values[r][0] || '').trim().toLowerCase();
    if (rowEmail !== email) continue;
    sheet.getRange(r + 1, 7).setValue('inactive');
    return { ok: true, email: email, status: 'inactive' };
  }

  return { ok: false, email: email };
}

function unsubscribeToken(email) {
  var secret = PropertiesService.getScriptProperties().getProperty('UNSUBSCRIBE_SECRET');
  if (!secret) return '';
  var sig = Utilities.computeHmacSha256Signature(email, secret);
  return bytesToHex(sig);
}

function verifyUnsubscribeToken(email, token) {
  var secret = PropertiesService.getScriptProperties().getProperty('UNSUBSCRIBE_SECRET');
  if (!secret || !token) return false;
  var expected = unsubscribeToken(email);
  return expected === token;
}

function bytesToHex(bytes) {
  return bytes
    .map(function (b) {
      var v = (b < 0 ? b + 256 : b).toString(16);
      return v.length === 1 ? '0' + v : v;
    })
    .join('');
}

function upsertSubscriber(body) {
  var email = String(body.email || '').trim().toLowerCase();
  var name = String(body.name || '').trim();
  var source = String(body.source || 'homepage').trim() || 'homepage';

  if (!EMAIL_RE.test(email)) {
    throw new Error('Invalid email address');
  }
  if (!ALLOWED_SOURCES[source]) {
    throw new Error('Invalid source');
  }

  var sheet = getNewsletterSheet();
  var values = sheet.getDataRange().getValues();
  if (values.length === 0) {
    sheet.getRange(1, 1, 1, HEADERS.length).setValues([HEADERS]);
    values = [HEADERS];
  }

  for (var r = 1; r < values.length; r++) {
    var rowEmail = String(values[r][0] || '').trim().toLowerCase();
    if (rowEmail !== email) continue;
    if (name) {
      sheet.getRange(r + 1, 2).setValue(name);
    }
    return {
      ok: true,
      email: email,
      created: false,
      status: String(values[r][6] || 'active').trim() || 'active'
    };
  }

  sheet.appendRow([email, name, source, mytNow(), '0', '', 'active', '']);

  return { ok: true, email: email, created: true, status: 'active' };
}

function getNewsletterSheet() {
  var ss = SpreadsheetApp.openById(SPREADSHEET_ID);
  var sheet = ss.getSheetByName(NEWSLETTER_TAB);
  if (!sheet) {
    throw new Error('Sheet not found: ' + NEWSLETTER_TAB);
  }
  return sheet;
}

function mytNow() {
  return Utilities.formatDate(new Date(), 'Asia/Kuala_Lumpur', 'dd/MM/yyyy HH:mm:ss');
}

function unsubscribePage(message, success) {
  var title = success ? "You're unsubscribed" : 'Unsubscribe';
  return (
    '<!DOCTYPE html><html lang="en"><head><meta charset="utf-8">' +
    '<meta name="viewport" content="width=device-width,initial-scale=1">' +
    '<title>' +
    title +
    ' — Kommu</title><style>' +
    'body{margin:0;min-height:100vh;display:grid;place-items:center;background:#000;' +
    'color:rgb(241,241,241);font-family:-apple-system,BlinkMacSystemFont,Segoe UI,Roboto,Helvetica,Arial,sans-serif}' +
    'main{max-width:28rem;padding:2rem;text-align:center;line-height:1.6}' +
    'h1{font-size:1.25rem;font-weight:600;margin:0 0 .75rem}' +
    'p{margin:0;color:#888}a{color:rgb(241,241,241)}</style></head><body><main>' +
    '<h1>' +
    title +
    '</h1><p>' +
    message +
    '</p><p style="margin-top:1rem"><a href="https://kommu.ai">Back to kommu.ai</a></p>' +
    '</main></body></html>'
  );
}

function jsonResponse(obj) {
  var output = ContentService.createTextOutput(JSON.stringify(obj));
  output.setMimeType(ContentService.MimeType.JSON);
  return output;
}

function htmlResponse(html) {
  return HtmlService.createHtmlOutput(html).setXFrameOptionsMode(
    HtmlService.XFrameOptionsMode.ALLOWALL
  );
}
