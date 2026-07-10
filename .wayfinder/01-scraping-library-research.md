# Ticket 01: Scraping Library Research

**Type**: Research (AFK)
**Blocked by**: None (frontier)
**Status**: RESOLVED

## Question

Which Go scraping library should we use for BCV's website?

## Resolution

**Recommendation: `net/http` + `goquery`** — no need for chromedp or colly.

### Analysis of bcv.org.ve

The BCV website is a **Drupal 7** site that renders all exchange rate data as **static server-side HTML**. No JavaScript rendering is required.

**Data found on the homepage and `/estadisticas/tipo-cambio-de-referencia-smc`:**

#### 1. Tipo de Cambio de Referencia SMC (weighted average)
Single rate per currency (no buy/sell split):
```html
<div id="dolar" class="col-sm-12 col-xs-12">
  <span> USD</span>
  <strong class="strong-tb">709,69350000</strong>
</div>
<div id="euro" class="col-sm-12 col-xs-12">
  <span> EUR</span>
  <strong class="strong-tb">811,44935403</strong>
</div>
<!-- Date: -->
<span class="date-display-single" content="2026-07-10T00:00:00-04:00">
  Viernes, 10 Julio 2026
</span>
```

#### 2. Tasas Informativas del Sistema Bancario (buy/sell per bank)
HTML table with bank-specific rates:
```html
<table class="views-table cols-3 table">
  <thead>
    <tr><th>Banco</th><th>Compra</th><th>Venta</th></tr>
  </thead>
  <tbody>
    <tr><td>Banesco</td><td>700,8342</td><td>697,8914</td></tr>
    <tr><td>Banco Mercantil</td><td>700,2248</td><td>700,2287</td></tr>
    <!-- ... -->
  </tbody>
</table>
```

### Why goquery over colly/chromedp

| Concern | goquery | colly | chromedp |
|---------|---------|-------|----------|
| JS rendering | N/A | Limited | Full Chrome |
| Dependencies | Minimal | Medium | Heavy (Chrome binary) |
| Speed | Fast | Fast | Slow |
| Learning curve | Low | Low | Medium |
| Fit for this use case | Perfect | Overkill | Overkill |

The BCV data is static HTML. `net/http` fetches the page, `goquery` parses it with jQuery-like selectors. No browser automation needed.

### CSS Selectors for scraping

```
#dolar .strong-tb          → USD reference rate
#euro .strong-tb           → EUR reference rate
.date-display-single[content] → ISO date
.views-table tbody tr      → Bank rates table rows
td.views-field-field-tasa-compra → Buy rate
td.views-field-field-tasa-venta  → Sell rate
```
