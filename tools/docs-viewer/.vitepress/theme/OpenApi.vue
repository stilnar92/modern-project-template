<script setup>
// The HTTP API page. It reads api/openapi.yaml at build time, so no second
// command is necessary. Telemetry and the request client are off, so the
// page sends no request to the network.
import { ApiReference } from '@scalar/api-reference'
import '@scalar/api-reference/style.css'
import spec from '../../../../api/openapi.yaml?raw'

const configuration = {
  content: spec,
  withDefaultFonts: false,
  hideClientButton: true,
  hideTestRequestButton: true,
  telemetry: false,
}
</script>

<template><ApiReference :configuration="configuration" /></template>
