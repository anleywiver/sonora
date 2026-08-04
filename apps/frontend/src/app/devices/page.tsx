"use client";

import { ChevronDown, Laptop, Smartphone, Monitor, Check } from "lucide-react";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";

import { apiFetch } from "@/lib/api";
import { useAuthStore } from "@/store/auth";

interface Device {
  id: string;
  name: string;
  type: "web" | "mobile" | "desktop";
  is_active: boolean;
}

const iconFor = { web: Laptop, mobile: Smartphone, desktop: Monitor };

export default function DevicesPage() {
  const router = useRouter();
  const thisDeviceId = useAuthStore((s) => s.deviceId);
  const [devices, setDevices] = useState<Device[]>([]);
  const [transferring, setTransferring] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  const load = () => {
    apiFetch<Device[]>("/devices")
      .then(setDevices)
      .catch(() => setError("Gagal memuat daftar device."));
  };

  useEffect(load, []);

  const handleTransfer = async (deviceId: string) => {
    setTransferring(deviceId);
    setError(null);
    try {
      await apiFetch("/player/transfer", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ device_id: deviceId }),
      });
      load();
    } catch {
      setError("Gagal transfer playback.");
    } finally {
      setTransferring(null);
    }
  };

  return (
    <main className="min-h-screen px-6 pb-32 pt-6">
      <div className="flex items-center justify-between">
        <button onClick={() => router.back()} aria-label="Back">
          <ChevronDown size={28} />
        </button>
        <h1 className="text-lg font-bold">Devices</h1>
        <div className="w-7" aria-hidden />
      </div>

      <p className="mt-6 text-xs text-text-secondary">
        Pilih device untuk melanjutkan playback di sana (Transfer Playback).
      </p>

      {error && <p className="mt-4 text-sm text-error">{error}</p>}

      <ul className="mt-4 space-y-2">
        {devices.map((d) => {
          const Icon = iconFor[d.type] ?? Laptop;
          const isThisDevice = d.id === thisDeviceId;
          return (
            <li key={d.id}>
              <button
                onClick={() => handleTransfer(d.id)}
                disabled={transferring === d.id}
                className="flex w-full items-center gap-3 rounded-[18px] border border-border bg-white/5 p-3.5 text-left disabled:opacity-50"
              >
                <Icon size={20} className="text-text-secondary" />
                <div className="min-w-0 flex-1">
                  <p className="truncate text-sm font-medium">
                    {d.name} {isThisDevice && "(perangkat ini)"}
                  </p>
                </div>
                {d.is_active && <Check size={18} className="text-success" />}
              </button>
            </li>
          );
        })}
      </ul>

      {devices.length === 0 && !error && (
        <p className="mt-4 text-sm text-text-secondary">Tidak ada device lain.</p>
      )}
    </main>
  );
}
