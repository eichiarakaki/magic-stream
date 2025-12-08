import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import * as fs from "node:fs";

// https://vite.dev/config/
export default defineConfig({
  server: {
    https: {
      key: fs.readFileSync("../../localhost-key.pem"),
      cert: fs.readFileSync("../../localhost.pem"),
    },
    port: 5173,
  },
  plugins: [react()],
});
