import { useState, useEffect } from "react";
import { useServerForm } from "./useServerForm";
import { ServerFormPageView } from "./ServerFormPage.view";

export function ServerFormPage() {
  const form = useServerForm();
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    const timer = setTimeout(() => setMounted(true), 100);
    return () => clearTimeout(timer);
  }, []);

  return (
    <ServerFormPageView
      isEditing={form.isEditing}
      formData={form.formData}
      errors={form.errors}
      submitting={form.submitting}
      loadingServer={form.loadingServer}
      frpVersion={form.frpVersion}
      mounted={mounted}
      onChange={form.handleChange}
      onAuthChange={form.handleAuthChange}
      onLogChange={form.handleLogChange}
      onTransportChange={form.handleTransportChange}
      onTlsChange={form.handleTlsChange}
      onTlsStringChange={form.handleTlsStringChange}
      onFileUpload={form.handleFileUpload}
      onSetFormData={form.setFormData}
      onSubmit={form.handleSubmit}
      onCancel={() => form.navigate("/servers")}
    />
  );
}

export default ServerFormPage;
