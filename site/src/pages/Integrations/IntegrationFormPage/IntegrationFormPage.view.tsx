import {
  Button,
  Card,
  Flex,
  Text,
  TextField,
  Separator,
} from "@radix-ui/themes";
import { Icon } from "@iconify/react";
import { motion } from "framer-motion";
import { PageHeader } from "../../../components/PageHeader";
import { Loading } from "../../../components/Loading";
import { useTranslation } from "react-i18next";
import type { IntegrationFormData } from "./useIntegrationForm";

interface IntegrationFormPageViewProps {
  isEditing: boolean;
  formData: IntegrationFormData;
  errors: Record<string, string>;
  submitting: boolean;
  loadingRecord: boolean;
  mounted: boolean;
  onChange: (field: keyof IntegrationFormData, value: string) => void;
  onSubmit: () => void;
  onCancel: () => void;
}

function SectionHeading({ title }: { title: string }) {
  return (
    <Text size="2" weight="bold" color="gray" style={{ textTransform: "uppercase", letterSpacing: "0.05em" }}>
      {title}
    </Text>
  );
}

function FormItem({
  label,
  required,
  error,
  children,
}: {
  label: string;
  required?: boolean;
  error?: string;
  children: React.ReactNode;
}) {
  return (
    <Flex direction="column" gap="1">
      <Text size="2" weight="medium">
        {label}
        {required && (
          <Text color="red" ml="1">
            *
          </Text>
        )}
      </Text>
      {children}
      {error && (
        <Text size="1" color="red">
          {error}
        </Text>
      )}
    </Flex>
  );
}

export function IntegrationFormPageView({
  isEditing,
  formData,
  errors,
  submitting,
  loadingRecord,
  mounted,
  onChange,
  onSubmit,
  onCancel,
}: IntegrationFormPageViewProps) {
  const { t } = useTranslation();

  if (loadingRecord) {
    return (
      <Flex align="center" justify="center" className="h-full py-20">
        <Loading size="small" />
      </Flex>
    );
  }

  return (
    <Flex direction="column" gap="5" className="flex flex-1 flex-col">
      <PageHeader
        title={isEditing ? t("integration.editIntegration") : t("integration.addIntegration")}
        description={
          isEditing ? t("integration.updateIntegrationDesc") : t("integration.addIntegrationDesc")
        }
        visible={mounted}
      />

      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={mounted ? { opacity: 1, y: 0 } : { opacity: 0, y: 20 }}
        transition={{ duration: 0.5, delay: 0.1, ease: "easeOut" }}
      >
        <Card>
          <Flex direction="column" gap="5" p="2">
            {/* Basic Info */}
            <Flex direction="column" gap="3">
              <SectionHeading title={t("integration.sectionBasic")} />
              <Separator size="4" />

              <FormItem label={t("common.name")} required error={errors.name}>
                <TextField.Root
                  placeholder={t("integration.namePlaceholder")}
                  value={formData.name}
                  onChange={(e) => onChange("name", e.target.value)}
                />
              </FormItem>

              <FormItem label={t("common.type")}>
                <Flex
                  align="center"
                  gap="3"
                  p="3"
                  style={{
                    border: "2px solid var(--orange-8)",
                    borderRadius: "var(--radius-3)",
                    backgroundColor: "var(--orange-a2)",
                  }}
                >
                  <Icon icon="simple-icons:cloudflare" width="24" height="24" color="var(--orange-11)" />
                  <Flex direction="column" gap="0">
                    <Text size="2" weight="bold">
                      Cloudflare
                    </Text>
                    <Text size="1" color="gray">
                      {t("integration.cloudflareDesc")}
                    </Text>
                  </Flex>
                </Flex>
              </FormItem>

            </Flex>

            {/* Credentials */}
            <Flex direction="column" gap="3">
              <SectionHeading title={t("integration.sectionCredentials")} />
              <Separator size="4" />

              <FormItem
                label={t("integration.apiToken")}
                required
                error={errors.credentials}
              >
                <TextField.Root
                  type="password"
                  placeholder={t("integration.apiTokenPlaceholder")}
                  value={formData.credentials}
                  onChange={(e) => onChange("credentials", e.target.value)}
                />
              </FormItem>
            </Flex>

            {/* Actions */}
            <Flex gap="3" justify="end">
              <Button variant="soft" color="gray" onClick={onCancel} disabled={submitting}>
                {t("common.cancel")}
              </Button>
              <Button onClick={onSubmit} disabled={submitting}>
                {submitting ? (
                  <>
                    <Icon icon="lucide:loader" width="14" height="14" className="animate-spin" />
                    {t("integration.saving")}
                  </>
                ) : isEditing ? (
                  t("integration.saveChanges")
                ) : (
                  t("integration.addIntegration")
                )}
              </Button>
            </Flex>
          </Flex>
        </Card>
      </motion.div>
    </Flex>
  );
}
