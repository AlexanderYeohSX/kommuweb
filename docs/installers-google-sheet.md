# Authorized Installers — Google Form / Sheet sync

The [installers page](/installers/) loads installer data at runtime from a **Google Apps Script** web app that reads your Form response spreadsheet. New form submissions appear after a browser refresh—no Jekyll rebuild.

## Form columns

Your response sheet headers:

| Column | Used for |
|--------|----------|
| Timestamp | (ignored by API) |
| Installer Name | Popup / table if no shop name |
| Contact Number | Contact line |
| Email | Email line |
| Shop Name | Primary display name |
| State | Table “State” column |
| City | (included in API; address is primary location text) |
| Latitude | Map pin latitude |
| Longitude | Map pin longitude |
| Address | Full address in popup and table |
| Approved | Must be approved for the row to appear on the site |

## Approval gate

Only rows with **Approved** set are published. The script treats these as approved:

- Checkbox checked (`TRUE`)
- Text: `Yes`, `yes`, `Y`, `Approved`, `approved`, `1`, `true`

Empty or any other value means the installer is **hidden** until approved in the sheet.

## Prerequisites

- Google Form linked to a spreadsheet (responses tab is usually **Form Responses 1**).
- Approved installers must have valid **Latitude** and **Longitude** (decimal degrees, within Malaysia).
- A Google account with edit access to the spreadsheet.

## 1. Apps Script

1. Open the **response spreadsheet** (from Form → Responses → spreadsheet icon).
2. **Extensions → Apps Script**.
3. Replace the default `Code.gs` with the template in [`docs/scripts/installers-sheet-api.gs`](scripts/installers-sheet-api.gs).
4. Confirm `COLUMN_MAP` matches your headers (defaults match the table above).
5. Save the project.

## 2. Deploy web app

1. **Deploy → New deployment**.
2. Type: **Web app**.
3. Execute as: **Me**.
4. Who has access: **Anyone** (required for public website `fetch`).
5. Deploy and copy the **Web app URL** (ends with `/exec`).

## 3. Website config

In [`_config.yml`](../_config.yml), set:

```yaml
installers_api_url: "https://script.google.com/macros/s/YOUR_DEPLOYMENT_ID/exec"
```

Rebuild or restart Jekyll so the URL is baked into the page.

## API response shape

```json
{
  "installers": [
    {
      "id": "row-2",
      "name": "Example Auto Sdn Bhd",
      "installer_name": "Ali Ahmad",
      "shop_name": "Example Auto Sdn Bhd",
      "state": "Selangor",
      "city": "Petaling Jaya",
      "contact": "+60 12-345 6789",
      "email": "ali@example.com",
      "address": "123 Jalan Example, 47500 Petaling Jaya, Selangor",
      "lat": 3.0825,
      "lng": 101.5850
    }
  ]
}
```

Rows are skipped when not approved, name/state is empty, or lat/lng is missing or outside Malaysia.

## Local development without the API

- Leave `installers_api_url` empty in `_config.yml`, or open `/installers/?dev=1`.
- The page uses sample data from [`_data/installers_sample.yaml`](../_data/installers_sample.yaml).

## Troubleshooting

| Issue | Fix |
|-------|-----|
| Page shows “Could not load installers” | Check URL in `_config.yml`; redeploy Script; ensure access is **Anyone**. |
| New response not listed | Set **Approved** to Yes/checkbox; confirm lat/lng; refresh the page. |
| Installer still hidden | Check **Executions** in Apps Script for skip reasons. |
| Wrong columns mapped | Update `COLUMN_MAP` in the script to match sheet headers exactly. |
| Pin in wrong place | Verify **Latitude** / **Longitude** columns in the sheet. |

## Updating the Script

After code changes, use **Deploy → Manage deployments → Edit → New version** so the `/exec` URL keeps working.
