import type { ProblemKind } from "@app/api-client";
import { m } from "@app/i18n";

// One message for each kind of error (UX standard U14, U31): what happened
// and what to do. The texts are in packages/i18n/messages.
export const PROBLEM_MESSAGES: Record<ProblemKind, { title: () => string; action: () => string }> =
  {
    validation: { title: m.error_validation_title, action: m.error_validation_action },
    refused: { title: m.error_refused_title, action: m.error_refused_action },
    unauthenticated: {
      title: m.error_unauthenticated_title,
      action: m.error_unauthenticated_action,
    },
    forbidden: { title: m.error_forbidden_title, action: m.error_forbidden_action },
    not_found: { title: m.error_not_found_title, action: m.error_not_found_action },
    service: { title: m.error_service_title, action: m.error_service_action },
  };
