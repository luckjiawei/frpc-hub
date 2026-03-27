import { useIntegrations } from "./useIntegrations";
import { IntegrationsView } from "./IntegrationsPage.view";

export function IntegrationsPage() {
  const state = useIntegrations();

  return <IntegrationsView {...state} />;
}

export default IntegrationsPage;
