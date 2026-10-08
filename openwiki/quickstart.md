---
type: "Reference"
title: "AgentTeams OpenWiki"
description: "Entry point that routes readers to the right section based on their goal (understand, run, develop). Lists all wiki sections with descriptions."
tags: ["index", "navigation", "quickstart", "wiki", "overview"]
openwiki_generated: true
verified:
  - by: openwiki/0.7.1
    at: 2026-10-08T08:19:13.018Z
sources:
  - id: openwiki-source-8037e2358a2c4f9b2c722a11
    resource: repo://AGENTS.md
  - id: openwiki-source-4be8d89602ed92b69ba13da1
    resource: repo://manager/agent/AGENTS.md
  - id: openwiki-source-23775c3de52f3ab95a13cb8b
    resource: repo://README.md
generated: { by: "openwiki/0.7.1", at: "2026-10-08T08:19:13.018Z" }
---

# AgentTeams OpenWiki

AgentTeams is an open-source collaborative multi-agent runtime platform. A **Manager** agent coordinates multiple **Worker** agents in Matrix IM rooms, with full human visibility and intervention. The system does not implement agent logic itself — it orchestrates and manages agent containers across multiple runtimes.

## What This Wiki Covers

This wiki is the generated knowledge base for the AgentTeams repository. It explains the architecture, components, and workflows so that both humans and AI agents can navigate the codebase effectively.

| Section | What It Covers |
|---------|---------------|
| [Architecture Overview](architecture/overview.md) | System layers, deployment shapes, component relationships, CRD model, credential flow, and data flow diagrams |
| [Controller: CRDs & Reconcilers](controller/crds-and-reconcilers.md) | Go operator, 5 CRD types, reconciler logic, service/provisioner layers, backend abstraction |
| [Controller: CLI & API](controller/cli-and-api.md) | `agt` CLI commands, REST API endpoints, lifecycle operations |
| [Worker Runtimes](workers/runtime-guide.md) | Five worker runtimes (OpenClaw, CoPaw, Hermes, OpenHuman, QwenPaw) — configuration and differences |
| [Manager](manager/overview.md) | Coordinator agent, 21 skills, bootstrap chain, agent config templates |
| [Plugin System](integrations/plugin-system.md) | TeamHarness and WorkerFlow plugin architecture, plugin.yaml contract, adapter pattern, CLI installation |
| [Shared Libraries](integrations/shared-libraries.md) | Five shared Python packages and shell libraries providing domain logic across all runtimes |
| [Operations: Install & Deploy](operations/installation.md) | Docker/Podman local install, Kubernetes Helm install, upgrade, uninstall procedures |
| [Operations: Dashboard](operations/dashboard.md) | Web dashboard architecture: Vite SPA, Node.js proxy, allowlist security model, milestone features |
| [Development: Build & Test](development/build-and-test.md) | Makefile targets, image build chain, CI/CD, testing, contribution guidelines |

## Quick Orientation

**If you want to understand the system:**
1. Start with [Architecture Overview](architecture/overview.md) for the big picture
2. Read [Controller: CRDs & Reconcilers](controller/crds-and-reconcilers.md) for the core data model
3. Browse [Worker Runtimes](workers/runtime-guide.md) to understand the multi-runtime design

**If you want to run the system:**
1. Follow [Operations: Install & Deploy](operations/installation.md) for installation
2. Use [Controller: CLI & API](controller/cli-and-api.md) for day-to-day operations
3. Explore the [Dashboard](operations/dashboard.md) for web-based monitoring and interaction

**If you want to develop or modify code:**
1. Read [Development: Build & Test](development/build-and-test.md) for build pipeline and CI
2. See [Manager](manager/overview.md) for agent-facing content and skills
3. Check [Controller: CRDs & Reconcilers](controller/crds-and-reconcilers.md) for operator development

**If you want to extend AgentTeams with plugins:**
1. Understand the [Plugin System](integrations/plugin-system.md) for TeamHarness and WorkerFlow extensions
2. Review [Shared Libraries](integrations/shared-libraries.md) to reuse common domain logic across runtimes
3. Consult [Worker Runtimes](workers/runtime-guide.md) for runtime-specific integration points

## Repository Map

```
AgentTeams/
├── agentteams-controller/   # Go operator: CRDs, reconcilers, CLI, REST API
├── helm/agentteams/         # Helm chart for Kubernetes deployment
├── manager/             # Manager images, agent configs, 21 skills, bootstrap scripts
├── worker/              # OpenClaw Worker base image
├── copaw/               # CoPaw Python worker runtime
├── hermes/              # Hermes Python worker runtime
├── openhuman/           # OpenHuman Rust worker runtime
├── qwenpaw/             # QwenPaw worker runtime (TeamHarness plugin)
├── openclaw-base/       # Base image: Ubuntu + Node.js + OpenClaw + mcporter
├── plugins/             # TeamHarness, WorkerFlow, runtime adapters
├── shared/              # Shell libs + 5 Python packages (protocol, sync, merge, format, policies)
├── dashboard/           # Web dashboard (Vite SPA + Node.js proxy)
├── install/             # Local install scripts (Bash, PowerShell)
├── docs/                # Existing upstream documentation
├── tests/               # Integration tests
└── .github/workflows/   # CI/CD: build, test, release
```

## Key Technology Stack

| Component | Technology | Purpose |
|-----------|-----------|---------|
| Kubernetes operator | Go (agentteams-controller) | Reconciles Worker/Manager/Team/Human/Project CRDs |
| AI Gateway | Higress | LLM proxy, MCP server hosting, consumer auth |
| Matrix Server | Tuwunel (conduwuit fork) | IM between agents and humans |
| File System | MinIO or Alibaba Cloud OSS | Centralized object storage for workspaces |
| Agent Frameworks | OpenClaw, CoPaw, Hermes, OpenHuman, QwenPaw | Multiple runtimes for different agent capabilities |
| MCP CLI | mcporter | Worker calls MCP server tools via CLI |
| Dashboard | Vite SPA + Node.js proxy | Web UI for monitoring and interaction |
| Plugin System | TeamHarness, WorkerFlow | Runtime extension via standardized plugin contract |

## Deployment Modes

- **Local (Docker/Podman):** One embedded controller container bundles Higress + Tuwunel + MinIO + Element Web + controller process. Manager and Worker run as separate containers via Docker API.
- **Kubernetes (Helm):** Each component runs as its own Pod. The controller reconciles CRDs to create Manager and Worker pods dynamically.

See [Architecture Overview](architecture/overview.md) for detailed deployment diagrams.

## Upstream Documentation

The `docs/` directory contains the original project documentation:

- [`docs/architecture.md`](../docs/architecture.md) — System architecture and component diagrams
- [`docs/quickstart.md`](../docs/quickstart.md) — End-to-end setup guide
- [`docs/development.md`](../docs/development.md) — Developer workflow
- [`docs/declarative-resource-management.md`](../docs/declarative-resource-management.md) — YAML-driven resource management
- [`docs/worker-guide.md`](../docs/worker-guide.md) — Worker development guide
- [`docs/manager-guide.md`](../docs/manager-guide.md) — Manager administration guide
- [`docs/faq.md`](../docs/faq.md) — Frequently asked questions
- [`docs/zh-cn/`](../docs/zh-cn/) — Chinese-language documentation mirror

## Git Context

- **Repository:** `agentscope-ai/AgentTeams` (GitHub)
- **Latest documented commit:** `f37b50c` — feat(agt): rewrite status as one-screen cluster overview
- **Key recent work:** Gastown-inspired features (escalation, health monitoring, dispatch gating, session recovery), QwenPaw runtime wiring, plugin platform (TeamHarness, WorkerFlow), dashboard web UI, remediation gates CI
