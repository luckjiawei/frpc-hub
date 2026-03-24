import { useState, useRef, useCallback } from "react";
import { apiGet } from "../../lib/api";
import pb from "../../lib/pocketbase";

export interface Device {
  ip: string;
  mac: string;
  hostname: string;
  interface: string;
}

export interface LocalNetwork {
  interface: string;
  cidr: string;
  hosts: number;
}

export interface PortResult {
  port: number;
  open: boolean;
  service: string;
  done?: boolean;
}

export type ScanStep = "discover" | "scan";
export type ScanMode = "common" | "custom" | "full";

const COMMON_PORT_COUNT = 23; // mirrors wellKnownPorts in service.go
const FULL_PORT_COUNT = 65535;

export function useScanner() {
  const [step, setStep] = useState<ScanStep>("discover");
  const [discovering, setDiscovering] = useState(false);
  const [devices, setDevices] = useState<Device[]>([]);
  const [localNetworks, setLocalNetworks] = useState<LocalNetwork[]>([]);
  const [selectedDevice, setSelectedDevice] = useState<Device | null>(null);
  const [scanning, setScanning] = useState(false);
  const [portResults, setPortResults] = useState<PortResult[]>([]);
  const [scanDone, setScanDone] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [scanMode, setScanMode] = useState<ScanMode>("common");
  const [customPorts, setCustomPorts] = useState("");
  const [totalPorts, setTotalPorts] = useState(0);
  const esRef = useRef<EventSource | null>(null);

  const discoverDevices = useCallback(async () => {
    setDiscovering(true);
    setDevices([]);
    setError(null);
    try {
      const [netRes, discoverRes] = await Promise.all([
        apiGet("/api/scanner/interfaces"),
        apiGet("/api/scanner/discover"),
      ]);
      if (netRes.ok) {
        const nets: LocalNetwork[] = await netRes.json();
        setLocalNetworks(nets ?? []);
      }
      if (!discoverRes.ok) {
        const body = await discoverRes.json().catch(() => ({}));
        throw new Error(body.error || `HTTP ${discoverRes.status}`);
      }
      const data: Device[] = await discoverRes.json();
      setDevices(data ?? []);
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : String(e));
    } finally {
      setDiscovering(false);
    }
  }, []);

  const selectDevice = useCallback((device: Device) => {
    setSelectedDevice(device);
    setPortResults([]);
    setScanDone(false);
    setStep("scan");
  }, []);

  const startPortScan = useCallback(() => {
    if (!selectedDevice) return;
    esRef.current?.close();

    setPortResults([]);
    setScanDone(false);
    setScanning(true);
    setError(null);

    const token = pb.authStore.token;
    const params = new URLSearchParams({ ip: selectedDevice.ip, token });

    let total = COMMON_PORT_COUNT;
    if (scanMode === "full") {
      params.set("full", "true");
      total = FULL_PORT_COUNT;
    } else if (scanMode === "custom" && customPorts.trim()) {
      params.set("ports", customPorts.trim());
      total = customPorts.split(",").filter((p) => p.trim()).length;
    }
    setTotalPorts(total);

    const es = new EventSource(`/api/scanner/ports/stream?${params.toString()}`);
    esRef.current = es;

    es.onmessage = (ev) => {
      try {
        const result: PortResult = JSON.parse(ev.data);
        if (result.done) {
          setScanDone(true);
          setScanning(false);
          es.close();
          return;
        }
        setPortResults((prev) => [...prev, result]);
      } catch {
        /* ignore */
      }
    };

    es.onerror = () => {
      setScanning(false);
      setScanDone(true);
      es.close();
    };
  }, [selectedDevice, scanMode, customPorts]);

  const stopScan = useCallback(() => {
    esRef.current?.close();
    setScanning(false);
    setScanDone(true);
  }, []);

  const backToDiscover = useCallback(() => {
    stopScan();
    setStep("discover");
    setSelectedDevice(null);
    setPortResults([]);
    setScanDone(false);
  }, [stopScan]);

  return {
    step,
    discovering,
    devices,
    localNetworks,
    selectedDevice,
    scanning,
    portResults,
    scanDone,
    error,
    scanMode,
    setScanMode,
    customPorts,
    setCustomPorts,
    totalPorts,
    discoverDevices,
    selectDevice,
    startPortScan,
    stopScan,
    backToDiscover,
  };
}
