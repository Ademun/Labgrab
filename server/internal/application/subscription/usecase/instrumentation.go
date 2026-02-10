package usecase

import "go.opentelemetry.io/otel"

var tracer = otel.Tracer("labgrab/internal/application/subscription/usecase")
