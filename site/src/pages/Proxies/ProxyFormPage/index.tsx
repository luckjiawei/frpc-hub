import { useState, useEffect } from "react";
import { useProxyForm } from "./useProxyForm";
import { ProxyFormPageView } from "./ProxyFormPage.view";

export function ProxyFormPage() {
  const form = useProxyForm();
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    const timer = setTimeout(() => setMounted(true), 100);
    return () => clearTimeout(timer);
  }, []);

  return (
    <ProxyFormPageView
      isEditing={form.isEditing}
      formData={form.formData}
      servers={form.servers}
      integrations={form.integrations}
      loadingServers={form.loadingServers}
      loadingIntegrations={form.loadingIntegrations}
      loadingProxy={form.loadingProxy}
      submitting={form.submitting}
      errors={form.errors}
      isHttpType={form.isHttpType}
      isSocks5Plugin={form.isSocks5Plugin}
      isCloudflare={form.isCloudflare}
      mounted={mounted}
      onChange={form.handleChange}
      onSubmit={form.handleSubmit}
      onCancel={() => form.navigate("/proxies")}
      onNavigateToServers={() => form.navigate("/servers")}
      onNavigateToIntegrations={() => form.navigate("/integrations")}
    />
  );
}

export default ProxyFormPage;
