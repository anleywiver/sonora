"use client";

import { useEffect } from "react";

export function RegisterServiceWorker() {
  useEffect(() => {
    if ("serviceWorker" in navigator) {
      navigator.serviceWorker.register("/sw.js").catch(() => {
        // No PWA install/offline support this session — the app still
        // works normally online.
      });
    }
  }, []);

  return null;
}
