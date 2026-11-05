import { createApiClient } from "@app/api-client";
import { CORE_URL } from "astro:env/server";

export const core = createApiClient(CORE_URL);
