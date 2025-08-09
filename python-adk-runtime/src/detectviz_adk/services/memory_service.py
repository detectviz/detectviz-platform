# MemoryService implementation for Python ADK Runtime
# According to spec.md: MemoryBank API 與一致性約束，支持多種後端
import asyncio
import logging
from datetime import datetime, timezone
from typing import Dict, List, Optional, Any

from detectviz.contracts.v1 import adk_bridge_pb2 as pb
from detectviz.contracts.v1 import adk_bridge_pb2_grpc as pb_grpc
from google.rpc import status_pb2 as grpc_status

logger = logging.getLogger(__name__)

class MemoryService(pb_grpc.MemoryServiceServicer):
    """
    MemoryService provides unified memory management across different backends.
    According to spec.md: inmem/redis/weaviate/chroma/vertex backends with consistent API
    """
    
    def __init__(self, config: Dict[str, Any] = None):
        self.config = config or {}
        self.backend_type = self.config.get('backend', 'inmem')
        
        # In-memory storage for development/testing
        self.memory_store = {}  # key -> MemoryItem
        self.scoped_keys = {}   # scope -> set of keys
        
        logger.info(f"MemoryService initialized with backend: {self.backend_type}")
    
    async def Write(self, request_iterator, context) -> grpc_status.Status:
        """Write memory items (streaming)"""
        try:
            write_count = 0
            async for memory_item in request_iterator:
                await self._write_item(memory_item)
                write_count += 1
            
            logger.info(f"Successfully wrote {write_count} memory items")
            return grpc_status.Status(
                code=grpc_status.Code.OK,
                message=f"Successfully wrote {write_count} items"
            )
            
        except Exception as e:
            logger.error(f"Memory write failed: {e}")
            return grpc_status.Status(
                code=grpc_status.Code.INTERNAL,
                message=f"Write failed: {str(e)}"
            )
    
    async def Read(self, request: pb.MemoryReadRequest, context):
        """Read memory items by keys (streaming response)"""
        try:
            for key in request.keys:
                if key in self.memory_store:
                    item = self.memory_store[key]
                    logger.debug(f"Reading memory item: {key}")
                    yield item
                else:
                    logger.warning(f"Memory item not found: {key}")
                    # Could yield empty item or skip
                    
        except Exception as e:
            logger.error(f"Memory read failed: {e}")
            # In streaming, we can't return error status, so we log and return
    
    async def Search(self, request: pb.MemorySearchRequest, context):
        """Search memory items by query (streaming response)"""
        try:
            tenant_id = request.tenant.tenant_id if request.tenant else "default"
            query = request.query
            top_k = request.top_k if request.top_k > 0 else 10
            scope = dict(request.scope) if request.scope else {}
            
            logger.info(f"Memory search: tenant={tenant_id}, query='{query}', top_k={top_k}")
            
            # Simple search implementation for development
            # In production, this would use vector similarity search
            matching_items = []
            
            for key, item in self.memory_store.items():
                # Check scope matching
                if scope:
                    item_scope = dict(item.scope) if item.scope else {}
                    if not all(item_scope.get(k) == v for k, v in scope.items()):
                        continue
                
                # Simple text matching for now
                if query.lower() in item.data.lower():
                    # Calculate mock similarity score
                    item.score = 0.8  # Mock score
                    matching_items.append((item.score, item))
            
            # Sort by score and limit results
            matching_items.sort(reverse=True, key=lambda x: x[0])
            
            for score, item in matching_items[:top_k]:
                logger.debug(f"Search result: key={item.key}, score={score}")
                yield item
                
        except Exception as e:
            logger.error(f"Memory search failed: {e}")
    
    async def Flush(self, request: pb.MemoryFlushRequest, context) -> grpc_status.Status:
        """Flush memory items by scope"""
        try:
            scope = dict(request.scope) if request.scope else {}
            
            if not scope:
                # Flush all memory
                cleared_count = len(self.memory_store)
                self.memory_store.clear()
                self.scoped_keys.clear()
                logger.info(f"Flushed all memory ({cleared_count} items)")
            else:
                # Flush by scope
                cleared_count = 0
                keys_to_remove = []
                
                for key, item in self.memory_store.items():
                    item_scope = dict(item.scope) if item.scope else {}
                    if all(item_scope.get(k) == v for k, v in scope.items()):
                        keys_to_remove.append(key)
                        cleared_count += 1
                
                for key in keys_to_remove:
                    del self.memory_store[key]
                
                # Update scoped_keys
                scope_key = "_".join(f"{k}={v}" for k, v in scope.items())
                if scope_key in self.scoped_keys:
                    del self.scoped_keys[scope_key]
                
                logger.info(f"Flushed memory by scope {scope} ({cleared_count} items)")
            
            return grpc_status.Status(
                code=grpc_status.Code.OK,
                message=f"Successfully flushed {cleared_count} items"
            )
            
        except Exception as e:
            logger.error(f"Memory flush failed: {e}")
            return grpc_status.Status(
                code=grpc_status.Code.INTERNAL,
                message=f"Flush failed: {str(e)}"
            )
    
    async def _write_item(self, item: pb.MemoryItem):
        """Write a single memory item to backend"""
        try:
            # Set timestamps if not provided
            if not item.created_at.seconds:
                now = datetime.now(timezone.utc)
                item.created_at.seconds = int(now.timestamp())
            
            # Store in memory (for development)
            self.memory_store[item.key] = item
            
            # Track by scope
            if item.scope:
                scope_key = "_".join(f"{k}={v}" for k, v in item.scope.items())
                if scope_key not in self.scoped_keys:
                    self.scoped_keys[scope_key] = set()
                self.scoped_keys[scope_key].add(item.key)
            
            logger.debug(f"Stored memory item: key={item.key}, kind={item.kind}")
            
            # In production, this would write to configured backend:
            # - Redis for fast access
            # - Weaviate/Chroma for vector search
            # - Vertex AI for managed service
            
        except Exception as e:
            logger.error(f"Failed to write memory item {item.key}: {e}")
            raise
    
    def get_stats(self) -> Dict[str, Any]:
        """Get memory service statistics"""
        return {
            "backend_type": self.backend_type,
            "total_items": len(self.memory_store),
            "scoped_namespaces": len(self.scoped_keys),
            "memory_kinds": {
                kind.name: sum(1 for item in self.memory_store.values() if item.kind == kind)
                for kind in [pb.TEXT, pb.JSON, pb.BINARY, pb.VECTOR]
            }
        }