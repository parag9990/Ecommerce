import type { Config } from "tailwindcss";

export default {
  content: ["./index.html", "./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      boxShadow: {
        panel: "0 1px 2px rgb(24 24 27 / 0.06)"
      }
    }
  },
  plugins: []
} satisfies Config;
