import { useState, useEffect, useCallback, useRef } from "react";
import pb from "../../lib/pocketbase";
import { apiPost } from "../../lib/api";
import { toast } from "sonner";

export interface Integration {
  id: string;
  name: string;
  integrationsType: "cloudflare";
  metadata: Record<string, string>;
  credentials: string;
  status: "active" | "invalid" | "expired";
  created: string;
  updated: string;
}

export function useIntegrations() {
  const [integrations, setIntegrations] = useState<Integration[]>([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const initializedRef = useRef(false);

  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const PER_PAGE = 10;

  const [search, setSearch] = useState("");

  const fetchIntegrations = useCallback(
    async (isRefresh = false) => {
      try {
        if (!initializedRef.current) {
          setLoading(true);
        } else if (isRefresh) {
          setRefreshing(true);
        }

        let filter = "";
        if (search) {
          filter = `name ~ "${search}"`;
        }

        const result = await pb.collection("fh_integrations").getList(page, PER_PAGE, {
          filter,
          sort: "-created",
        });

        setIntegrations(result.items as unknown as Integration[]);
        setTotalPages(result.totalPages);
        initializedRef.current = true;
      } catch (err) {
        if ((err as any)?.isAbort) return;
        toast.error(err instanceof Error ? err.message : "Failed to fetch integrations");
      } finally {
        setLoading(false);
        setRefreshing(false);
      }
    },
    [page, search]
  );

  useEffect(() => {
    setPage(1);
    if (initializedRef.current) {
      setLoading(true);
    }
  }, [search]);

  useEffect(() => {
    const timer = setTimeout(() => {
      fetchIntegrations();
    }, 300);
    return () => clearTimeout(timer);
  }, [fetchIntegrations]);

  const [validatingId, setValidatingId] = useState<string | null>(null);

  const deleteIntegration = async (id: string) => {
    try {
      await pb.collection("fh_integrations").delete(id);
      await fetchIntegrations();
      toast.success("Integration deleted successfully");
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to delete integration");
    }
  };

  const validateIntegration = async (id: string) => {
    try {
      setValidatingId(id);
      const res = await apiPost("/api/integrations/validate", { id });
      const data = await res.json();
      if (!res.ok) {
        toast.error(data.error || "Validation failed");
      } else {
        toast.success("Validation successful");
      }
      await fetchIntegrations();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Validation failed");
    } finally {
      setValidatingId(null);
    }
  };

  return {
    integrations,
    loading,
    refreshing,
    deleteIntegration,
    validateIntegration,
    validatingId,
    page,
    setPage,
    totalPages,
    search,
    setSearch,
    refreshIntegrations: () => fetchIntegrations(true),
  };
}
