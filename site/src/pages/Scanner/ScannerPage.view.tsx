import {
  Badge,
  Box,
  Button,
  Card,
  Callout,
  Flex,
  Spinner,
  Table,
  Text,
  TextField,
  SegmentedControl,
  Progress,
} from "@radix-ui/themes";
import { useState, useEffect } from "react";
import { Icon } from "@iconify/react";
import { motion, AnimatePresence } from "framer-motion";
import { useTranslation } from "react-i18next";
import { PageHeader } from "../../components/PageHeader";
import type { Device, LocalNetwork, PortResult, ScanMode, ScanStep } from "./useScanner";

interface ScannerViewProps {
  step: ScanStep;
  discovering: boolean;
  devices: Device[];
  localNetworks: LocalNetwork[];
  selectedDevice: Device | null;
  scanning: boolean;
  portResults: PortResult[];
  scanDone: boolean;
  error: string | null;
  scanMode: ScanMode;
  setScanMode: (m: ScanMode) => void;
  customPorts: string;
  setCustomPorts: (v: string) => void;
  totalPorts: number;
  discoverDevices: () => void;
  selectDevice: (device: Device) => void;
  startPortScan: () => void;
  stopScan: () => void;
  backToDiscover: () => void;
}

export function ScannerView({
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
}: ScannerViewProps) {
  const { t } = useTranslation();
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    const timer = setTimeout(() => setMounted(true), 100);
    return () => clearTimeout(timer);
  }, []);

  const openPorts = portResults.filter((r) => r.open);
  const scannedCount = portResults.length;
  const progressPct = totalPorts > 0 ? Math.round((scannedCount / totalPorts) * 100) : 0;

  return (
    <Flex direction="column" gap="5" className="flex flex-1 flex-col">
      <PageHeader
        title={t("scanner.title")}
        description={t("scanner.description")}
        visible={mounted}
      />

      <AnimatePresence>
        {error && (
          <motion.div
            initial={{ opacity: 0, y: -8 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -8 }}
          >
            <Callout.Root color="red">
              <Callout.Icon>
                <Icon icon="lucide:alert-circle" width="16" height="16" />
              </Callout.Icon>
              <Callout.Text>{error}</Callout.Text>
            </Callout.Root>
          </motion.div>
        )}
      </AnimatePresence>

      {/* ── Step 1: Device Discovery ── */}
      {step === "discover" && (
        <motion.div
          key="discover"
          initial={{ opacity: 0, y: 16 }}
          animate={{ opacity: 1, y: 0 }}
          exit={{ opacity: 0, y: -16 }}
        >
          <Flex direction="column" gap="4">
            <Card>
              <Flex direction="column" gap="3" p="2">
                <Flex align="center" gap="2">
                  <Icon icon="lucide:wifi" width="20" height="20" style={{ color: "var(--accent-9)" }} />
                  <Text size="3" weight="bold">{t("scanner.step1Title")}</Text>
                </Flex>
                <Text size="2" color="gray">{t("scanner.step1Desc")}</Text>

                {localNetworks.length > 0 && (
                  <Flex gap="2" wrap="wrap">
                    {localNetworks.map((n) => (
                      <Badge key={n.cidr} variant="outline" color="blue">
                        <Icon icon="lucide:ethernet-port" width="12" height="12" />
                        {n.interface} · {n.cidr} · {n.hosts} {t("scanner.hosts")}
                      </Badge>
                    ))}
                  </Flex>
                )}

                <Button onClick={discoverDevices} disabled={discovering} style={{ alignSelf: "flex-start" }}>
                  {discovering ? (
                    <Flex align="center" gap="2">
                      <Spinner size="1" />
                      {t("scanner.discovering")}
                    </Flex>
                  ) : (
                    <Flex align="center" gap="2">
                      <Icon icon="lucide:search" width="16" height="16" />
                      {t("scanner.discoverBtn")}
                    </Flex>
                  )}
                </Button>
              </Flex>
            </Card>

            <AnimatePresence>
              {devices.length > 0 && (
                <motion.div
                  initial={{ opacity: 0, height: 0 }}
                  animate={{ opacity: 1, height: "auto" }}
                  exit={{ opacity: 0, height: 0 }}
                >
                  <Card>
                    <Flex direction="column" gap="3" p="2">
                      <Flex align="center" justify="between">
                        <Flex align="center" gap="2">
                          <Icon icon="lucide:monitor-check" width="18" height="18" style={{ color: "var(--green-9)" }} />
                          <Text size="3" weight="bold">
                            {t("scanner.devicesFound", { count: devices.length })}
                          </Text>
                        </Flex>
                        <Badge color="green" variant="soft">
                          {devices.length} {t("scanner.hosts")}
                        </Badge>
                      </Flex>

                      <Table.Root variant="surface">
                        <Table.Header>
                          <Table.Row>
                            <Table.ColumnHeaderCell>{t("scanner.ipAddress")}</Table.ColumnHeaderCell>
                            <Table.ColumnHeaderCell>{t("scanner.macAddress")}</Table.ColumnHeaderCell>
                            <Table.ColumnHeaderCell>{t("scanner.hostname")}</Table.ColumnHeaderCell>
                            <Table.ColumnHeaderCell>{t("scanner.iface")}</Table.ColumnHeaderCell>
                            <Table.ColumnHeaderCell>{t("common.actions")}</Table.ColumnHeaderCell>
                          </Table.Row>
                        </Table.Header>
                        <Table.Body>
                          {devices.map((device) => (
                            <motion.tr
                              key={device.ip}
                              initial={{ opacity: 0, x: -8 }}
                              animate={{ opacity: 1, x: 0 }}
                            >
                              <Table.Cell>
                                <Flex align="center" gap="2">
                                  <Icon icon="lucide:network" width="14" height="14" style={{ color: "var(--accent-9)" }} />
                                  <Text size="2" weight="medium" style={{ fontFamily: "monospace" }}>
                                    {device.ip}
                                  </Text>
                                </Flex>
                              </Table.Cell>
                              <Table.Cell>
                                <Text size="2" color="gray" style={{ fontFamily: "monospace" }}>
                                  {device.mac || "—"}
                                </Text>
                              </Table.Cell>
                              <Table.Cell>
                                <Text size="2" color="gray">{device.hostname || "—"}</Text>
                              </Table.Cell>
                              <Table.Cell>
                                <Badge variant="outline" size="1">{device.interface || "—"}</Badge>
                              </Table.Cell>
                              <Table.Cell>
                                <Button size="1" variant="soft" onClick={() => selectDevice(device)}>
                                  <Icon icon="lucide:scan-line" width="14" height="14" />
                                  {t("scanner.scanPorts")}
                                </Button>
                              </Table.Cell>
                            </motion.tr>
                          ))}
                        </Table.Body>
                      </Table.Root>
                    </Flex>
                  </Card>
                </motion.div>
              )}
            </AnimatePresence>
          </Flex>
        </motion.div>
      )}

      {/* ── Step 2: Port Scanner ── */}
      {step === "scan" && selectedDevice && (
        <motion.div
          key="scan"
          initial={{ opacity: 0, y: 16 }}
          animate={{ opacity: 1, y: 0 }}
          exit={{ opacity: 0, y: -16 }}
        >
          <Flex direction="column" gap="4">
            {/* Target + controls */}
            <Card>
              <Flex direction="column" gap="4" p="2">
                {/* Row 1: back / target / start-stop */}
                <Flex align="center" justify="between" wrap="wrap" gap="3">
                  <Flex align="center" gap="3">
                    <Button variant="ghost" size="2" onClick={backToDiscover}>
                      <Icon icon="lucide:arrow-left" width="16" height="16" />
                      {t("common.back")}
                    </Button>
                    <Flex align="center" gap="2">
                      <Icon icon="lucide:target" width="20" height="20" style={{ color: "var(--accent-9)" }} />
                      <Text size="3" weight="bold">
                        {t("scanner.scanTarget")}: {selectedDevice.ip}
                      </Text>
                      {selectedDevice.hostname && (
                        <Badge variant="soft" color="blue">{selectedDevice.hostname}</Badge>
                      )}
                    </Flex>
                  </Flex>

                  <Flex align="center" gap="2">
                    {scanning ? (
                      <Button color="red" variant="soft" onClick={stopScan}>
                        <Icon icon="lucide:square" width="14" height="14" />
                        {t("scanner.stop")}
                      </Button>
                    ) : (
                      <Button onClick={startPortScan}>
                        <Icon icon="lucide:play" width="14" height="14" />
                        {t("scanner.startScan")}
                      </Button>
                    )}
                  </Flex>
                </Flex>

                {/* Row 2: mode selector */}
                <Flex align="center" gap="3" wrap="wrap">
                  <Text size="2" color="gray" style={{ whiteSpace: "nowrap" }}>
                    {t("scanner.mode")}:
                  </Text>
                  <SegmentedControl.Root
                    value={scanMode}
                    onValueChange={(v) => setScanMode(v as ScanMode)}
                    disabled={scanning}
                    size="2"
                  >
                    <SegmentedControl.Item value="common">
                      <Flex align="center" gap="1">
                        <Icon icon="lucide:star" width="13" height="13" />
                        {t("scanner.modeCommon")}
                      </Flex>
                    </SegmentedControl.Item>
                    <SegmentedControl.Item value="custom">
                      <Flex align="center" gap="1">
                        <Icon icon="lucide:list" width="13" height="13" />
                        {t("scanner.modeCustom")}
                      </Flex>
                    </SegmentedControl.Item>
                    <SegmentedControl.Item value="full">
                      <Flex align="center" gap="1">
                        <Icon icon="lucide:infinity" width="13" height="13" />
                        {t("scanner.modeFull")}
                      </Flex>
                    </SegmentedControl.Item>
                  </SegmentedControl.Root>
                </Flex>

                {/* Row 3: custom ports input (only when mode=custom) */}
                {scanMode === "custom" && (
                  <Flex align="center" gap="2" wrap="wrap">
                    <Box style={{ flex: 1, minWidth: 200, maxWidth: 420 }}>
                      <TextField.Root
                        placeholder={t("scanner.customPortsPlaceholder")}
                        value={customPorts}
                        onChange={(e) => setCustomPorts(e.target.value)}
                        disabled={scanning}
                      />
                    </Box>
                    <Text size="1" color="gray">{t("scanner.customPortsHint")}</Text>
                  </Flex>
                )}

                {/* Row 4: full-scan warning */}
                {scanMode === "full" && (
                  <Callout.Root color="amber" size="1">
                    <Callout.Icon>
                      <Icon icon="lucide:triangle-alert" width="14" height="14" />
                    </Callout.Icon>
                    <Callout.Text>{t("scanner.fullScanWarning")}</Callout.Text>
                  </Callout.Root>
                )}
              </Flex>
            </Card>

            {/* Progress */}
            {(scanning || scannedCount > 0) && (
              <Flex direction="column" gap="2">
                <Flex align="center" justify="between" wrap="wrap" gap="2">
                  <Flex align="center" gap="3">
                    {scanning && <Spinner size="1" />}
                    <Text size="2" color="gray">
                      {t("scanner.scanned")}: {scannedCount}
                      {totalPorts > 0 && ` / ${totalPorts}`}
                    </Text>
                    <Badge color="green" variant="soft">
                      <Icon icon="lucide:unlock" width="12" height="12" />
                      {t("scanner.openPorts")}: {openPorts.length}
                    </Badge>
                    {scanDone && (
                      <Badge color="blue" variant="soft">
                        <Icon icon="lucide:check" width="12" height="12" />
                        {t("scanner.complete")}
                      </Badge>
                    )}
                  </Flex>
                  {totalPorts > 0 && (
                    <Text size="1" color="gray">{progressPct}%</Text>
                  )}
                </Flex>

                {totalPorts > 0 && (
                  <Progress value={progressPct} size="1" />
                )}
              </Flex>
            )}

            {/* Results */}
            <AnimatePresence>
              {openPorts.length > 0 && (
                <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }}>
                  <Card>
                    <Table.Root variant="surface">
                      <Table.Header>
                        <Table.Row>
                          <Table.ColumnHeaderCell>{t("scanner.port")}</Table.ColumnHeaderCell>
                          <Table.ColumnHeaderCell>{t("scanner.status")}</Table.ColumnHeaderCell>
                          <Table.ColumnHeaderCell>{t("scanner.service")}</Table.ColumnHeaderCell>
                        </Table.Row>
                      </Table.Header>
                      <Table.Body>
                        {[...openPorts]
                          .sort((a, b) => a.port - b.port)
                          .map((r) => (
                            <motion.tr
                              key={r.port}
                              initial={{ opacity: 0, x: -8 }}
                              animate={{ opacity: 1, x: 0 }}
                            >
                              <Table.Cell>
                                <Text size="2" weight="medium" style={{ fontFamily: "monospace" }}>
                                  {r.port}
                                </Text>
                              </Table.Cell>
                              <Table.Cell>
                                <Badge color="green" variant="soft">
                                  <Icon icon="lucide:unlock" width="12" height="12" />
                                  {t("scanner.open")}
                                </Badge>
                              </Table.Cell>
                              <Table.Cell>
                                <Text size="2" color="gray">{r.service || "—"}</Text>
                              </Table.Cell>
                            </motion.tr>
                          ))}
                      </Table.Body>
                    </Table.Root>
                  </Card>
                </motion.div>
              )}
            </AnimatePresence>
          </Flex>
        </motion.div>
      )}
    </Flex>
  );
}
