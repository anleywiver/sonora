import type { Config } from "tailwindcss";

// Design tokens final dari STEP 1-4 (Design System)
const config: Config = {
  content: ["./src/**/*.{js,ts,jsx,tsx,mdx}"],
  darkMode: "class",
  theme: {
    extend: {
      colors: {
        background: "#050816",
        card: "#0F172A",
        primary: "#1D4ED8",
        secondary: "#2563EB",
        accent: "#3B82F6",
        hover: "#60A5FA",
        "text-primary": "#FFFFFF",
        "text-secondary": "#94A3B8",
        // Status colors — used sparingly for state (healthy/degraded/down
        // badges), never decoratively. See docs/design-system.md.
        success: "#4ADE80",
        warning: "#FACC15",
        error: "#F87171",
        info: "#60A5FA",
      },
      borderColor: {
        DEFAULT: "rgba(255,255,255,.06)",
      },
      fontFamily: {
        sans: ["var(--font-inter)", "sans-serif"],
      },
      borderRadius: {
        card: "20px",
        control: "16px",
      },
    },
  },
  plugins: [],
};

export default config;
