/**
 * Authorized Installers API — paste into Apps Script bound to your Form response spreadsheet.
 * Deploy: Deploy → New deployment → Web app → Execute as Me → Who has access: Anyone
 *
 * Form headers (Form Responses 1):
 *   Timestamp | Installer Name | Contact Number | Email | Shop Name | State | City |
 *   Latitude | Longitude | Address | Approved
 *
 * Only rows with Approved = true are returned. Map pins use Latitude / Longitude.
 */
var SHEET_NAME = 'Form Responses 1';

var COLUMN_MAP = {
  installerName: 'Installer Name',
  contact: 'Contact Number',
  email: 'Email',
  shopName: 'Shop Name',
  state: 'State',
  city: 'City',
  lat: 'Latitude',
  lng: 'Longitude',
  address: 'Address',
  approved: 'Approved'
};

function doGet() {
  var sheet = SpreadsheetApp.getActiveSpreadsheet().getSheetByName(SHEET_NAME);
  if (!sheet) {
    return jsonResponse({ error: 'Sheet not found: ' + SHEET_NAME, installers: [] });
  }

  var values = sheet.getDataRange().getValues();
  if (values.length < 2) {
    return jsonResponse({ installers: [] });
  }

  var headers = values[0];
  var colIndex = buildColumnIndex(headers);
  var installers = [];

  for (var r = 1; r < values.length; r++) {
    var row = values[r];
    var record = rowToInstaller(row, colIndex, r + 1);
    if (record) installers.push(record);
  }

  installers.sort(function (a, b) {
    var s = String(a.state).localeCompare(String(b.state));
    if (s !== 0) return s;
    return String(a.name).localeCompare(String(b.name));
  });

  return jsonResponse({ installers: installers });
}

function buildColumnIndex(headers) {
  var index = {};
  var keys = Object.keys(COLUMN_MAP);
  for (var k = 0; k < keys.length; k++) {
    var field = keys[k];
    var wanted = COLUMN_MAP[field];
    var idx = headers.indexOf(wanted);
    if (idx === -1) {
      idx = findHeaderIndex(headers, wanted);
    }
    index[field] = idx;
  }
  return index;
}

function findHeaderIndex(headers, label) {
  var norm = normalizeHeader(label);
  for (var i = 0; i < headers.length; i++) {
    if (normalizeHeader(headers[i]) === norm) return i;
  }
  return -1;
}

function normalizeHeader(h) {
  return String(h || '').toLowerCase().replace(/\s+/g, ' ').trim();
}

function isApproved(value) {
  if (value === true) return true;
  if (value === false || value === null || value === undefined) return false;
  var s = String(value).toLowerCase().trim();
  if (!s) return false;
  if (s === 'true' || s === 'yes' || s === 'y' || s === '1' || s === 'approved') return true;
  return false;
}

function rowToInstaller(row, colIndex, rowNumber) {
  if (!isApproved(rawCell(row, colIndex.approved))) {
    return null;
  }

  var installerName = cell(row, colIndex.installerName);
  var shopName = cell(row, colIndex.shopName);
  var state = cell(row, colIndex.state);
  var city = cell(row, colIndex.city);
  var contact = cell(row, colIndex.contact);
  var email = cell(row, colIndex.email);
  var address = cell(row, colIndex.address);
  var lat = parseCoord(cell(row, colIndex.lat));
  var lng = parseCoord(cell(row, colIndex.lng));

  var displayName = shopName || installerName;
  if (!displayName || !state) {
    Logger.log('Skipping row ' + rowNumber + ': missing name or state');
    return null;
  }
  if (lat === null || lng === null) {
    Logger.log('Skipping row ' + rowNumber + ': missing lat/lng for ' + displayName);
    return null;
  }
  if (lat < 0.5 || lat > 8 || lng < 98 || lng > 120) {
    Logger.log('Skipping row ' + rowNumber + ': coords outside Malaysia for ' + displayName);
    return null;
  }

  return {
    id: 'row-' + rowNumber,
    name: displayName,
    installer_name: installerName || '',
    shop_name: shopName || '',
    state: state,
    city: city || '',
    contact: contact || '',
    email: email || '',
    address: address || '',
    lat: lat,
    lng: lng
  };
}

function rawCell(row, idx) {
  if (idx < 0 || idx >= row.length) return '';
  return row[idx];
}

function cell(row, idx) {
  var v = rawCell(row, idx);
  if (v === null || v === undefined) return '';
  return String(v).trim();
}

function parseCoord(v) {
  if (v === '' || v === null || v === undefined) return null;
  var n = parseFloat(String(v).replace(/,/g, ''));
  return isFinite(n) ? n : null;
}

function jsonResponse(obj) {
  return ContentService
    .createTextOutput(JSON.stringify(obj))
    .setMimeType(ContentService.MimeType.JSON);
}
