#!/usr/bin/env python3
"""
Python ADK Runtime Server
According to spec.md: Python 端是所有 AI 業務邏輯與可插拔能力的唯一執行域
"""
import asyncio
import logging
import os
import signal
import sys
from typing import Optional

import grpc
from detectviz.contracts.v1 import adk_bridge_pb2_grpc as pb_grpc

# Import all service implementations
from detectviz_adk.services.agent_service import AgentService
from detectviz_adk.services.workflow_service import WorkflowService
from detectviz_adk.services.memory_service import MemoryService
from detectviz_adk.services.plugin_service import PluginService
from detectviz_adk.services.health_service import HealthService

from detectviz_adk.config.loader import load_config
from detectviz_adk.observability.otel import init_otel

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='{"timestamp": "%(asctime)s", "level": "%(levelname)s", "logger": "%(name)s", "message": "%(message)s"}',
    stream=sys.stdout
)
logger = logging.getLogger(__name__)

class ADKRuntimeServer:
    """Python ADK Runtime Server - handles all AI execution domain logic"""
    
    def __init__(self, config_path: str = "./configs/config.yaml"):
        self.config = load_config(config_path)
        self.server: Optional[grpc.aio.Server] = None
        self.services = {}
        
        # Initialize observability first
        self._init_observability()
        
        logger.info("🚀 Python ADK Runtime initializing...")
        
    def _init_observability(self):
        """Initialize unified observability according to spec.md"""
        obs_config = self.config.get("observability", {})
        mode = obs_config.get("mode", "lgtm_local")
        endpoint = obs_config.get("otlpEndpoint", "http://localhost:4317")
        service_name = obs_config.get("serviceName", "detectviz-adk-runtime")
        service_version = obs_config.get("serviceVersion", "1.0.0")
        
        logger.info(f"🔍 Initializing observability: mode={mode}, endpoint={endpoint}")
        self.otel_cleanup = init_otel(
            mode=mode, 
            endpoint=endpoint,
            service_name=service_name,
            service_version=service_version
        )
        
    async def start(self, listen_addr: str = "0.0.0.0:9090"):
        """Start the gRPC server with all services"""
        try:
            # Create gRPC server
            self.server = grpc.aio.server()
            
            # Initialize and register all services according to spec.md
            logger.info("📋 Registering gRPC services...")
            
            # AgentService - Core agent execution
            agent_config = self.config.get("agent", {})
            go_plugin_api_url = agent_config.get("goPluginApiUrl", "http://localhost:8080")
            agent_service = AgentService(go_plugin_api_base_url=go_plugin_api_url)
            pb_grpc.add_AgentServiceServicer_to_server(agent_service, self.server)
            self.services['agent'] = agent_service
            logger.info("✅ AgentService registered")
            
            # WorkflowService - Sequential/Parallel/Loop workflows
            workflow_service = WorkflowService()
            pb_grpc.add_WorkflowServiceServicer_to_server(workflow_service, self.server)
            self.services['workflow'] = workflow_service
            logger.info("✅ WorkflowService registered")
            
            # MemoryService - MemoryBank with multiple backends
            memory_config = self.config.get("memory", {})
            memory_service = MemoryService(memory_config)
            pb_grpc.add_MemoryServiceServicer_to_server(memory_service, self.server)
            self.services['memory'] = memory_service
            logger.info("✅ MemoryService registered")
            
            # PluginService - Hot plugin loading and execution
            plugin_config = self.config.get("plugin", {})
            plugin_service = PluginService(plugin_config)
            pb_grpc.add_PluginServiceServicer_to_server(plugin_service, self.server)
            self.services['plugin'] = plugin_service
            logger.info("✅ PluginService registered")
            
            # HealthService - Health checks
            health_service = HealthService()
            pb_grpc.add_HealthServiceServicer_to_server(health_service, self.server)
            self.services['health'] = health_service
            logger.info("✅ HealthService registered")
            
            # Start server
            self.server.add_insecure_port(listen_addr)
            await self.server.start()
            
            logger.info(f"🎯 Python ADK Runtime listening on {listen_addr}")
            logger.info("📡 Registered services: AgentService, WorkflowService, MemoryService, PluginService, HealthService")
            logger.info("🔄 Ready to receive requests from Go Gateway...")
            
            # Wait for termination
            await self.server.wait_for_termination()
            
        except Exception as e:
            logger.error(f"❌ Failed to start server: {e}")
            raise
    
    async def stop(self):
        """Gracefully stop the server"""
        if self.server:
            logger.info("🛑 Shutting down Python ADK Runtime...")
            
            # Stop server with grace period
            await self.server.stop(grace=5.0)
            
            # Cleanup services
            for service_name, service in self.services.items():
                if hasattr(service, 'cleanup'):
                    try:
                        await service.cleanup()
                        logger.info(f"✅ {service_name} service cleaned up")
                    except Exception as e:
                        logger.warning(f"⚠️ {service_name} cleanup failed: {e}")
            
            logger.info("✅ Python ADK Runtime stopped")

async def serve():
    """Main server entry point"""
    # Get configuration
    config_path = os.getenv("CONFIG", "./configs/config.yaml")
    listen_addr = os.getenv("PY_BACKEND_LISTEN", "0.0.0.0:9090")
    
    # Create and start server
    runtime_server = ADKRuntimeServer(config_path)
    
    # Setup signal handlers for graceful shutdown
    def signal_handler(signum, frame):
        logger.info(f"Received signal {signum}, initiating shutdown...")
        asyncio.create_task(runtime_server.stop())
    
    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGTERM, signal_handler)
    
    try:
        await runtime_server.start(listen_addr)
    except KeyboardInterrupt:
        logger.info("KeyboardInterrupt received, shutting down...")
    except Exception as e:
        logger.error(f"Runtime server error: {e}")
        sys.exit(1)

if __name__ == "__main__":
    asyncio.run(serve())
