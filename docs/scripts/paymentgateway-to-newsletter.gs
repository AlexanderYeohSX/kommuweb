/**
 * Sync emails from PaymentGateway → Newsletter (skip if already subscribed).
 *
 * Add as a NEW file in your existing KA Inventory Apps Script project
 * (the one that already has newsletter subscribe + onEdit). Do not paste
 * into Code.gs — File → + → Script, name it e.g. PaymentGatewayNewsletter.
 *
 * Reuses SPREADSHEET_ID, NEWSLETTER_TAB, EMAIL_RE, HEADERS from
 * newsletter-subscribe-api (already in Code.gs).
 *
 * Setup (once):
 *   1. Paste this file as a new script in the same project
 *   2. Run → installPaymentGatewayNewsletterTrigger
 *   3. Approve permissions when prompted
 *
 * Manual run: syncPaymentGatewayToNewsletter
 * Auto: time-driven trigger every 10 minutes
 *
 * New Newsletter rows: source=checkout, sequence_step=0, status=active
 * Existing emails: left alone (optional name fill if Newsletter name is blank)
 */
var PAYMENT_GATEWAY_TAB = 'PaymentGateway';
var PAYMENT_GATEWAY_SOURCE = 'checkout';
var EMAIL_HEADER_ALIASES = {
  email: true,
  'customer email': true,
  customer_email: true,
  'e-mail': true,
  'buyer email': true,
  'payer email': true
};
var NAME_HEADER_ALIASES = {
  name: true,
  'customer name': true,
  customer_name: true,
  'buyer name': true,
  'full name': true,
  'payer name': true
};

/**
 * Install (or refresh) a 10-minute sync trigger.
 * Deletes prior triggers for syncPaymentGatewayToNewsletter first.
 */
function installPaymentGatewayNewsletterTrigger() {
  var handlers = ScriptApp.getProjectTriggers();
  for (var i = 0; i < handlers.length; i++) {
    if (handlers[i].getHandlerFunction() === 'syncPaymentGatewayToNewsletter') {
      ScriptApp.deleteTrigger(handlers[i]);
    }
  }
  ScriptApp.newTrigger('syncPaymentGatewayToNewsletter')
    .timeBased()
    .everyMinutes(10)
    .create();
  Logger.log('Installed: syncPaymentGatewayToNewsletter every 10 minutes');
}

/**
 * Main sync — safe to run manually or from the time trigger.
 */
function syncPaymentGatewayToNewsletter() {
  var lock = LockService.getScriptLock();
  if (!lock.tryLock(15000)) {
    Logger.log('Skipped: another sync is already running');
    return { ok: false, error: 'locked' };
  }

  try {
    var ss = SpreadsheetApp.openById(SPREADSHEET_ID);
    var paymentSheet = ss.getSheetByName(PAYMENT_GATEWAY_TAB);
    if (!paymentSheet) {
      throw new Error('Sheet not found: ' + PAYMENT_GATEWAY_TAB);
    }
    var newsletterSheet = getOrCreateNewsletterSheetPg_(ss);

    var paymentValues = paymentSheet.getDataRange().getValues();
    if (paymentValues.length < 2) {
      Logger.log('PaymentGateway has no data rows');
      return { ok: true, added: 0, skipped: 0, invalid: 0 };
    }

    var headers = paymentValues[0].map(function (h) {
      return String(h || '')
        .trim()
        .toLowerCase();
    });
    var emailCol = findColumnIndexPg_(headers, EMAIL_HEADER_ALIASES);
    var nameCol = findColumnIndexPg_(headers, NAME_HEADER_ALIASES);
    if (emailCol < 0) {
      throw new Error(
        'No email column in PaymentGateway. Headers: ' + paymentValues[0].join(', ')
      );
    }

    var existing = loadNewsletterEmailIndexPg_(newsletterSheet);
    var added = 0;
    var skipped = 0;
    var invalid = 0;
    var seenInBatch = {};

    for (var r = 1; r < paymentValues.length; r++) {
      var row = paymentValues[r];
      if (emailCol >= row.length) continue;

      var email = String(row[emailCol] || '')
        .trim()
        .toLowerCase();
      if (!email) continue;
      if (!EMAIL_RE.test(email)) {
        invalid++;
        continue;
      }
      if (seenInBatch[email]) {
        skipped++;
        continue;
      }
      seenInBatch[email] = true;

      var name =
        nameCol >= 0 && nameCol < row.length ? String(row[nameCol] || '').trim() : '';

      if (existing[email]) {
        // Already in Newsletter — optionally fill blank name
        if (name && !existing[email].name) {
          newsletterSheet.getRange(existing[email].row, 2).setValue(name);
          existing[email].name = name;
        }
        skipped++;
        continue;
      }

      newsletterSheet.appendRow([
        email,
        name,
        PAYMENT_GATEWAY_SOURCE,
        mytNow(),
        '0',
        '',
        'active',
        ''
      ]);
      existing[email] = { row: newsletterSheet.getLastRow(), name: name };
      added++;
    }

    var summary = {
      ok: true,
      added: added,
      skipped: skipped,
      invalid: invalid
    };
    Logger.log(
      'PaymentGateway → Newsletter: added=' +
        added +
        ' skipped=' +
        skipped +
        ' invalid=' +
        invalid
    );
    return summary;
  } finally {
    lock.releaseLock();
  }
}

function getOrCreateNewsletterSheetPg_(ss) {
  var sheet = ss.getSheetByName(NEWSLETTER_TAB);
  if (!sheet) {
    sheet = ss.insertSheet(NEWSLETTER_TAB);
    sheet.getRange(1, 1, 1, HEADERS.length).setValues([HEADERS]);
    return sheet;
  }
  var values = sheet.getDataRange().getValues();
  if (values.length === 0) {
    sheet.getRange(1, 1, 1, HEADERS.length).setValues([HEADERS]);
  }
  return sheet;
}

function loadNewsletterEmailIndexPg_(sheet) {
  var values = sheet.getDataRange().getValues();
  var index = {};
  for (var r = 1; r < values.length; r++) {
    var email = String(values[r][0] || '')
      .trim()
      .toLowerCase();
    if (!email) continue;
    index[email] = {
      row: r + 1,
      name: String(values[r][1] || '').trim()
    };
  }
  return index;
}

function findColumnIndexPg_(headers, aliases) {
  for (var i = 0; i < headers.length; i++) {
    if (aliases[headers[i]]) return i;
  }
  return -1;
}
