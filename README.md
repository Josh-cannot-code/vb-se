# Vlogbrothers Search Engine

A modern search engine for the Vlogbrothers YouTube channel, featuring advanced search capabilities and a responsive user interface.

## Technology Stack

### Search Engine
- **OpenSearch**: Utilizing hybrid search capabilities for improved search relevance
  - Text-based search for video titles and descriptions
  - Semantic search on video transcripts for better understanding of user queries
  - Adjustable weights for scoring and ranking

### Infrastructure

#### Deployment
- **Kustomize**: Kubernetes native configuration management
  - Environment-specific overlays
  - Resource customization
  - Secure secret management with Bitnami Sealed Secrets
  - GitOps-friendly encrypted configuration

#### Search Infrastructure
- **OpenSearch**: Self-hosted on bare metal Kubernetes cluster
  - High-availability deployment with multiple master nodes
  - Replicated data for fault tolerance using longhorn

#### Ingress
- **Cloudflare Tunnels**: Zero-trust secure ingress
  - Sidecar container pattern for tunnel deployment
  - Handles consumer ISP ip lease expiration (no public ip required)
  - Automatic SSL certificate management

### Frontend
- **Go Templ**: Type-safe HTML templating
- **Tailwind CSS**: Utility-first CSS framework for responsive design
- **Echo**: High performance web framework for Go

### Development
- **Go**: Backend service with structured logging and dependency injection
- **Python**: Video data collection and processing
  - youtube-dl for metadata extraction
  - youtube-transcript-api for automated transcript fetching
- **YouTube API**: Fetching all video IDs for a channel