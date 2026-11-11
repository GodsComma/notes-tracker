import { defineConfig, devices } from '@playwright/test'

const PORT = 5199

export default defineConfig({
  testDir: './e2e',
  timeout: 90_000,
  expect: { timeout: 5_000 },
  fullyParallel: true,
  reporter: [['list']],
  use: {
    baseURL: `http://localhost:${PORT}`,
    viewport: { width: 1400, height: 1000 },
    trace: 'retain-on-failure',
  },
  projects: [
    { name: 'chromium-dpr1', use: { ...devices['Desktop Chrome'], viewport: { width: 1400, height: 1000 }, deviceScaleFactor: 1 } },
    { name: 'chromium-dpr2', use: { ...devices['Desktop Chrome'], viewport: { width: 1400, height: 1000 }, deviceScaleFactor: 2 } },
    { name: 'firefox-dpr1',  use: { ...devices['Desktop Firefox'], viewport: { width: 1400, height: 1000 }, deviceScaleFactor: 1 } },
    // WebKit is opt-in (PW_WEBKIT=1): on macOS 14 Playwright ships a frozen
    // WebKit build that fails to open pages with current Playwright
    // ("Unknown setting: PushAPIEnabled"). Works on macOS 15+ and Linux CI.
    ...(process.env.PW_WEBKIT
      ? [{ name: 'webkit-dpr2', use: { ...devices['Desktop Safari'], viewport: { width: 1400, height: 1000 }, deviceScaleFactor: 2 } }]
      : []),
  ],
  webServer: {
    command: `npx vite --port ${PORT} --strictPort`,
    url: `http://localhost:${PORT}/__e2e/pdf-viewer`,
    reuseExistingServer: !process.env.CI,
    timeout: 60_000,
  },
})
