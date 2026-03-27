import { useState, useEffect } from "react";
import { useIntegrationForm } from "./useIntegrationForm";
import { IntegrationFormPageView } from "./IntegrationFormPage.view";

export function IntegrationFormPage() {
  const form = useIntegrationForm();
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    const timer = setTimeout(() => setMounted(true), 100);
    return () => clearTimeout(timer);
  }, []);

  return (
    <IntegrationFormPageView
      isEditing={form.isEditing}
      formData={form.formData}
      errors={form.errors}
      submitting={form.submitting}
      loadingRecord={form.loadingRecord}
      mounted={mounted}
      onChange={form.handleChange}
      onSubmit={form.handleSubmit}
      onCancel={() => form.navigate("/integrations")}
    />
  );
}

export default IntegrationFormPage;
