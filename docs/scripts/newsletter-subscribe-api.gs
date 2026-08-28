/**
 * Newsletter subscribe API — paste into Apps Script on KA Inventory spreadsheet.
 * Deploy: Deploy → New deployment → Web app → Execute as Me → Who has access: Anyone
 *
 * Writes to tab "Newsletter" (headers must match docs/newsletter-setup.md).
 * Use the /exec URL in kommuweb _config.yml → newsletter_api_url
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
  return jsonResponse({ ok: true, service: 'kommu-newsletter-subscribe' });
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

  sheet.appendRow([
    email,
    name,
    source,
    mytNow(),
    '0',
    '',
    'active',
    ''
  ]);

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

function jsonResponse(obj, status) {
  var output = ContentService.createTextOutput(JSON.stringify(obj));
  output.setMimeType(ContentService.MimeType.JSON);
  return output;
}
