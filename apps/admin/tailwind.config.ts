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
