import {
  Button,
  Card,
  Flex,
  Heading,
  Text,
  Badge,
  Table,
  TextField,
  AlertDialog,
} from "@radix-ui/themes";
import { useNavigate } from "react-router-dom";
import { useState, useEffect } from "react";
import { Icon } from "@iconify/react";
import { motion, AnimatePresence } from "framer-motion";
import { AnimateNumber } from "../../components/AnimateNumber";
import { EmptyState } from "../../components/EmptyState";
import type { Integration } from "./useIntegrations";
import { PageHeader } from "../../components/PageHeader";
import { Loading } from "../../components/Loading";
import { StatCard } from "../../components/StatCard";
import { useTranslation } from "react-i18next";

interface IntegrationsViewProps {
  integrations: Integration[];
  loading?: boolean;
  refreshing: boolean;
  deleteIntegration: (id: string) => void;
  validateIntegration: (id: string) => void;
  validatingId: string | null;
  search: string;
  setSearch: (value: string) => void;
  refreshIntegrations: () => void;
  page: number;
  setPage: (page: number) => void;
  totalPages: number;
}

const STATUS_COLOR: Record<string, "green" | "red" | "orange" | "gray"> = {
  active: "green",
  invalid: "red",
  expired: "orange",
};

export function IntegrationsView({
  integrations,
  loading,
  refreshing,
  deleteIntegration,
  validateIntegration,
  validatingId,
  search,
  setSearch,
  refreshIntegrations,
  page,
  setPage,
  totalPages,
}: IntegrationsViewProps) {
  const { t } = useTranslation();
  const navigate = useNavigate();

  const [deleteId, setDeleteId] = useState<string | null>(null);
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    const timer = setTimeout(() => setMounted(true), 100);
    return () => clearTimeout(timer);
  }, []);

  return (
    <>
      <Flex direction="column" gap="5" className="flex flex-1 flex-col">
        {/* Header */}
        <PageHeader
          title={t("integration.title")}
          description={t("integration.manageIntegrations")}
          visible={mounted}
        />

        {/* Stats Cards */}
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={mounted ? { opacity: 1, y: 0 } : { opacity: 0, y: 20 }}
            transition={{ duration: 0.5, delay: 0.1, ease: "easeOut" }}
          >
            <StatCard
              title={t("integration.totalIntegrations")}
              value={<AnimateNumber value={integrations.length} decimalPlaces={0} />}
              color="gray"
              icon={<Icon icon="lucide:plug" width="32" height="32" color="var(--gray-11)" />}
            />
          </motion.div>
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={mounted ? { opacity: 1, y: 0 } : { opacity: 0, y: 20 }}
            transition={{ duration: 0.5, delay: 0.2, ease: "easeOut" }}
          >
            <StatCard
              title={t("integration.active")}
              value={
                <span className="text-[#30A46C]">
                  <AnimateNumber
                    value={integrations.filter((i) => i.status === "active").length}
                    decimalPlaces={0}
                  />
                </span>
              }
              color="green"
              icon={<Icon icon="lucide:check-circle" width="32" height="32" color="var(--green-11)" />}
            />
          </motion.div>
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={mounted ? { opacity: 1, y: 0 } : { opacity: 0, y: 20 }}
            transition={{ duration: 0.5, delay: 0.3, ease: "easeOut" }}
          >
            <StatCard
              title={t("integration.inactive")}
              value={
                <span className="text-[#E5484D]">
                  <AnimateNumber
                    value={integrations.filter((i) => i.status !== "active").length}
                    decimalPlaces={0}
                  />
                </span>
              }
              color="red"
              icon={<Icon icon="lucide:x-circle" width="32" height="32" color="var(--red-11)" />}
            />
          </motion.div>
        </div>

        {/* Table */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={mounted ? { opacity: 1, y: 0 } : { opacity: 0, y: 20 }}
          transition={{ duration: 0.5, delay: 0.4, ease: "easeOut" }}
          className="flex flex-1 flex-col overflow-hidden"
        >
          <Card className="flex flex-1 flex-col">
            <Flex direction="column" gap="3" mb="4" p="2">
              <Flex justify="between" align="center" wrap="wrap" gap="2">
                <Text size="3" weight="bold">
                  {t("integration.allIntegrations")}
                </Text>
                <Flex gap="2" wrap="wrap" align="center">
                  <TextField.Root
                    size="2"
                    placeholder={t("integration.searchIntegrations")}
                    className="min-w-32 flex-1"
                    value={search}
                    onChange={(e) => setSearch(e.target.value)}
                  >
                    <TextField.Slot>
                      <Icon icon="lucide:search" width="16" height="16" />
                    </TextField.Slot>
                    {search && (
                      <TextField.Slot
                        pr="2"
                        onClick={() => setSearch("")}
                        style={{ cursor: "pointer" }}
                      >
                        <Icon icon="lucide:x" width="16" height="16" />
                      </TextField.Slot>
                    )}
                  </TextField.Root>
                  <Button variant="soft" onClick={refreshIntegrations} disabled={refreshing}>
                    <Icon
                      icon="lucide:refresh-cw"
                      width="14"
                      height="14"
                      className={refreshing ? "animate-spin" : ""}
                    />
                    {t("integration.refresh")}
                  </Button>
                  <Button onClick={() => navigate("/integrations/new")}>
                    <Icon icon="lucide:plus" width="14" height="14" />
                    {t("integration.addIntegration")}
                  </Button>
                </Flex>
              </Flex>
            </Flex>

            <AnimatePresence mode="wait">
              {loading ? (
                <motion.div
                  key="loading"
                  initial={{ opacity: 0 }}
                  animate={{ opacity: 1 }}
                  exit={{ opacity: 0 }}
                  transition={{ duration: 0.2 }}
                  className="flex-1"
                >
                  <Flex
                    direction="column"
                    align="center"
                    justify="center"
                    gap="2"
                    className="h-full py-12"
                  >
                    <Loading size="small" />
                  </Flex>
                </motion.div>
              ) : integrations.length === 0 ? (
                search ? (
                  <motion.div
                    key="no-results"
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1 }}
                    exit={{ opacity: 0 }}
                    transition={{ duration: 0.2 }}
                    className="flex-1"
                  >
                    <Flex
                      direction="column"
                      align="center"
                      justify="center"
                      gap="2"
                      className="h-full"
                    >
                      <Icon icon="lucide:search" width="32" height="32" color="gray" />
                      <Heading size="4">{t("integration.noResultsFound")}</Heading>
                      <Text color="gray" size="2">
                        {t("integration.noIntegrationsMatch")} "{search}"
                      </Text>
                      <Button variant="soft" mt="2" onClick={() => setSearch("")}>
                        {t("integration.clearSearch")}
                      </Button>
                    </Flex>
                  </motion.div>
                ) : (
                  <motion.div
                    key="empty-state"
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1 }}
                    exit={{ opacity: 0 }}
                    transition={{ duration: 0.2 }}
                    className="flex-1"
                  >
                    <Flex direction="column" align="center" justify="center" className="h-full">
                      <EmptyState
                        title={t("integration.noIntegrations")}
                        description={t("integration.addFirstIntegration")}
                        actionText={t("integration.addYourFirstIntegration")}
                        onAction={() => navigate("/integrations/new")}
                      />
                    </Flex>
                  </motion.div>
                )
              ) : (
                <motion.div
                  key="table"
                  initial={{ opacity: 0 }}
                  animate={{ opacity: 1 }}
                  exit={{ opacity: 0 }}
                  transition={{ duration: 0.2 }}
                >
                  <div className="overflow-x-auto">
                    <Table.Root>
                      <Table.Header>
                        <Table.Row>
                          <Table.ColumnHeaderCell>{t("common.name")}</Table.ColumnHeaderCell>
                          <Table.ColumnHeaderCell>{t("common.type")}</Table.ColumnHeaderCell>
                          <Table.ColumnHeaderCell>{t("common.status")}</Table.ColumnHeaderCell>
                          <Table.ColumnHeaderCell>{t("common.created")}</Table.ColumnHeaderCell>
                          <Table.ColumnHeaderCell>{t("common.actions")}</Table.ColumnHeaderCell>
                        </Table.Row>
                      </Table.Header>
                      <Table.Body>
                        {integrations.map((integration) => (
                          <Table.Row key={integration.id}>
                            <Table.Cell>
                              <Text weight="medium">{integration.name}</Text>
                            </Table.Cell>
                            <Table.Cell>
                              <Flex align="center" gap="1">
                                <Icon icon="simple-icons:cloudflare" width="16" height="16" color="var(--orange-11)" />
                                <Badge variant="surface" color="orange">
                                  {integration.integrationsType}
                                </Badge>
                              </Flex>
                            </Table.Cell>
                            <Table.Cell>
                              <Badge
                                radius="full"
                                color={STATUS_COLOR[integration.status] ?? "gray"}
                              >
                                <span
                                  className={`status-dot ${
                                    integration.status === "active"
                                      ? "status-dot-green"
                                      : "status-dot-red"
                                  }`}
                                />
                                {t(`integration.status_${integration.status}`)}
                              </Badge>
                            </Table.Cell>
                            <Table.Cell>
                              <Text size="2" color="gray">
                                {new Date(integration.created).toLocaleDateString()}
                              </Text>
                            </Table.Cell>
                            <Table.Cell>
                              <Flex gap="2">
                                <Button
                                  size="1"
                                  variant="soft"
                                  color="green"
                                  className="cursor-pointer"
                                  disabled={validatingId === integration.id}
                                  onClick={() => validateIntegration(integration.id)}
                                >
                                  <Icon
                                    icon={validatingId === integration.id ? "lucide:loader" : "lucide:shield-check"}
                                    width="14"
                                    height="14"
                                    className={validatingId === integration.id ? "animate-spin" : ""}
                                  />
                                  {t("integration.validate")}
                                </Button>
                                <Button
                                  size="1"
                                  variant="soft"
                                  className="cursor-pointer"
                                  onClick={() => navigate(`/integrations/${integration.id}/edit`)}
                                >
                                  <Icon icon="lucide:pencil" width="14" height="14" />
                                  {t("common.edit")}
                                </Button>
                                <Button
                                  size="1"
                                  variant="soft"
                                  className="cursor-pointer"
                                  color="red"
                                  onClick={() => setDeleteId(integration.id)}
                                >
                                  <Icon icon="lucide:trash" width="14" height="14" />
                                  {t("common.delete")}
                                </Button>
                              </Flex>
                            </Table.Cell>
                          </Table.Row>
                        ))}
                      </Table.Body>
                    </Table.Root>
                  </div>

                  {totalPages > 1 && (
                    <Flex justify="between" align="center" pt="4" px="1">
                      <Text size="2" color="gray">
                        {page} / {totalPages}
                      </Text>
                      <Flex
                        align="center"
                        style={{
                          border: "1px solid var(--gray-6)",
                          borderRadius: "var(--radius-3)",
                          overflow: "hidden",
                        }}
                      >
                        <button
                          disabled={page === 1}
                          onClick={() => setPage(page - 1)}
                          style={{
                            padding: "6px 10px",
                            background: "none",
                            border: "none",
                            borderRight: "1px solid var(--gray-6)",
                            cursor: page === 1 ? "not-allowed" : "pointer",
                            opacity: page === 1 ? 0.4 : 1,
                            display: "flex",
                            alignItems: "center",
                          }}
                        >
                          <Icon icon="lucide:chevron-left" width="14" height="14" />
                        </button>
                        {Array.from({ length: totalPages }, (_, i) => i + 1).map((p) => (
                          <button
                            key={p}
                            onClick={() => setPage(p)}
                            style={{
                              padding: "6px 12px",
                              background: p === page ? "var(--accent-9)" : "none",
                              color: p === page ? "white" : "inherit",
                              border: "none",
                              borderRight: p === totalPages ? "none" : "1px solid var(--gray-6)",
                              cursor: "pointer",
                              fontSize: "var(--font-size-2)",
                              fontWeight: p === page ? 600 : 400,
                            }}
                          >
                            {p}
                          </button>
                        ))}
                        <button
                          disabled={page === totalPages}
                          onClick={() => setPage(page + 1)}
                          style={{
                            padding: "6px 10px",
                            background: "none",
                            border: "none",
                            borderLeft: "1px solid var(--gray-6)",
                            cursor: page === totalPages ? "not-allowed" : "pointer",
                            opacity: page === totalPages ? 0.4 : 1,
                            display: "flex",
                            alignItems: "center",
                          }}
                        >
                          <Icon icon="lucide:chevron-right" width="14" height="14" />
                        </button>
                      </Flex>
                    </Flex>
                  )}
                </motion.div>
              )}
            </AnimatePresence>
          </Card>
        </motion.div>
      </Flex>

      <AlertDialog.Root open={!!deleteId} onOpenChange={() => setDeleteId(null)}>
        <AlertDialog.Content maxWidth="450px">
          <AlertDialog.Title>{t("integration.confirmDeletion")}</AlertDialog.Title>
          <AlertDialog.Description size="2">
            {t("integration.deleteConfirmMessage")}
          </AlertDialog.Description>
          <Flex gap="3" mt="4" justify="end">
            <AlertDialog.Cancel>
              <Button variant="soft" color="gray">
                {t("common.cancel")}
              </Button>
            </AlertDialog.Cancel>
            <AlertDialog.Action>
              <Button
                variant="solid"
                color="red"
                onClick={() => {
                  if (deleteId) {
                    deleteIntegration(deleteId);
                    setDeleteId(null);
                  }
                }}
              >
                {t("common.delete")}
              </Button>
            </AlertDialog.Action>
          </Flex>
        </AlertDialog.Content>
      </AlertDialog.Root>
    </>
  );
}
