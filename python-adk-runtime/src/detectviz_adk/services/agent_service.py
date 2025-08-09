# AgentService implementation for Python ADK Runtime
# According to spec.md: Python 端是所有 AI 業務邏輯的唯一執行域
from datetime import datetime, timezone
import asyncio
import logging
import json
import aiohttp
from typing import Dict, Any, Optional

from detectviz.contracts.v1 import adk_bridge_pb2 as pb
from detectviz.contracts.v1 import adk_bridge_pb2_grpc as pb_grpc
from google.rpc import status_pb2 as grpc_status
from opentelemetry import trace

logger = logging.getLogger(__name__)
tracer = trace.get_tracer(__name__)

class AgentService(pb_grpc.AgentServiceServicer):
    """
    AgentService handles core agent execution in Python ADK Runtime.
    According to spec.md: 實際的 AI 推理、工具調用、記憶管理等核心業務邏輯
    """
    
    def __init__(self, go_plugin_api_base_url: str = "http://localhost:8080"):
        self.go_plugin_api_base_url = go_plugin_api_base_url
        self.active_runs = {}  # run_id -> AgentRun
        logger.info(f"AgentService initialized with Go plugin API: {go_plugin_api_base_url}")
    
    async def TriggerAgent(self, request: pb.AgentRequest, context) -> pb.AgentRun:
        """Trigger agent execution with tool calling capability"""
        with tracer.start_as_current_span("agent.trigger") as span:
            try:
                # Generate unique run ID
                now = datetime.now(timezone.utc)
                run_id = f"agent_run_{int(now.timestamp() * 1000)}"
                
                span.set_attributes({
                    "agent.name": request.agent_name,
                    "agent.run_id": run_id,
                    "input.length": len(request.input_text)
                })
                
                logger.info(f"Triggering agent: {request.agent_name}, run_id: {run_id}")
                
                # Create agent run
                agent_run = pb.AgentRun(
                    run_id=run_id,
                    status=pb.PENDING,
                    output_text="",
                    queued_at=pb.Timestamp(seconds=int(now.timestamp()))
                )
                
                # Store active run
                self.active_runs[run_id] = {
                    'request': request,
                    'run': agent_run,
                    'started_at': now
                }
                
                # Start agent execution asynchronously
                asyncio.create_task(self._execute_agent(run_id, request))
                
                # Return immediately with PENDING status
                return agent_run
                
            except Exception as e:
                logger.error(f"Failed to trigger agent {request.agent_name}: {e}")
                span.record_exception(e)
                
                now = datetime.now(timezone.utc)
                return pb.AgentRun(
                    run_id=f"error_{int(now.timestamp())}",
                    status=pb.FAILED,
                    output_text=f"Agent execution failed: {str(e)}",
                    queued_at=pb.Timestamp(seconds=int(now.timestamp())),
                    ended_at=pb.Timestamp(seconds=int(now.timestamp()))
                )

    async def Cancel(self, request: pb.RunId, context) -> grpc_status.Status:
        """Cancel a running agent"""
        try:
            run_id = request.id
            if run_id in self.active_runs:
                run_data = self.active_runs[run_id]
                run_data['run'].status = pb.CANCELLED
                run_data['run'].ended_at.seconds = int(datetime.now(timezone.utc).timestamp())
                logger.info(f"Cancelled agent run: {run_id}")
                
                return grpc_status.Status(
                    code=grpc_status.Code.OK,
                    message=f"Agent run {run_id} cancelled"
                )
            else:
                return grpc_status.Status(
                    code=grpc_status.Code.NOT_FOUND,
                    message=f"Agent run {run_id} not found"
                )
        except Exception as e:
            logger.error(f"Failed to cancel agent run: {e}")
            return grpc_status.Status(
                code=grpc_status.Code.INTERNAL,
                message=f"Cancel failed: {str(e)}"
            )

    async def GetStatus(self, request: pb.RunId, context) -> pb.AgentRun:
        """Get agent run status"""
        run_id = request.id
        
        if run_id not in self.active_runs:
            logger.warning(f"Agent run not found: {run_id}")
            now = datetime.now(timezone.utc)
            return pb.AgentRun(
                run_id=run_id,
                status=pb.FAILED,
                output_text="Agent run not found",
                queued_at=pb.Timestamp(seconds=int(now.timestamp())),
                ended_at=pb.Timestamp(seconds=int(now.timestamp()))
            )
        
        run_data = self.active_runs[run_id]
        return run_data['run']
    
    async def _execute_agent(self, run_id: str, request: pb.AgentRequest):
        """Execute agent with tool calling capability"""
        with tracer.start_as_current_span("agent.execute") as span:
            try:
                run_data = self.active_runs[run_id]
                agent_run = run_data['run']
                
                # Update status to RUNNING
                agent_run.status = pb.RUNNING
                now = datetime.now(timezone.utc)
                agent_run.started_at.seconds = int(now.timestamp())
                
                logger.info(f"Executing agent {request.agent_name} with run_id {run_id}")
                span.set_attributes({
                    "agent.execution.started": True,
                    "agent.input": request.input_text[:100]  # Truncate for logging
                })
                
                # Simulate AI processing with tool calling
                result_parts = []
                result_parts.append(f"[Agent: {request.agent_name}] Processing input: {request.input_text}")
                
                # Example: Call Go plugin service for command execution
                if "execute" in request.input_text.lower() or "run" in request.input_text.lower():
                    tool_result = await self._call_go_plugin_service("exec", {
                        "command": "echo",
                        "args": ["Hello from Go plugin service!"],
                        "timeout": "5s"
                    })
                    
                    if tool_result:
                        result_parts.append(f"Tool execution result: {tool_result.get('stdout', 'No output')}")
                        span.set_attributes({
                            "agent.tool_calls": 1,
                            "agent.tool_success": tool_result.get('success', False)
                        })
                
                # Simulate some processing time
                await asyncio.sleep(0.5)
                
                # Generate final result
                final_output = "\n".join(result_parts)
                final_output += f"\n\n[Processing completed in {datetime.now(timezone.utc) - run_data['started_at']}]"
                
                # Update final status
                agent_run.status = pb.SUCCEEDED
                agent_run.output_text = final_output
                agent_run.ended_at.seconds = int(datetime.now(timezone.utc).timestamp())
                
                span.set_attributes({
                    "agent.execution.completed": True,
                    "agent.output_length": len(final_output)
                })
                
                logger.info(f"Agent {request.agent_name} execution completed: {run_id}")
                
            except Exception as e:
                logger.error(f"Agent execution failed for {run_id}: {e}")
                span.record_exception(e)
                
                if run_id in self.active_runs:
                    agent_run = self.active_runs[run_id]['run']
                    agent_run.status = pb.FAILED
                    agent_run.output_text = f"Agent execution error: {str(e)}"
                    agent_run.ended_at.seconds = int(datetime.now(timezone.utc).timestamp())
    
    async def _call_go_plugin_service(self, plugin_name: str, request_data: Dict[str, Any]) -> Optional[Dict[str, Any]]:
        """Call Go plugin service via HTTP API"""
        try:
            # Prepare request for Go plugin service
            plugin_request = {
                "plugin_name": plugin_name,
                "request": json.dumps(request_data)
            }
            
            async with aiohttp.ClientSession() as session:
                async with session.post(
                    f"{self.go_plugin_api_base_url}/api/v1/plugins/execute",
                    json=plugin_request,
                    timeout=aiohttp.ClientTimeout(total=30)
                ) as response:
                    if response.status == 200:
                        result = await response.json()
                        if result.get('success'):
                            # Parse the response JSON
                            response_data = json.loads(result.get('response', '{}'))
                            logger.info(f"Go plugin call successful: {plugin_name}")
                            return response_data
                        else:
                            logger.error(f"Go plugin call failed: {result.get('error')}")
                            return None
                    else:
                        logger.error(f"HTTP error calling Go plugin: {response.status}")
                        return None
                        
        except Exception as e:
            logger.error(f"Failed to call Go plugin service: {e}")
            return None
    
    def get_stats(self) -> Dict[str, Any]:
        """Get agent service statistics"""
        status_counts = {}
        for run_data in self.active_runs.values():
            status = run_data['run'].status
            status_name = pb.AgentRunStatus.Name(status)
            status_counts[status_name] = status_counts.get(status_name, 0) + 1
        
        return {
            "total_runs": len(self.active_runs),
            "status_counts": status_counts,
            "go_plugin_api": self.go_plugin_api_base_url
        }
