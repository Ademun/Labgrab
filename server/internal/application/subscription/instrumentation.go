package subscription

import "go.opentelemetry.io/otel"

var tracer = otel.Tracer("labgrab/internal/application/subscription")
