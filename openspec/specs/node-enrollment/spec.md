# node-enrollment Specification

## Purpose
Give administrators a single Nodes-page workflow for enrolling Native and Docker Agents into an existing Hub using the real registration protocol and copyable commands.
## Requirements
### Requirement: Show enrollment methods

The Nodes page SHALL provide an administrator-only Add Node workflow with required node name and node address fields accepting an IPv4/IPv6 address or DNS hostname, an editable Hub address defaulted to the current web page origin, and Native, `docker run`, and standalone `docker compose` methods.

#### Scenario: Open Add Node

- **WHEN** an administrator opens Add Node
- **THEN** the page shows required node name and node address inputs, a Hub address prefilled from the current page origin, and the three enrollment methods

#### Scenario: Generate enrollment commands

- **WHEN** the administrator submits a valid node name, IPv4/IPv6 address or DNS hostname, and Hub address
- **THEN** the Hub returns commands containing the node identity, the supplied address unchanged, registration token, and an Agent Gateway URL derived from the supplied Hub address and configured Gateway port

#### Scenario: Reject invalid node identity

- **WHEN** the administrator submits an empty/invalid node name or an invalid node address
- **THEN** the Hub rejects the request without generating commands

#### Scenario: Copy enrollment command

- **WHEN** the administrator copies a displayed command or Compose document
- **THEN** the exact command text is copied on both secure and ordinary HTTP origins without placing credentials in persistent browser storage

#### Scenario: Clipboard API unavailable

- **WHEN** the browser does not expose or rejects `navigator.clipboard.writeText`
- **THEN** the UI uses a browser-native fallback and reports an error only if both copy paths fail

### Requirement: Use the existing registration flow

Enrollment instructions SHALL configure the existing Agent binary/container with the Hub URL, registration token, node name, and node IP, preserve persistent Agent state, and not create a synthetic node.

#### Scenario: Agent reports configured identity

- **WHEN** an Agent starts with generated enrollment configuration
- **THEN** its first `HELLO` reports the configured node name and IP, falling back to local discovery only when those values are absent

#### Scenario: Native enrollment

- **WHEN** an administrator runs the Native instructions on a Linux host
- **THEN** the Agent starts with the generated configuration, sends `HELLO`, and the resulting real node appears in the Nodes list after acceptance

#### Scenario: Docker enrollment

- **WHEN** an administrator runs the Docker or Compose instructions against an existing Hub
- **THEN** the container uses a persistent `/var/lib/cadentra` volume, connects to the configured Gateway, and registers as a real node

### Requirement: Protect enrollment metadata

The enrollment metadata endpoint SHALL require administrator authorization and SHALL return a clear error when the Hub Gateway URL is not configured or derivable.

#### Scenario: Unauthorized enrollment metadata

- **WHEN** a Viewer, unauthenticated caller, or invalid session requests enrollment metadata
- **THEN** the Hub returns an authorization error and does not expose the registration token
