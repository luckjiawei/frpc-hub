import { useState, useEffect } from "react";
import { useNavigate, useParams } from "react-router-dom";
import pb from "../../../lib/pocketbase";
import { toast } from "sonner";

export interface IntegrationFormData {
  name: string;
  integrationsType: "cloudflare";
  credentials: string;
}

const defaultData: IntegrationFormData = {
  name: "",
  integrationsType: "cloudflare",
  credentials: "",
};

export function useIntegrationForm() {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const isEditing = !!id;

  const [formData, setFormData] = useState<IntegrationFormData>(defaultData);
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [submitting, setSubmitting] = useState(false);
  const [loadingRecord, setLoadingRecord] = useState(isEditing);

  useEffect(() => {
    if (!id) return;
    setLoadingRecord(true);
    pb.collection("fh_integrations")
      .getOne(id)
      .then((record) => {
        setFormData({
          name: record.name || "",
          integrationsType: record.integrationsType || "cloudflare",
          credentials: record.credentials || "",
        });
      })
      .catch(() => {
        toast.error("Failed to load integration");
        navigate("/integrations");
      })
      .finally(() => setLoadingRecord(false));
  }, [id, navigate]);

  const handleChange = (field: keyof IntegrationFormData, value: string) => {
    setFormData((prev) => ({ ...prev, [field]: value }));
    if (field === "name") {
      setErrors((prev) => ({ ...prev, name: value ? "" : "Required" }));
    }
    if (field === "credentials") {
      setErrors((prev) => ({ ...prev, credentials: value ? "" : "Required" }));
    }
  };

  const validate = (): boolean => {
    const newErrors: Record<string, string> = {};
    if (!formData.name.trim()) newErrors.name = "Required";
    if (!formData.credentials.trim()) newErrors.credentials = "Required";
    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async () => {
    if (!validate()) return;

    try {
      setSubmitting(true);
      const payload = {
        name: formData.name.trim(),
        integrationsType: formData.integrationsType,
        credentials: formData.credentials.trim(),
        status: "active",
      };

      if (isEditing) {
        await pb.collection("fh_integrations").update(id!, payload);
        toast.success("Integration updated successfully");
      } else {
        await pb.collection("fh_integrations").create(payload);
        toast.success("Integration created successfully");
      }
      navigate("/integrations");
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to save integration");
    } finally {
      setSubmitting(false);
    }
  };

  return {
    isEditing,
    formData,
    errors,
    submitting,
    loadingRecord,
    handleChange,
    handleSubmit,
    navigate,
  };
}
