import { defineConfig } from "@playwright/test";

export default defineConfig({
  testDir: ".",
  testMatch: "web-demo.ts",
  timeout: 120_000,
  use: {
    baseURL: "http://localhost:3000",
    viewport: { width: 1800, height: 1000 },
    video: {
      mode: "on",
      size: { width: 1800, height: 1000 },
    },
    launchOptions: {
      slowMo: 50,
    },
  },
  reporter: "list",
  outputDir: "web-demo-results",
});
