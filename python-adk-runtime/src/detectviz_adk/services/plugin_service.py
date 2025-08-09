# PluginService implementation for Python ADK Runtime
# According to spec.md: Python 端負責插件的實際載入與執行，Go 端僅管理 metadata
import asyncio
import logging
import importlib
import sys
from typing import Dict, Any, Optional, List
from pathlib import Path

from detectviz.contracts.v1 import adk_bridge_pb2 as pb
from detectviz.contracts.v1 import adk_bridge_pb2_grpc as pb_grpc
from google.rpc import status_pb2 as grpc_status

logger = logging.getLogger(__name__)

class PluginService(pb_grpc.PluginServiceServicer):
    """
    PluginService handles actual plugin loading and execution in Python Runtime.
    According to spec.md: 實際載入與執行交由 Python 端，熱載入機制
    """
    
    def __init__(self, config: Dict[str, Any] = None):
        self.config = config or {}
        self.loaded_plugins = {}  # plugin_id -> plugin_instance
        self.plugin_metadata = {}  # plugin_id -> manifest
        self.plugin_status = {}   # plugin_id -> status
        
        # Plugin search paths
        self.plugin_paths = self.config.get('paths', ['./plugins', './src/plugins'])
        
        logger.info(f"PluginService initialized with paths: {self.plugin_paths}")
    
    async def Register(self, request: pb.PluginManifest, context) -> grpc_status.Status:
        """Register and load a plugin"""
        try:
            plugin_id = request.id
            logger.info(f"Registering plugin: {plugin_id} v{request.version}")
            
            # Validate plugin manifest
            if not self._validate_manifest(request):
                return grpc_status.Status(
                    code=grpc_status.Code.INVALID_ARGUMENT,
                    message=f"Invalid plugin manifest for {plugin_id}"
                )
            
            # Store metadata
            self.plugin_metadata[plugin_id] = request
            self.plugin_status[plugin_id] = "registered"
            
            # Load plugin module
            try:
                plugin_instance = await self._load_plugin(plugin_id, request)
                self.loaded_plugins[plugin_id] = plugin_instance
                self.plugin_status[plugin_id] = "loaded"
                
                logger.info(f"✅ Plugin {plugin_id} registered and loaded successfully")
                return grpc_status.Status(
                    code=grpc_status.Code.OK,
                    message=f"Plugin {plugin_id} registered successfully"
                )
                
            except Exception as load_error:
                logger.error(f"Failed to load plugin {plugin_id}: {load_error}")
                self.plugin_status[plugin_id] = "error"
                return grpc_status.Status(
                    code=grpc_status.Code.INTERNAL,
                    message=f"Plugin loading failed: {str(load_error)}"
                )
                
        except Exception as e:
            logger.error(f"Plugin registration failed: {e}")
            return grpc_status.Status(
                code=grpc_status.Code.INTERNAL,
                message=f"Registration failed: {str(e)}"
            )
    
    async def Unregister(self, request: pb.PluginId, context) -> grpc_status.Status:
        """Unregister and unload a plugin"""
        try:
            plugin_id = request.id
            logger.info(f"Unregistering plugin: {plugin_id}")
            
            if plugin_id not in self.plugin_metadata:
                return grpc_status.Status(
                    code=grpc_status.Code.NOT_FOUND,
                    message=f"Plugin {plugin_id} not found"
                )
            
            # Cleanup plugin instance
            if plugin_id in self.loaded_plugins:
                plugin_instance = self.loaded_plugins[plugin_id]
                
                # Call cleanup if available
                if hasattr(plugin_instance, 'cleanup'):
                    try:
                        await plugin_instance.cleanup()
                    except Exception as cleanup_error:
                        logger.warning(f"Plugin cleanup failed for {plugin_id}: {cleanup_error}")
                
                del self.loaded_plugins[plugin_id]
            
            # Remove metadata
            del self.plugin_metadata[plugin_id]
            del self.plugin_status[plugin_id]
            
            logger.info(f"✅ Plugin {plugin_id} unregistered successfully")
            return grpc_status.Status(
                code=grpc_status.Code.OK,
                message=f"Plugin {plugin_id} unregistered successfully"
            )
            
        except Exception as e:
            logger.error(f"Plugin unregistration failed: {e}")
            return grpc_status.Status(
                code=grpc_status.Code.INTERNAL,
                message=f"Unregistration failed: {str(e)}"
            )
    
    async def List(self, request: pb.PluginListRequest, context) -> pb.PluginList:
        """List all registered plugins"""
        try:
            plugins = []
            
            for plugin_id, manifest in self.plugin_metadata.items():
                # Add runtime status to manifest
                runtime_manifest = pb.PluginManifest()
                runtime_manifest.CopyFrom(manifest)
                
                # Add status as metadata (if possible)
                status = self.plugin_status.get(plugin_id, "unknown")
                logger.debug(f"Plugin {plugin_id}: status={status}")
                
                plugins.append(runtime_manifest)
            
            logger.info(f"Listed {len(plugins)} registered plugins")
            return pb.PluginList(plugins=plugins)
            
        except Exception as e:
            logger.error(f"Failed to list plugins: {e}")
            return pb.PluginList(plugins=[])
    
    async def _load_plugin(self, plugin_id: str, manifest: pb.PluginManifest) -> Any:
        """Load plugin module and create instance"""
        try:
            entry_point = manifest.entry_point  # e.g., "main.py" or "module:class"
            
            # Find plugin directory
            plugin_dir = self._find_plugin_directory(plugin_id)
            if not plugin_dir:
                raise ValueError(f"Plugin directory not found for {plugin_id}")
            
            # Add plugin directory to Python path
            plugin_path = str(plugin_dir)
            if plugin_path not in sys.path:
                sys.path.insert(0, plugin_path)
            
            # Load module
            if ":" in entry_point:
                # Format: "module:class"
                module_name, class_name = entry_point.split(":")
                module = importlib.import_module(module_name)
                plugin_class = getattr(module, class_name)
                plugin_instance = plugin_class()
            else:
                # Format: "main.py" - expect register_plugin() function
                module_name = entry_point.replace(".py", "")
                module = importlib.import_module(module_name)
                
                if hasattr(module, 'register_plugin'):
                    plugin_instance = module.register_plugin()
                else:
                    raise ValueError(f"Plugin {plugin_id} missing register_plugin() function")
            
            # Initialize plugin if it has init method
            if hasattr(plugin_instance, 'init'):
                await plugin_instance.init()
            
            logger.info(f"Plugin {plugin_id} loaded from {plugin_path}")
            return plugin_instance
            
        except Exception as e:
            logger.error(f"Failed to load plugin {plugin_id}: {e}")
            raise
    
    def _find_plugin_directory(self, plugin_id: str) -> Optional[Path]:
        """Find plugin directory in search paths"""
        for search_path in self.plugin_paths:
            plugin_dir = Path(search_path) / plugin_id
            if plugin_dir.exists() and plugin_dir.is_dir():
                return plugin_dir
        
        # Also try with namespace prefix removed
        simple_id = plugin_id.split('.')[-1]
        for search_path in self.plugin_paths:
            plugin_dir = Path(search_path) / simple_id
            if plugin_dir.exists() and plugin_dir.is_dir():
                return plugin_dir
        
        return None
    
    def _validate_manifest(self, manifest: pb.PluginManifest) -> bool:
        """Validate plugin manifest"""
        try:
            # Required fields
            if not manifest.id:
                logger.error("Plugin ID is required")
                return False
            
            if not manifest.version:
                logger.error("Plugin version is required")
                return False
            
            if not manifest.entry_point:
                logger.error("Plugin entry point is required")
                return False
            
            # Validate type
            valid_types = ["agent", "tool", "capability", "memory"]
            if manifest.type not in valid_types:
                logger.error(f"Invalid plugin type: {manifest.type}")
                return False
            
            # Validate host compatibility
            if not manifest.host_compatibility:
                logger.error("Host compatibility is required")
                return False
            
            return True
            
        except Exception as e:
            logger.error(f"Manifest validation error: {e}")
            return False
    
    def get_plugin(self, plugin_id: str) -> Optional[Any]:
        """Get loaded plugin instance"""
        return self.loaded_plugins.get(plugin_id)
    
    def get_plugin_status(self, plugin_id: str) -> Optional[str]:
        """Get plugin status"""
        return self.plugin_status.get(plugin_id)
    
    def get_stats(self) -> Dict[str, Any]:
        """Get plugin service statistics"""
        status_counts = {}
        for status in self.plugin_status.values():
            status_counts[status] = status_counts.get(status, 0) + 1
        
        return {
            "total_plugins": len(self.plugin_metadata),
            "loaded_plugins": len(self.loaded_plugins),
            "status_counts": status_counts,
            "plugin_paths": self.plugin_paths
        }