# OpenTelemetry initialization for Python ADK Runtime
# According to spec.md: 統一可觀察性，單一 OTLP 端點導出到 Alloy/Collector
import logging
import os
from typing import Optional, Callable

from opentelemetry import trace, metrics
from opentelemetry.sdk.resources import Resource
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor, ConsoleSpanExporter
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
from opentelemetry.sdk.metrics import MeterProvider
from opentelemetry.exporter.otlp.proto.grpc.metric_exporter import OTLPMetricExporter
from opentelemetry.sdk.metrics.export import PeriodicExportingMetricReader, ConsoleMetricExporter
from opentelemetry.semconv.resource import ResourceAttributes

logger = logging.getLogger(__name__)

# Global cleanup function reference
_cleanup_func: Optional[Callable[[], None]] = None

def init_otel(mode: str = "lgtm_local", endpoint: str = "http://localhost:4317", 
              service_name: str = "detectviz-adk-runtime", service_version: str = "1.0.0") -> Callable[[], None]:
    """Initialize OpenTelemetry with unified observability according to spec.md"""
    global _cleanup_func
    
    try:
        logger.info(f"🔍 Initializing OpenTelemetry: mode={mode}, endpoint={endpoint}")
        
        # Create resource with required attributes according to spec.md
        resource = Resource.create({
            ResourceAttributes.SERVICE_NAME: service_name,
            ResourceAttributes.SERVICE_VERSION: service_version,
            ResourceAttributes.SERVICE_NAMESPACE: "detectviz",
            ResourceAttributes.DEPLOYMENT_ENVIRONMENT: os.getenv("DETECTVIZ_ENV", "development"),
            "detectviz.component": "adk-runtime",
            "detectviz.language": "python",
            "detectviz.observability.mode": mode
        })
        
        # Initialize tracing
        tracer_provider, trace_cleanup = _init_tracing(mode, endpoint, resource)
        trace.set_tracer_provider(tracer_provider)
        
        # Initialize metrics
        meter_provider, metrics_cleanup = _init_metrics(mode, endpoint, resource)
        metrics.set_meter_provider(meter_provider)
        
        # Create unified cleanup function
        def cleanup():
            logger.info("🎉 Shutting down OpenTelemetry...")
            if trace_cleanup:
                trace_cleanup()
            if metrics_cleanup:
                metrics_cleanup()
            logger.info("✅ OpenTelemetry shutdown completed")
        
        _cleanup_func = cleanup
        
        logger.info(f"✅ OpenTelemetry initialized successfully")
        logger.info(f"📊 Tracing and metrics export to: {endpoint}")
        
        return cleanup
        
    except Exception as e:
        logger.error(f"❌ Failed to initialize OpenTelemetry: {e}")
        
        # Return no-op cleanup function
        def noop_cleanup():
            pass
        return noop_cleanup

def _init_tracing(mode: str, endpoint: str, resource: Resource) -> tuple:
    """Initialize distributed tracing"""
    try:
        tracer_provider = TracerProvider(resource=resource)
        
        if mode == "disabled":
            logger.info("Tracing disabled")
            return tracer_provider, None
        
        # Add appropriate span processors based on mode
        if mode in ["lgtm_local", "production", "cloud"]:
            # OTLP exporter for unified observability
            otlp_exporter = OTLPSpanExporter(endpoint=endpoint, insecure=True)
            span_processor = BatchSpanProcessor(otlp_exporter)
            tracer_provider.add_span_processor(span_processor)
            logger.info(f"Added OTLP trace exporter: {endpoint}")
        
        if mode == "development" or os.getenv("DEBUG_TRACES"):
            # Console exporter for development
            console_processor = BatchSpanProcessor(ConsoleSpanExporter())
            tracer_provider.add_span_processor(console_processor)
            logger.info("Added console trace exporter for development")
        
        def cleanup():
            # Shutdown span processors
            for processor in tracer_provider._active_span_processor._span_processors:
                processor.shutdown()
        
        return tracer_provider, cleanup
        
    except Exception as e:
        logger.error(f"Failed to initialize tracing: {e}")
        return TracerProvider(resource=resource), None

def _init_metrics(mode: str, endpoint: str, resource: Resource) -> tuple:
    """Initialize metrics collection"""
    try:
        if mode == "disabled":
            meter_provider = MeterProvider(resource=resource)
            logger.info("Metrics disabled")
            return meter_provider, None
        
        metric_readers = []
        
        if mode in ["lgtm_local", "production", "cloud"]:
            # OTLP metric exporter for unified observability
            otlp_metric_exporter = OTLPMetricExporter(endpoint=endpoint, insecure=True)
            otlp_reader = PeriodicExportingMetricReader(
                exporter=otlp_metric_exporter,
                export_interval_millis=10000  # 10 seconds
            )
            metric_readers.append(otlp_reader)
            logger.info(f"Added OTLP metric exporter: {endpoint}")
        
        if mode == "development" or os.getenv("DEBUG_METRICS"):
            # Console exporter for development
            console_reader = PeriodicExportingMetricReader(
                exporter=ConsoleMetricExporter(),
                export_interval_millis=30000  # 30 seconds
            )
            metric_readers.append(console_reader)
            logger.info("Added console metric exporter for development")
        
        meter_provider = MeterProvider(
            resource=resource,
            metric_readers=metric_readers
        )
        
        def cleanup():
            # Shutdown metric readers
            for reader in metric_readers:
                reader.shutdown()
        
        return meter_provider, cleanup
        
    except Exception as e:
        logger.error(f"Failed to initialize metrics: {e}")
        return MeterProvider(resource=resource), None

def get_tracer(name: str) -> trace.Tracer:
    """Get tracer instance"""
    return trace.get_tracer(name)

def get_meter(name: str) -> metrics.Meter:
    """Get meter instance"""
    return metrics.get_meter(name)

def shutdown_otel():
    """Shutdown OpenTelemetry (called on application exit)"""
    global _cleanup_func
    if _cleanup_func:
        _cleanup_func()
        _cleanup_func = None
    else:
        logger.warning("OpenTelemetry cleanup function not found")

# Export common observability utilities
__all__ = [
    "init_otel",
    "shutdown_otel", 
    "get_tracer",
    "get_meter"
]
